package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/projecttree"
)

func TestOpenProjectLoadsSettingsAndSourceTree(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{
  "version": 1,
  "tableRoots": ["tables"],
  "schemaRoots": ["schema"],
  "exclude": ["tables/generated"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "items", "item.cue"), `package item
	#PrimaryKey: string
#Row: {name: string}
rows: {sword: "sword"}
_tool: {table: {namespace: "rpg", name: "item"}}
`)
	writeFile(t, filepath.Join(projectRoot, "tables", "generated", "ignored.cue"), "package generated\n")
	writeFile(t, filepath.Join(projectRoot, "schema", "common", "common.cue"), `package common
#GameID: string
_schema: {definitions: {GameID: {kind: "type", type: "string"}}}
`)

	configRoot := t.TempDir()
	writeFile(t, SettingsPath(configRoot), `{
  "version": 1,
  "build": {"mode": "manual"}
}`)

	view, err := NewService(configRoot).OpenProject(filepath.Join(projectRoot, "game.dtproj"))
	if err != nil {
		t.Fatalf("OpenProject() error = %v", err)
	}
	if view.Settings.Build.Mode != "manual" {
		t.Fatalf("build mode = %q, want manual", view.Settings.Build.Mode)
	}
	if got := findNode(view.Tree.Root, "rpg.item"); got == nil || got.Kind != projecttree.NodeTable {
		t.Fatalf("table source was not discovered: %#v", view.Tree)
	}
	if got := findNodeByKind(view.Tree.Root, "common", projecttree.NodeSchemaFile); got == nil {
		t.Fatalf("schema source was not discovered: %#v", view.Tree)
	}
	if got := findNode(view.Tree.Root, "ignored.cue"); got != nil {
		t.Fatalf("excluded source was discovered: %#v", got)
	}
	if len(view.Schemas.Definitions) != 1 || view.Schemas.Definitions[0].ID != "common.#GameID" {
		t.Fatalf("schema index = %#v", view.Schemas)
	}
}

func TestOpenProjectIndexesSchemaDiagnosticsAndUsages(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{
  "version": 1,
  "tableRoots": ["tables"],
  "schemaRoots": ["schema"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)
	writeFile(t, filepath.Join(projectRoot, "schema", "common.cue"), `package common
#GameID: string
#Positive: >0
_schema: {definitions: {
  GameID: {kind: "type", type: "string"}
  Positive: {kind: "constraint", appliesTo: ["integer", "number"]}
}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: common.#GameID
#Row: {price: string & common.#Positive}
rows: {sword: "sword"}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	view, err := NewService(t.TempDir()).OpenProject(filepath.Join(projectRoot, "game.dtproj"))
	if err != nil {
		t.Fatalf("OpenProject() error = %v", err)
	}
	if len(view.Schemas.Definitions) != 2 {
		t.Fatalf("definitions = %#v", view.Schemas.Definitions)
	}
	if len(view.Schemas.Diagnostics) != 1 || view.Schemas.Diagnostics[0].Field != "price" {
		t.Fatalf("diagnostics = %#v", view.Schemas.Diagnostics)
	}
	if got := view.Schemas.Definitions[0].Usages; len(got) != 1 || got[0].Field != "#PrimaryKey" {
		t.Fatalf("GameID usages = %#v", got)
	}
}

func TestLoadEditorSettingsUsesDefaultsWhenFileIsMissing(t *testing.T) {
	settings, err := LoadEditorSettings(t.TempDir())
	if err != nil {
		t.Fatalf("LoadEditorSettings() error = %v", err)
	}
	if settings != DefaultEditorSettings() {
		t.Fatalf("settings = %#v, want %#v", settings, DefaultEditorSettings())
	}
}

func TestLoadProjectRejectsInvalidConfig(t *testing.T) {
	projectRoot := t.TempDir()
	path := filepath.Join(projectRoot, "invalid.dtproj")
	writeFile(t, path, `{
  "version": 1,
  "tableRoots": ["../outside"],
  "schemaRoots": ["schema"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)

	_, err := LoadProject(path)
	if err == nil || !strings.Contains(err.Error(), "escapes the project directory") {
		t.Fatalf("LoadProject() error = %v, want path escape error", err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func findNode(node projecttree.Node, name string) *projecttree.Node {
	if node.Name == name {
		return &node
	}
	for _, child := range node.Children {
		if found := findNode(child, name); found != nil {
			return found
		}
	}
	return nil
}

func findNodeByKind(node projecttree.Node, name string, kind projecttree.NodeKind) *projecttree.Node {
	if node.Name == name && node.Kind == kind {
		return &node
	}
	for _, child := range node.Children {
		if found := findNodeByKind(child, name, kind); found != nil {
			return found
		}
	}
	return nil
}
