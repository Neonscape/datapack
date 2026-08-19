package projecttree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanKeepsInvalidTableAsDamagedNode(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, "{}")
	writeFile(t, filepath.Join(root, "tables", "broken.cue"), "package broken\n")
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	tree, err := Scan(projectPath, []string{"tables"}, []string{"schema"}, nil)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	broken := find(tree.Root, "broken.cue")
	if broken == nil || broken.Kind != NodeDamaged {
		t.Fatalf("broken table = %#v, want damaged node", broken)
	}
	if len(broken.Diagnostics) == 0 {
		t.Fatal("damaged node has no diagnostics")
	}
}

func TestScanBuildsLogicalTableAndLocalSchemaNodes(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, "{}")
	writeFile(t, filepath.Join(root, "tables", "rpg", "item", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_local: {#Position: {x: number, y: number}}
_tool: {table: {namespace: "rpg", name: "item"}}
`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	tree, err := Scan(projectPath, []string{"tables"}, []string{"schema"}, nil)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	namespace := find(tree.Root, "rpg")
	if namespace == nil || namespace.Kind != NodeNamespace {
		t.Fatalf("namespace = %#v, want namespace node", namespace)
	}
	table := find(tree.Root, "item")
	if table == nil || table.Kind != NodeTable || table.Path != filepath.Join(root, "tables", "rpg", "item", "item.cue") {
		t.Fatalf("table = %#v, want logical table", table)
	}
	if got := find(*table, "local_schema"); got == nil || got.Kind != NodeLocalSchema {
		t.Fatalf("local schema = %#v", got)
	}
	if got := find(*table, "#Position"); got == nil || got.Kind != NodeLocalSchema {
		t.Fatalf("local definition = %#v", got)
	}
}

func TestScanBuildsSchemaDefinitionNodes(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, "{}")
	if err := os.MkdirAll(filepath.Join(root, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "schema", "common", "common.cue"), `package common
#GameID: string
_schema: {definitions: {GameID: {kind: "type", type: "string"}}}
`)

	tree, err := Scan(projectPath, []string{"tables"}, []string{"schema"}, nil)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	namespace := find(tree.Root, "common")
	if namespace == nil || namespace.Kind != NodeNamespace {
		t.Fatalf("namespace = %#v, want namespace node", namespace)
	}
	definition := find(*namespace, "#GameID")
	if definition == nil || definition.Kind != NodeSchemaDefinition {
		t.Fatalf("schema definition = %#v", definition)
	}
}

func TestScanKeepsEmptySchemaFileNode(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, "{}")
	if err := os.MkdirAll(filepath.Join(root, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "schema", "item", "item.cue"), `package item
_schema: {definitions: {}}
`)

	tree, err := Scan(projectPath, []string{"tables"}, []string{"schema"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	file := find(tree.Root, "item.cue")
	if file == nil || file.Kind != NodeSchemaFile || file.Path != filepath.Join(root, "schema", "item", "item.cue") {
		t.Fatalf("schema file = %#v", file)
	}
}

func TestScanGroupsSameNameTablesByNamespace(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, "{}")
	writeFile(t, filepath.Join(root, "tables", "rpg", "item", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}
`)
	writeFile(t, filepath.Join(root, "tables", "combat", "item", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {fist: {name: "Fist"}}
_tool: {table: {namespace: "combat", name: "item"}}
`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	tree, err := Scan(projectPath, []string{"tables"}, []string{"schema"}, nil)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	rpg := find(tree.Root, "rpg")
	combat := find(tree.Root, "combat")
	if rpg == nil || combat == nil {
		t.Fatalf("namespaces missing: rpg=%v combat=%v", rpg, combat)
	}
	if find(*rpg, "item") == nil || find(*combat, "item") == nil {
		t.Fatal("same-named tables in different namespaces were not both discovered")
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

func find(node Node, name string) *Node {
	if node.Name == name {
		return &node
	}
	for _, child := range node.Children {
		if found := find(child, name); found != nil {
			return found
		}
	}
	return nil
}
