package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"datapack/internal/document"
)

func TestFormulaSafetyCompileRejectsDefaultMaskedCycles(t *testing.T) {
	for _, formulas := range []string{`a: "b+1", b: "a+1"`, `a: "b == nil ? 9 : b+1", b: "mystery(1)"`, `a: "a+1"`} {
		t.Run(formulas, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "item.cue")
			source := `package item
#PrimaryKey: string
#Row: {a: *0 | int, b: *0 | int}
rows: {one: {}}
_tool: {table: {namespace: "rpg", name: "item"}, formulas: {one: {` + formulas + `}}}
_export: {csv: {enabled: true, columns: ["id", "a", "b"]}}
`
			writeFile(t, path, source)
			table, err := document.LoadTable(path)
			if err != nil {
				t.Fatal(err)
			}
			result := CompileAll(nil, []document.Table{table})
			if result.Package != nil || len(result.Diagnostics) == 0 {
				t.Fatalf("unsafe compile=%#v", result)
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != source {
				t.Fatal("compile changed source")
			}
		})
	}
}

func TestFormulaSafetyCompileUsesResolvedDependencies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {a: *0 | int, b: *0 | int}
rows: {one: {}}
_tool: {table: {namespace: "rpg", name: "item"}, formulas: {one: {a: "b+1", b: "10"}}}
`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := CompileAll(nil, []document.Table{table})
	if len(result.Diagnostics) != 0 || result.Package == nil {
		t.Fatal(result)
	}
	got := result.Package.Tables[0].Rows[0].Values["a"]
	if fmt.Sprint(got) != "11" {
		t.Fatalf("a=%v", got)
	}
}

func TestFormulaSafetyNamedIntegerTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
_local: {#Level: int}
_schema: {localDefinitions: {Level: {kind: "type", type: "integer"}}}
#Row: {a: _local.#Level}
rows: {one: {}}
_tool: {table: {namespace: "rpg", name: "item"}, formulas: {one: {a: "round(1.4)"}}}
`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := CompileAll(nil, []document.Table{table})
	if len(result.Diagnostics) != 0 || result.Package == nil {
		t.Fatalf("named integer formula failed: %#v", result)
	}
	if got := fmt.Sprint(result.Package.Tables[0].Rows[0].Values["a"]); got != "1" {
		t.Fatalf("a=%s", got)
	}
}
