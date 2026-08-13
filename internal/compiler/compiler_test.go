package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"datapack/internal/document"
)

func TestCompileAllProducesCleanPackage(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {baseAttack: int, attack: int & >=0}
rows: {level_01: {baseAttack: 10}}
_tool: {table: {namespace: "combat", name: "item"}, formulas: {level_01: {attack: "baseAttack * 2"}}}
_export: {csv: {enabled: true, columns: ["id", "attack"]}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := CompileAll(nil, []document.Table{table})
	if len(result.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
	if result.Package == nil || len(result.Package.Tables) != 1 {
		t.Fatalf("package = %#v", result.Package)
	}
	compiled := result.Package.Tables[0]
	if len(compiled.Rows) != 1 || fmt.Sprint(compiled.Rows[0].Values["attack"]) != "20" {
		t.Fatalf("compiled rows = %#v", compiled.Rows)
	}
	if compiled.CSV == nil || len(compiled.CSV.Columns) != 2 {
		t.Fatalf("csv = %#v", compiled.CSV)
	}
}

func TestCompileAllReportsRowAndSchemaProblems(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int & >=0}
rows: {sword: {price: -1}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := CompileAll(nil, []document.Table{table})
	if result.Package != nil {
		t.Fatalf("package = %#v, want nil", result.Package)
	}
	found := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.RowID == "sword" && diagnostic.Column == "price" {
			found = true
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestCompileAllReportsMissingSchemaReferences(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: common.#Missing}
rows: {sword: {price: 1}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := CompileAll(nil, []document.Table{table})
	if result.Package != nil {
		t.Fatalf("package = %#v, want nil", result.Package)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}
}

func TestCompileAllReportsForeignKeyProblems(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {effectID: string}
rows: {sword: {effectID: "missing"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := CompileAll(nil, []document.Table{table})
	if result.Package != nil || len(result.Diagnostics) == 0 {
		t.Fatalf("result = %#v", result)
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
