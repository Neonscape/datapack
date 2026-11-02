package workspace

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func dependencyFixture(t *testing.T, imports, missingValue bool) (*Service, string, string) {
	t.Helper()
	t.Setenv("DATAPACK_FAST_EDIT", "1")
	t.Setenv("DATAPACK_COMPILE_CACHE", "1")
	root := t.TempDir()
	config := t.TempDir()
	path := filepath.Join(root, "project.dtproj")
	writeFile(t, path, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module: \"example.com/inputguard\"\nlanguage: {version: \"v0.9.0\"}\n")
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0755); err != nil {
		t.Fatal(err)
	}
	header := "package item\n#PrimaryKey: string\n#Row: {price: int}\n"
	if imports {
		writeFile(t, filepath.Join(root, "shared", "base", "base.cue"), "package base\n#Price: int & <=20\n")
		writeFile(t, filepath.Join(root, "shared", "common", "common.cue"), "package common\nimport \"example.com/inputguard/shared/base\"\n#Price: base.#Price\n")
		header = "package item\nimport \"example.com/inputguard/shared/common\"\n#PrimaryKey: string\n#Row: {price: common.#Price}\n"
	}
	row := "rows: {one: {price: 10}}\n"
	if missingValue {
		row = "rows: {one: {}}\n"
	}
	table := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, table, header+row+`_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "price"]}}
`)
	settings := DefaultEditorSettings()
	settings.Build.Mode = "manual"
	if err := SaveEditorSettings(config, settings); err != nil {
		t.Fatal(err)
	}
	s := NewService(config)
	if _, err := s.OpenProject(path); err != nil {
		t.Fatal(err)
	}
	return s, root, table
}

func TestAncestorConstraintChangesCannotReuseSuccess(t *testing.T) {
	for _, operation := range []string{"modify", "add", "delete", "rename"} {
		t.Run(operation, func(t *testing.T) {
			missing := operation == "delete" || operation == "rename"
			s, root, _ := dependencyFixture(t, false, missing)
			constraint := filepath.Join(root, "constraints.cue")
			if operation != "add" {
				body := "package item\n#Row: {price: <=20}\n"
				if missing {
					body = "package item\n#Row: {price: *10 | int}\n"
				}
				writeFile(t, constraint, body)
			}
			if _, err := s.Export(); err != nil {
				t.Fatal("initial export:", err)
			}
			output := filepath.Join(root, "exports", "rpg.item.csv")
			before, _ := os.ReadFile(output)
			switch operation {
			case "modify", "add":
				writeFile(t, constraint, "package item\n#Row: {price: <=5}\n")
			case "delete":
				if err := os.Remove(constraint); err != nil {
					t.Fatal(err)
				}
			case "rename":
				if err := os.Rename(constraint, constraint+".txt"); err != nil {
					t.Fatal(err)
				}
			}
			compiled, err := s.Compile()
			if err != nil {
				t.Fatal(err)
			}
			if len(compiled.Workspace.Compile.Diagnostics) == 0 {
				t.Fatal("stale successful result survived ancestor constraint change")
			}
			if _, err := s.Export(); err == nil {
				t.Fatal("illegal current inputs exported")
			}
			after, _ := os.ReadFile(output)
			if !bytes.Equal(before, after) {
				t.Fatal("rejected export replaced previous valid output")
			}
		})
	}
}
func TestTransitiveImportAndModuleChangesInvalidate(t *testing.T) {
	s, root, _ := dependencyFixture(t, true, false)
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "shared", "base", "base.cue"), "package base\n#Price: int & <=5\n")
	if _, err := s.Export(); err == nil {
		t.Fatal("transitive import constraint change not validated")
	}
	writeFile(t, filepath.Join(root, "shared", "base", "base.cue"), "package base\n#Price: int & <=20\n")
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), "module: 123\nlanguage: {version: \"v0.9.0\"}\n")
	if _, err := s.Export(); err == nil {
		t.Fatal("invalid current module used previous success")
	}
}
