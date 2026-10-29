package csv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"datapack/internal/compiler"
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
	compiled := compiler.CompileAll(nil, []document.Table{table})
	if compiled.Package == nil {
		t.Fatalf("compile diagnostics = %#v", compiled.Diagnostics)
	}
	result, err := Export(filepath.Join(root, "exports"), compiled.Package)
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

func TestExportBlocksOnValidationProblems(t *testing.T) {
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
	compiled := compiler.CompileAll(nil, []document.Table{table})
	if compiled.Package != nil {
		t.Fatal("CompileAll() produced a package despite invalid rows")
	}
	if _, err := Export(destination, compiled.Package); err == nil {
		t.Fatal("Export() accepted a nil compiled package")
	}
	contents, err := os.ReadFile(filepath.Join(destination, "previous.csv"))
	if err != nil || string(contents) != "previous" {
		t.Fatalf("previous output = %q, error = %v", contents, err)
	}
}

func TestExportCalculatesFormulaValuesBeforeValidation(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "growth.cue")
	writeFile(t, path, `package growth
#PrimaryKey: string
#Row: {baseAttack: int, attack: int & >=0}
rows: {
    level_01: {baseAttack: 10}
    level_10: {baseAttack: 20}
}

_tool: {
    table: {namespace: "combat", name: "growth"}
    formulas: {
        level_01: {attack: "baseAttack * 2"}
        level_10: {attack: "combat.growth.level_01.attack + baseAttack"}
    }
}
_export: {csv: {enabled: true, columns: ["id", "baseAttack", "attack"]}}`)
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	compiled := compiler.CompileAll(nil, []document.Table{table})
	if compiled.Package == nil {
		t.Fatalf("compile diagnostics = %#v", compiled.Diagnostics)
	}
	result, err := Export(filepath.Join(root, "exports"), compiled.Package)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(result.Files[0])
	if err != nil || string(contents) != "id,baseAttack,attack\nlevel_01,10,20\nlevel_10,20,40\n" {
		t.Fatalf("CSV = %q, error = %v", contents, err)
	}
}

func TestExportRejectsUnsafeNamesBeforeWriting(t *testing.T) {
	for _, test := range []struct{ namespace, name string }{
		{namespace: ".", name: "/previous"},
		{namespace: ".", name: `\previous`},
		{name: "../previous"},
		{name: "/absolute"},
		{name: `C:\previous`},
		{name: "stream:previous"},
		{name: "NUL"},
		{name: "con.data"},
		{name: "COM1"},
		{name: "trailing."},
		{name: "trailing "},
		{name: "line\nbreak"},
		{name: "nul\x00byte"},
	} {
		t.Run(test.namespace+test.name, func(t *testing.T) {
			root := t.TempDir()
			destination := filepath.Join(root, "exports")
			writeFile(t, filepath.Join(root, "previous.csv"), "must survive")
			writeFile(t, filepath.Join(destination, "old.csv"), "old output")
			before, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			pkg := &compiler.Package{Tables: []compiler.Table{
				{Name: "valid", CSV: &compiler.CSV{Columns: []string{"id"}}},
				{Namespace: test.namespace, Name: test.name, CSV: &compiler.CSV{Columns: []string{"id"}}},
			}}
			if _, err := Export(destination, pkg); err == nil {
				t.Fatal("Export accepted an unsafe filename")
			}
			for path, want := range map[string]string{filepath.Join(root, "previous.csv"): "must survive", filepath.Join(destination, "old.csv"): "old output"} {
				if got, err := os.ReadFile(path); err != nil || string(got) != want {
					t.Fatalf("%s = %q, error = %v", path, got, err)
				}
			}
			after, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("rejected export left staging files")
			}
		})
	}
}

func TestExportRejectsCollidingNamesBeforeWriting(t *testing.T) {
	for _, name := range []string{"item", "ITEM"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			destination := filepath.Join(root, "exports")
			writeFile(t, filepath.Join(destination, "previous.csv"), "previous")
			pkg := &compiler.Package{Tables: []compiler.Table{
				{Name: "item", CSV: &compiler.CSV{Columns: []string{"id"}}},
				{Name: name, CSV: &compiler.CSV{Columns: []string{"id"}}},
			}}
			if _, err := Export(destination, pkg); err == nil {
				t.Fatal("Export accepted colliding filenames")
			}
			if got, err := os.ReadFile(filepath.Join(destination, "previous.csv")); err != nil || string(got) != "previous" {
				t.Fatalf("previous output = %q, error = %v", got, err)
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 1 {
				t.Fatalf("rejected export left staging files: %#v, %v", entries, err)
			}
		})
	}
}

func TestExportPreservesUnrelatedLegacyBackupAndReplacesOldOutput(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "exports")
	writeFile(t, filepath.Join(destination, "old.csv"), "old output")
	legacyBackup := filepath.Join(destination+".datapack-backup", "source.cue")
	writeFile(t, legacyBackup, "source data")
	pkg := &compiler.Package{Tables: []compiler.Table{{Namespace: "rpg.items", Name: "loot.v2", CSV: &compiler.CSV{Columns: []string{"id"}}}}}
	for attempt := 0; attempt < 2; attempt++ {
		result, err := Export(destination, pkg)
		if err != nil || len(result.Files) != 1 || filepath.Base(result.Files[0]) != "rpg.items.loot.v2.csv" {
			t.Fatalf("Export() = %#v, %v", result, err)
		}
	}
	if got, err := os.ReadFile(legacyBackup); err != nil || string(got) != "source data" {
		t.Fatalf("legacy backup = %q, error = %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(destination, "old.csv")); !os.IsNotExist(err) {
		t.Fatalf("old output remains: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 2 {
		t.Fatalf("export left staging directories: %#v, %v", entries, err)
	}
}

func TestReplaceDirectoryRestoresPreviousOutputOnFailure(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "exports")
	writeFile(t, filepath.Join(destination, "previous.csv"), "previous")
	if err := replaceDirectory(destination, filepath.Join(root, "missing-stage")); err == nil {
		t.Fatal("expected replacement to fail")
	}
	if got, err := os.ReadFile(filepath.Join(destination, "previous.csv")); err != nil || string(got) != "previous" {
		t.Fatalf("restored output = %q, error = %v", got, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("rollback left staging directories: %#v, %v", entries, err)
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
