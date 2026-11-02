package workspace

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
	"datapack/internal/perftrace"
)

// inputSnapshot covers directory membership as well as file bytes. Reuse is
// deliberately narrower than consistency checking: imported packages are
// tracked but never reused until a complete immutable import policy is added.
type inputSnapshot struct {
	fingerprint string
	cacheable   bool
	files       map[string][]byte
}
type inputInventory struct {
	dirs    map[string]bool
	files   map[string]bool
	modules map[string]bool
}

func moduleRoot(directory string) string {
	for d := directory; ; d = filepath.Dir(d) {
		if info, err := os.Stat(filepath.Join(d, "cue.mod")); err == nil && info.IsDir() {
			return d
		}
		if filepath.Dir(d) == d {
			return ""
		}
	}
}
func (i *inputInventory) addPackage(directory, root string) {
	for d := directory; ; d = filepath.Dir(d) {
		i.dirs[d] = true
		if d == root || filepath.Dir(d) == d {
			break
		}
	}
	if root != "" {
		i.modules[filepath.Join(root, "cue.mod")] = true
	}
}
func (i *inputInventory) addTree(root string) error {
	resolved, err := filepath.EvalSymlinks(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return filepath.WalkDir(resolved, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			i.dirs[path] = true
		}
		return nil
	})
}
func (i *inputInventory) read() (map[string][]byte, error) {
	paths := map[string]bool{}
	for p := range i.files {
		paths[p] = true
	}
	for dir := range i.dirs {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".cue" {
				paths[filepath.Join(dir, entry.Name())] = true
			}
		}
	}
	for root := range i.modules {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if !entry.IsDir() {
				paths[path] = true
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	result := map[string][]byte{}
	for path := range paths {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		result[path] = data
		// Retargeting a file symlink is itself an input change, even if bytes match.
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return nil, err
			}
			result[path+"\x00link"] = []byte(target)
		}
	}
	return result, nil
}
func fingerprintFiles(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for path := range files {
		names = append(names, path)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, path := range names {
		fmt.Fprintf(h, "%d:%s:%d:", len(path), path, len(files[path]))
		h.Write(files[path])
	}
	return hex.EncodeToString(h.Sum(nil))
}
func captureProjectInputs(project Project, seeds []string) (inputSnapshot, error) {
	defer perftrace.Start("inputs.capture")()
	inv := inputInventory{dirs: map[string]bool{}, files: map[string]bool{project.Path: true}, modules: map[string]bool{}}
	root := moduleRoot(project.Root)
	inv.addPackage(project.Root, root)
	roots := append(append([]string{}, project.Config.TableRoots...), project.Config.SchemaRoots...)
	for _, relative := range roots {
		path := filepath.Join(project.Root, relative)
		if err := inv.addTree(path); err != nil {
			return inputSnapshot{}, err
		}
		inv.addPackage(path, root)
	}
	files, err := inv.read()
	if err != nil {
		return inputSnapshot{}, err
	}
	imported := false
	for path, data := range files {
		if filepath.Ext(path) != ".cue" {
			continue
		}
		header, err := parser.ParseFile(path, data, parser.ImportsOnly)
		if err != nil {
			return inputSnapshot{}, err
		}
		for _, d := range header.Decls {
			if _, ok := d.(*ast.ImportDecl); ok {
				imported = true
			}
		}
		if err := rejectUnsupportedExternalInput(path, data); err != nil {
			return inputSnapshot{}, err
		}
	}
	if !imported {
		again, err := inv.read()
		if err != nil {
			return inputSnapshot{}, err
		}
		if fingerprintFiles(files) != fingerprintFiles(again) {
			return inputSnapshot{}, fmt.Errorf("project inputs changed while taking snapshot")
		}
		return inputSnapshot{fingerprintFiles(again), true, again}, nil
	}
	if len(seeds) == 0 {
		return inputSnapshot{}, fmt.Errorf("cannot resolve imported inputs without a discovered package")
	}
	// Resolve using captured source overlays. Repeat only to close newly found
	// transitive dependencies; an unstable filesystem never becomes a hit.
	for attempt := 0; attempt < 3; attempt++ {
		overlay := map[string]load.Source{}
		for path, data := range files {
			if filepath.Ext(path) == ".cue" {
				overlay[path] = load.FromBytes(data)
			}
		}
		visited := map[*build.Instance]bool{}
		packages := map[string]bool{}
		var visit func(*build.Instance) error
		visit = func(instance *build.Instance) error {
			if visited[instance] {
				return nil
			}
			visited[instance] = true
			if instance.Err != nil {
				return instance.Err
			}
			if instance.Dir != "" {
				inv.addPackage(instance.Dir, instance.Root)
			}
			for _, f := range instance.BuildFiles {
				if f.Filename != "" {
					inv.files[f.Filename] = true
					if raw, ok := f.Source.([]byte); ok {
						if err := rejectUnsupportedExternalInput(f.Filename, raw); err != nil {
							return err
						}
					}
					data, ok := files[f.Filename]
					if !ok {
						var err error
						data, err = os.ReadFile(f.Filename)
						if err != nil {
							return err
						}
					}
					if err := rejectUnsupportedExternalInput(f.Filename, data); err != nil {
						return err
					}
				}
			}
			for _, dependency := range instance.Imports {
				if err := visit(dependency); err != nil {
					return err
				}
			}
			// Inspect the complete transitive source closure before evaluation.
			// Bottoms in imported packages cannot hide behind a root projection.
			if err := cuecontext.New().BuildInstance(instance).Validate(cue.Definitions(true), cue.Optional(true)); err != nil {
				return inputSnapshotDependencyError(instance.ImportPath, err)
			}
			return nil
		}
		for _, seed := range seeds {
			data, ok := files[seed]
			if !ok {
				data, err = os.ReadFile(seed)
				if err != nil {
					return inputSnapshot{}, err
				}
			}
			header, err := parser.ParseFile(seed, data, parser.ImportsOnly)
			if err != nil {
				return inputSnapshot{}, err
			}
			name := header.PackageName()
			key := filepath.Dir(seed) + ":" + name
			if packages[key] {
				continue
			}
			packages[key] = true
			instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(seed), Package: name, Overlay: overlay})
			if len(instances) != 1 {
				return inputSnapshot{}, fmt.Errorf("cannot resolve exactly one package for %s", seed)
			}
			if err := visit(instances[0]); err != nil {
				return inputSnapshot{}, err
			}
		}
		next, err := inv.read()
		if err != nil {
			return inputSnapshot{}, err
		}
		for path, data := range next {
			if err := rejectUnsupportedExternalInput(path, data); err != nil {
				return inputSnapshot{}, err
			}
		}
		if fingerprintFiles(files) == fingerprintFiles(next) {
			return inputSnapshot{fingerprintFiles(next), false, next}, nil
		}
		files = next
	}
	return inputSnapshot{}, fmt.Errorf("CUE input dependency closure did not stabilize")
}
func (s *Service) captureInputs() (inputSnapshot, error) {
	var seeds []string
	collectSourcePaths(s.current.Tree.Root, &seeds)
	return captureProjectInputs(s.current.Project, seeds)
}
func (s *Service) sourceFingerprint() (string, error) {
	v, err := s.captureInputs()
	return v.fingerprint, err
}

type inputConsistencyHooks struct {
	afterCompile       func()
	afterCacheHit      func()
	beforeExportCommit func()
}

func inputSnapshotDependencyError(path string, err error) error {
	return fmt.Errorf("invalid CUE input package %s: %w", path, err)
}

// Match the conservative initial-input policy for every discovered CUE file.
func rejectUnsupportedExternalInput(path string, data []byte) error {
	if filepath.Ext(path) == ".cue" && (bytes.Contains(data, []byte("@embed")) || bytes.Contains(data, []byte("@extern"))) {
		return fmt.Errorf("cannot establish complete external input dependencies for %s", path)
	}
	return nil
}
