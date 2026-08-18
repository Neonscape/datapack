package functional

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/compiler"
)

func TestValidationErrorsAreLocatableAndBlockExport(t *testing.T) {
	cases := []struct {
		name    string
		files   map[string]string
		rowID   string
		column  string
		message string
	}{
		{
			name: "type and range",
			files: map[string]string{
				"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {price: int & >=0}
rows: {sword: {price: -1}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "price"]}}
`,
			},
			rowID:  "sword",
			column: "price",
		},
		{
			name: "invalid primary key",
			files: map[string]string{
				"tables/item/item.cue": `package item
#PrimaryKey: string & =~"^[a-z][a-z0-9_]*$"
#Row: {name: string}
rows: {"1sword": {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}
`,
			},
			rowID: "1sword",
		},
		{
			name: "missing required field",
			files: map[string]string{
				"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}
`,
			},
			rowID: "sword",
		},
		{
			name: "invalid enum",
			files: map[string]string{
				"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {rarity: "N" | "R"}
rows: {sword: {rarity: "SSR"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "rarity"]}}
`,
			},
			rowID:  "sword",
			column: "rarity",
		},
		{
			name: "missing scalar foreign key",
			files: map[string]string{
				"tables/effect/effect.cue": `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}
`,
				"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {effectID: string}
rows: {sword: {effectID: "missing"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "effectID"]}}
`,
			},
			rowID:   "sword",
			column:  "effectID",
			message: "missing",
		},
		{
			name: "formula cycle",
			files: map[string]string{
				"tables/growth/growth.cue": `package growth
#PrimaryKey: string
#Row: {attack: int}
rows: {level_01: {}}
_tool: {
	table: {namespace: "combat", name: "growth"}
	foreignKeys: {}
	formulas: {level_01: {attack: "attack + 1"}}
}
_export: {csv: {enabled: true, columns: ["id", "attack"]}}
`,
			},
			rowID:  "level_01",
			column: "attack",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := writeProject(t, test.files)
			service := openProject(t, filepath.Join(root, "game.dtproj"))
			view, ok := service.Current()
			if !ok {
				t.Fatal("Current() returned no workspace")
			}
			if len(view.Compile.Diagnostics) == 0 {
				t.Fatal("compile produced no diagnostics")
			}
			if !hasDiagnosticAt(view.Compile.Diagnostics, test.rowID, test.column) {
				t.Fatalf("diagnostics = %s\nwant row %q column %q", formatDiagnostics(view.Compile.Diagnostics), test.rowID, test.column)
			}
			if test.message != "" && !diagnosticMessageContains(view.Compile.Diagnostics, test.message) {
				t.Fatalf("diagnostics = %s\nwant substring %q", formatDiagnostics(view.Compile.Diagnostics), test.message)
			}
			if _, err := service.Export(); err == nil {
				t.Fatal("Export() succeeded despite diagnostics")
			}
		})
	}
}

func TestMissingExportColumnBlocksExport(t *testing.T) {
	root := writeProject(t, map[string]string{
		"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "missing"]}}
`,
	})
	service := openProject(t, filepath.Join(root, "game.dtproj"))
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() accepted a missing export column")
	}
}

func writeProject(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	for relative, contents := range files {
		writeFile(t, filepath.Join(root, relative), contents)
	}
	return root
}

func diagnosticMessageContains(diagnostics []compiler.Diagnostic, fragment string) bool {
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Message, fragment) {
			return true
		}
	}
	return false
}
