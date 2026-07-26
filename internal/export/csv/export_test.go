package csv

import (
	"os"
	"path/filepath"
	"testing"

	"datapack/internal/document"
)

func TestExportWritesConfiguredCSV(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string, tags: [...string]}
rows: {sword: {name: "Sword", tags: ["weapon"]}}
_tool: {table: {namespace: "rpg", name: "item"}}
_export: {csv: {enabled: true, columns: ["id", "name", "tags"]}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Export(filepath.Join(root, "exports"), []document.Table{table})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("files = %#v", result.Files)
	}
	contents, err := os.ReadFile(result.Files[0])
	if err != nil || string(contents) != "id,name,tags\nsword,Sword,\"[\"\"weapon\"\"]\"\n" {
		t.Fatalf("CSV = %q, error = %v", contents, err)
	}
}

func TestExportLeavesExistingOutputWhenValidationFails(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int & >=0}
rows: {sword: {price: -1}}
_tool: {table: {namespace: "rpg", name: "item"}}
_export: {csv: {enabled: true, columns: ["id", "price"]}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(root, "exports")
	writeFile(t, filepath.Join(destination, "previous.csv"), "previous")
	if _, err := Export(destination, []document.Table{table}); err == nil {
		t.Fatal("Export() accepted an invalid table")
	}
	contents, err := os.ReadFile(filepath.Join(destination, "previous.csv"))
	if err != nil || string(contents) != "previous" {
		t.Fatalf("previous output = %q, error = %v", contents, err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
