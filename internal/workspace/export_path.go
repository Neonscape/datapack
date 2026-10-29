package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// exportDirectory validates without writing anything. Recheck at export time:
// filesystem aliases may have changed since the project/settings were loaded.
func exportDirectory(project Project) (string, error) {
	if err := validateProjectConfig(project.Config); err != nil {
		return "", err
	}
	root, err := filepath.Abs(project.Root)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}
	destination := filepath.Join(root, project.Config.Export.Directory)
	resolvedRoot, err := resolveExportPath(root)
	if err != nil {
		return "", err
	}
	resolvedDestination, err := resolveExportPath(destination)
	if err != nil {
		return "", err
	}
	if !isWithin(resolvedDestination, resolvedRoot) || isWithin(resolvedRoot, resolvedDestination) {
		return "", fmt.Errorf("export directory %q must be strictly inside the project directory", project.Config.Export.Directory)
	}

	checkProtected := func(path string) error {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		resolved, err := resolveExportPath(absolute)
		if err != nil {
			return err
		}
		if pathsOverlap(destination, absolute) || pathsOverlap(resolvedDestination, resolved) {
			return fmt.Errorf("export directory %q overlaps protected project path %q", project.Config.Export.Directory, path)
		}
		return nil
	}
	for _, path := range []string{project.Path, filepath.Join(root, "cue.mod"), filepath.Join(root, ".git")} {
		if err := checkProtected(path); err != nil {
			return "", err
		}
	}
	sourceRoots := append(append([]string{}, project.Config.TableRoots...), project.Config.SchemaRoots...)
	for _, sourceRoot := range sourceRoots {
		path := filepath.Join(root, sourceRoot)
		if err := checkProtected(path); err != nil {
			return "", err
		}
		resolved, err := resolveExportPath(path)
		if err != nil {
			return "", err
		}
		// A CUE source file can itself be a symlink into an otherwise disjoint
		// output directory. Protect its target too, even when excluded from
		// discovery or deselected for export. WalkDir does not follow directory
		// symlinks, matching project-tree discovery below a source root.
		err = filepath.WalkDir(resolved, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if path == resolved && os.IsNotExist(walkErr) {
					return nil // Saving a new project need not create its roots yet.
				}
				return walkErr
			}
			if !entry.IsDir() && filepath.Ext(path) == ".cue" {
				return checkProtected(path)
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("check export source paths: %w", err)
		}
	}
	if info, err := os.Stat(resolvedDestination); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("export directory %q is not a directory", project.Config.Export.Directory)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect export directory: %w", err)
	}
	// Use the checked physical path rather than traversing an alias again.
	return resolvedDestination, nil
}

func pathsOverlap(first, second string) bool {
	return isWithin(first, second) || isWithin(second, first)
}

// resolveExportPath resolves existing ancestors even when several final path
// components do not exist. Lstat distinguishes missing paths from dangling
// symlinks; those and other resolution errors must fail closed.
func resolveExportPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve export safety path %q: %w", path, err)
	}
	var missing []string
	current := absolute
	for {
		_, err := os.Lstat(current)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", fmt.Errorf("resolve export safety path %q: %w", path, err)
			}
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) || filepath.Dir(current) == current {
			return "", fmt.Errorf("inspect export safety path %q: %w", path, err)
		}
		missing = append(missing, filepath.Base(current))
		current = filepath.Dir(current)
	}
}
