package functional

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/projecttree"
	"datapack/internal/workspace"
)

func TestInvalidCueExpressionDoesNotWriteTheSource(t *testing.T) {
	root := writeProject(t, map[string]string{
		"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
`,
	})
	path := filepath.Join(root, "tables", "item", "item.cue")
	before := readFile(t, path)
	service := openProject(t, filepath.Join(root, "game.dtproj"))
	if _, err := service.UpdateCell(workspace.CellUpdate{Path: path, RowID: "sword", Column: "name", Expression: "{"}); err == nil {
		t.Fatal("UpdateCell() accepted invalid CUE")
	}
	if readFile(t, path) != before {
		t.Fatal("invalid CUE overwrote the table source")
	}
}

func TestBusinessErrorSavesWithoutExporting(t *testing.T) {
	root := writeProject(t, map[string]string{
		"tables/item/item.cue": `package item
#PrimaryKey: string
#Row: {price: int & >=0}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "price"]}}
`,
	})
	path := filepath.Join(root, "tables", "item", "item.cue")
	service := openProject(t, filepath.Join(root, "game.dtproj"))
	if _, err := service.Export(); err != nil {
		t.Fatalf("initial Export(): %v", err)
	}
	if _, err := service.UpdateCell(workspace.CellUpdate{Path: path, RowID: "sword", Column: "price", Expression: "-4"}); err != nil {
		t.Fatalf("UpdateCell(-4): %v", err)
	}
	if got := readFile(t, path); !strings.Contains(got, "price: -4") {
		t.Fatalf("saved source = %s", got)
	}
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() accepted a business error")
	}
}

func TestDamagedTableStaysOnDiskAndBlocksExport(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(root, "tables", "broken.cue")
	original := "this is not valid CUE {"
	writeFile(t, broken, original)

	service := openProject(t, filepath.Join(root, "game.dtproj"))
	view, ok := service.Current()
	if !ok {
		t.Fatal("Current() returned no workspace")
	}
	node := findNode(view.Tree.Root, "broken.cue")
	if node == nil || node.Kind != projecttree.NodeDamaged {
		t.Fatalf("damaged node = %#v", node)
	}
	if len(node.Diagnostics) == 0 {
		t.Fatal("damaged node has no diagnostics")
	}
	if readFile(t, broken) != original {
		t.Fatal("opening the project rewrote the damaged file")
	}
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() succeeded with a damaged table")
	}
	if readFile(t, broken) != original {
		t.Fatal("export attempt rewrote the damaged file")
	}
}
