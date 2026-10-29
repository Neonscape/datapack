package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const projectExtension = ".dtproj"

type ProjectConfig struct {
	Version     int          `json:"version"`
	TableRoots  []string     `json:"tableRoots"`
	SchemaRoots []string     `json:"schemaRoots"`
	Exclude     []string     `json:"exclude,omitempty"`
	Export      ExportConfig `json:"export"`
}

type ExportConfig struct {
	Directory string          `json:"directory"`
	Targets   []string        `json:"targets"`
	Tables    map[string]bool `json:"tables"`
}

type Project struct {
	Path   string        `json:"path"`
	Root   string        `json:"root"`
	Config ProjectConfig `json:"config"`
}

func LoadProject(path string) (Project, error) {
	if filepath.Ext(path) != projectExtension {
		return Project{}, fmt.Errorf("project file must use %s", projectExtension)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Project{}, fmt.Errorf("resolve project path: %w", err)
	}
	file, err := os.Open(absPath)
	if err != nil {
		return Project{}, fmt.Errorf("open project file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var config ProjectConfig
	if err := decoder.Decode(&config); err != nil {
		return Project{}, fmt.Errorf("decode project file: %w", err)
	}
	project := Project{Path: absPath, Root: filepath.Dir(absPath), Config: config}
	if _, err := exportDirectory(project); err != nil {
		return Project{}, err
	}
	return project, nil
}

// SaveProject atomically persists the project configuration.
func SaveProject(project Project) error {
	if _, err := exportDirectory(project); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(project.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode project file: %w", err)
	}
	if err := writeFileAtomic(project.Path, encoded); err != nil {
		return fmt.Errorf("save project file: %w", err)
	}
	return nil
}

func validateProjectConfig(config ProjectConfig) error {
	if config.Version != 1 {
		return fmt.Errorf("unsupported project format version %d", config.Version)
	}
	if len(config.TableRoots) == 0 {
		return fmt.Errorf("project must declare at least one table root")
	}
	if len(config.SchemaRoots) == 0 {
		return fmt.Errorf("project must declare at least one schema root")
	}
	if err := validateRoots("tableRoots", config.TableRoots); err != nil {
		return err
	}
	if err := validateRoots("schemaRoots", config.SchemaRoots); err != nil {
		return err
	}
	if strings.TrimSpace(config.Export.Directory) == "" {
		return fmt.Errorf("project must declare an export directory")
	}
	cleanExport := filepath.Clean(config.Export.Directory)
	if filepath.IsAbs(config.Export.Directory) || cleanExport == "." || cleanExport == ".." || strings.HasPrefix(cleanExport, ".."+string(filepath.Separator)) {
		return fmt.Errorf("export directory %q escapes the project directory", config.Export.Directory)
	}
	// Output replaces an entire directory. Keep it disjoint from source roots,
	// including excluded files and roots that have not been created yet.
	protected := append([]string{"cue.mod", ".git"}, config.TableRoots...)
	protected = append(protected, config.SchemaRoots...)
	for _, path := range protected {
		if pathsOverlap(cleanExport, filepath.Clean(path)) {
			return fmt.Errorf("export directory %q overlaps protected project path %q", config.Export.Directory, path)
		}
	}
	if len(config.Export.Targets) == 0 {
		return fmt.Errorf("project must declare at least one export target")
	}
	return nil
}

func validateRoots(name string, roots []string) error {
	seen := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		if root == "" || filepath.IsAbs(root) {
			return fmt.Errorf("%s must contain non-empty relative paths", name)
		}
		clean := filepath.Clean(root)
		if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("%s path %q escapes the project directory", name, root)
		}
		if _, exists := seen[clean]; exists {
			return fmt.Errorf("%s contains duplicate path %q", name, root)
		}
		seen[clean] = struct{}{}
	}
	return nil
}
