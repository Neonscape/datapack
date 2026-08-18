package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateProjectBuildsInitialLayout(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "demo", "demo.dtproj")
	service := NewService(t.TempDir())
	view, err := service.CreateProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if view.Project.Path != path {
		t.Fatalf("project path = %q", view.Project.Path)
	}
	for _, directory := range []string{"tables", "schema", "cue.mod"} {
		if info, err := os.Stat(filepath.Join(root, "demo", directory)); err != nil || !info.IsDir() {
			t.Fatalf("created directory %q: %v", directory, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "demo", "cue.mod", "module.cue")); err != nil {
		t.Fatal(err)
	}
}

func TestSaveProjectAsCopiesSourcesAndSwitchesWorkspace(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "source", "source.dtproj")
	writeProjectTestFile(t, sourcePath, `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"], "tables": {}}}`)
	writeProjectTestFile(t, filepath.Join(root, "source", "tables", "item", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(root, "source", "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(sourcePath); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(root, "copy", "copy.dtproj")
	view, err := service.SaveProjectAs(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if view.Project.Path != targetPath {
		t.Fatalf("saved-as project path = %q", view.Project.Path)
	}
	if _, err := os.Stat(filepath.Join(root, "copy", "tables", "item", "item.cue")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("source project missing: %v", err)
	}
}

func TestSaveProjectAsRejectsDestinationInsideSource(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "source.dtproj")
	writeProjectTestFile(t, sourcePath, `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"], "tables": {}}}`)
	if err := os.MkdirAll(filepath.Join(root, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(sourcePath); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveProjectAs(filepath.Join(root, "copy", "copy.dtproj")); err == nil {
		t.Fatal("SaveProjectAs() accepted a nested destination")
	}
}

func writeProjectTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
