package workspace

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// CreateProject creates an empty project at path and immediately opens it.
func (s *Service) CreateProject(path string) (View, error) {
	project, err := newProjectAt(path)
	if err != nil {
		return View{}, err
	}
	if err := createProjectFiles(project); err != nil {
		return View{}, err
	}
	return s.OpenProject(project.Path)
}

// SaveProjectAs copies the source portion of the current project to path and
// switches the editor to the copied project.
func (s *Service) SaveProjectAs(path string) (View, error) {
	if s.current == nil {
		return View{}, s.userError("workspace.noProject")
	}
	target, err := newProjectAt(path)
	if err != nil {
		return View{}, err
	}
	canonicalTarget, err := canonicalPath(target.Root)
	if err != nil {
		return View{}, err
	}
	canonicalSource, err := canonicalPath(s.current.Project.Root)
	if err != nil {
		return View{}, err
	}
	if isWithin(canonicalTarget, canonicalSource) {
		return View{}, fmt.Errorf("save-as destination must not be inside the current project")
	}
	if err := validateProjectSlots(target.Root, target.Path); err != nil {
		return View{}, err
	}
	if err := os.MkdirAll(target.Root, 0o755); err != nil {
		return View{}, fmt.Errorf("create save-as directory: %w", err)
	}
	stage, err := os.MkdirTemp(filepath.Dir(target.Root), ".datapack-save-as-*")
	if err != nil {
		return View{}, fmt.Errorf("create save-as staging directory: %w", err)
	}
	defer os.RemoveAll(stage)
	if err := copyProjectSource(s.current.Project, stage); err != nil {
		return View{}, err
	}
	target.Config = s.current.Project.Config
	if err := SaveProject(targetInRoot(target, stage)); err != nil {
		return View{}, err
	}
	entries, err := os.ReadDir(stage)
	if err != nil {
		return View{}, fmt.Errorf("read save-as staging directory: %w", err)
	}
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(target.Root, entry.Name())); err == nil {
			return View{}, fmt.Errorf("save-as target already contains %q", entry.Name())
		} else if !os.IsNotExist(err) {
			return View{}, fmt.Errorf("inspect save-as target %q: %w", entry.Name(), err)
		}
	}
	moved := []string{}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(stage, entry.Name()), filepath.Join(target.Root, entry.Name())); err != nil {
			for index := len(moved) - 1; index >= 0; index-- {
				name := moved[index]
				_ = os.Rename(filepath.Join(target.Root, name), filepath.Join(stage, name))
			}
			return View{}, fmt.Errorf("finish save-as copy: %w", err)
		}
		moved = append(moved, entry.Name())
	}
	return s.OpenProject(target.Path)
}

func newProjectAt(path string) (Project, error) {
	if filepath.Ext(path) != projectExtension {
		return Project{}, fmt.Errorf("project file must use %s", projectExtension)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Project{}, fmt.Errorf("resolve project path: %w", err)
	}
	if _, err := os.Stat(absPath); err == nil {
		return Project{}, fmt.Errorf("project file %q already exists", absPath)
	} else if !os.IsNotExist(err) {
		return Project{}, fmt.Errorf("inspect project path: %w", err)
	}
	project := Project{Path: absPath, Root: filepath.Dir(absPath), Config: ProjectConfig{
		Version:     1,
		TableRoots:  []string{"tables"},
		SchemaRoots: []string{"schema"},
		Exclude:     []string{"exports", ".git"},
		Export:      ExportConfig{Directory: "exports", Targets: []string{"csv"}, Tables: map[string]bool{}},
	}}
	if err := validateProjectSlots(project.Root, project.Path); err != nil {
		return Project{}, err
	}
	return project, nil
}

func createProjectFiles(project Project) error {
	if err := os.MkdirAll(project.Root, 0o755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}
	for _, directory := range append(append([]string{}, project.Config.TableRoots...), project.Config.SchemaRoots...) {
		if err := os.MkdirAll(filepath.Join(project.Root, directory), 0o755); err != nil {
			return fmt.Errorf("create project source directory: %w", err)
		}
	}
	commonDirectory := filepath.Join(project.Root, project.Config.SchemaRoots[0], "common")
	if err := os.MkdirAll(commonDirectory, 0o755); err != nil {
		return fmt.Errorf("create common Schema directory: %w", err)
	}
	commonSchema := `package common

#ID: string & =~"^[A-Za-z_][A-Za-z0-9_]*$"
#NonEmptyString: string & != ""
#NonNegative: >=0
#Probability: >=0 & <=1

_schema: {
	version: 1
	definitions: {
		ID: {kind: "type", type: "string"}
		NonEmptyString: {kind: "constraint", appliesTo: ["string"]}
		NonNegative: {kind: "constraint", appliesTo: ["integer", "number"]}
		Probability: {kind: "constraint", appliesTo: ["integer", "number"]}
	}
}
`
	if err := writeFileAtomic(filepath.Join(commonDirectory, "common.cue"), []byte(commonSchema)); err != nil {
		return fmt.Errorf("write common Schema: %w", err)
	}
	moduleDirectory := filepath.Join(project.Root, "cue.mod")
	if err := os.MkdirAll(moduleDirectory, 0o755); err != nil {
		return fmt.Errorf("create CUE module directory: %w", err)
	}
	module := fmt.Sprintf("module: %q\nlanguage: {\n\tversion: \"v0.9.0\"\n}\n", "example.com/datapack/"+moduleName(project.Path))
	if err := writeFileAtomic(filepath.Join(moduleDirectory, "module.cue"), []byte(module)); err != nil {
		return fmt.Errorf("write CUE module: %w", err)
	}
	if err := SaveProject(project); err != nil {
		return err
	}
	return nil
}

func validateProjectSlots(root, projectPath string) error {
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("project file %q already exists", projectPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect project path: %w", err)
	}
	for _, name := range []string{"tables", "schema", "cue.mod"} {
		path := filepath.Join(root, name)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect project directory %q: %w", name, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("project path %q must be a directory", name)
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("read project directory %q: %w", name, err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("project directory %q is not empty", name)
		}
	}
	return nil
}

func copyProjectSource(project Project, destination string) error {
	skipped := map[string]bool{".git": true, filepath.Clean(project.Config.Export.Directory): true}
	for _, path := range project.Config.Exclude {
		skipped[filepath.Clean(path)] = true
	}
	return filepath.WalkDir(project.Root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(project.Root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if relative == filepath.Base(project.Path) {
			return nil
		}
		for skippedPath := range skipped {
			if relative == skippedPath || strings.HasPrefix(relative, skippedPath+string(filepath.Separator)) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("save-as cannot copy symbolic link %q", relative)
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()
		info, err := entry.Info()
		if err != nil {
			return err
		}
		targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(targetFile, source)
		closeErr := targetFile.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func targetInRoot(project Project, root string) Project {
	project.Root = root
	project.Path = filepath.Join(root, filepath.Base(project.Path))
	return project
}

func isWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func canonicalPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("resolve path symlinks: %w", err)
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil {
		return "", fmt.Errorf("resolve path parent symlinks: %w", err)
	}
	return filepath.Join(parent, filepath.Base(absPath)), nil
}

func moduleName(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	var result strings.Builder
	for _, character := range strings.ToLower(name) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) || character == '-' {
			result.WriteRune(character)
			continue
		}
		result.WriteByte('-')
	}
	clean := strings.Trim(result.String(), "-")
	if clean == "" {
		return "project"
	}
	return clean
}
