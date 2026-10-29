package document

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRenameRowUpdatesSameRowForeignKeys(t *testing.T) {
	for _, array := range []bool{false, true} {
		name, schema, value, want := "scalar", "string", `"burn"`, `"flame"`
		if array {
			name, schema, value, want = "array", "[...string]", `["burn", "other", "burn"]`, `["flame", "other", "flame"]`
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "effect.cue")
			other := filepath.Join(root, "item.cue")
			writeModuleFile(t, target, `package effect
#PrimaryKey: string
#Row: {link: `+schema+`, label: string}
rows: {
    burn: {link: `+value+`, label: "burn"}
    other: {link: `+value+`, label: "other"}
}
_tool: {table: {namespace: "rpg", name: "effect"}}`)
			// The other table deliberately has the same row ID as the renamed
			// row: a self-reference fix must not redirect this row lookup.
			writeModuleFile(t, other, `package item
#PrimaryKey: string
#Row: {link: `+schema+`}
rows: {burn: {link: `+value+`}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
			var references []RowReference
			for _, location := range []struct{ path, row string }{{target, "burn"}, {target, "other"}, {other, "burn"}} {
				references = append(references, RowReference{Path: location.path, RowID: location.row, Column: "link", Array: array})
				if array {
					references = append(references, RowReference{Path: location.path, RowID: location.row, Column: "link", Array: true, Index: 2})
				}
			}
			table, err := RenameRow(target, "burn", "flame", references)
			if err != nil {
				t.Fatal(err)
			}
			if len(table.Rows) != 2 || table.Rows[0].ID != "flame" || table.Rows[1].ID != "other" {
				t.Fatalf("renamed rows = %#v", table.Rows)
			}
			for _, row := range table.Rows {
				if row.Values["link"].Expression != want {
					t.Fatalf("row %s link = %s, want %s", row.ID, row.Values["link"].Expression, want)
				}
			}
			if table.Rows[0].Values["label"].Display != "burn" {
				t.Fatal("rename changed an unrelated string")
			}
			otherTable, err := LoadTable(other)
			if err != nil || len(otherTable.Rows) != 1 || otherTable.Rows[0].ID != "burn" || otherTable.Rows[0].Values["link"].Expression != want {
				t.Fatalf("other table = %#v, %v", otherTable, err)
			}
		})
	}
}

func TestRenameRowMovesCellComments(t *testing.T) {
	path := writeSource(t, "effect.cue", `package effect
#PrimaryKey: string
#Row: {name: string, power: int}
rows: {burn: {name: "Burn", power: 1}, other: {name: "Other", power: 2}}
_tool: {table: {namespace: "rpg", name: "effect"}}
_editor: {
    columns: {name: {title: "Effect name", width: 240}}
    cellComments: {
        // Keep this source comment with its row.
        "burn": {
            name: "第一行\n带 \"引号\" 和 \\ 符号"
            power: "power note"
        }
        other: {name: "other name", power: "other power"}
        orphan: {name: "unrelated orphan metadata"}
    }
}`)
	before, err := LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]string{
		"flame": before.CellComments["burn"], "other": before.CellComments["other"], "orphan": before.CellComments["orphan"],
	}
	table, err := RenameRow(path, "burn", "flame", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(table.CellComments, want) || !reflect.DeepEqual(table.Columns, before.Columns) {
		t.Fatalf("metadata after rename = %#v, columns = %#v", table.CellComments, table.Columns)
	}
	source, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(source), "// Keep this source comment with its row.") {
		t.Fatalf("source comment lost: %s, %v", source, err)
	}
	// Reusing the old ID must not resurrect the renamed row's comments.
	table, err = AddRow(path, "burn", `{name: "New Burn", power: 3}`)
	if err != nil || !reflect.DeepEqual(table.CellComments, want) {
		t.Fatalf("comments after reusing old ID = %#v, %v", table.CellComments, err)
	}
}

func TestRenameRowRejectsDestinationCellCommentsWithoutWrites(t *testing.T) {
	for _, oldComments := range []string{"", `burn: {name: "original note"}`} {
		for _, destination := range []string{`{name: "orphan note"}`, `{}`} {
			t.Run(oldComments+destination, func(t *testing.T) {
				root := t.TempDir()
				target := filepath.Join(root, "effect.cue")
				other := filepath.Join(root, "item.cue")
				writeModuleFile(t, target, `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}}
_editor: {cellComments: {`+oldComments+`
flame: `+destination+`}}`)
				writeModuleFile(t, other, `package item
#PrimaryKey: string
#Row: {link: string}
rows: {sword: {link: "burn"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
				before := rowRenameSources(t, root)
				_, err := RenameRow(target, "burn", "flame", []RowReference{{Path: other, RowID: "sword", Column: "link"}})
				if err == nil || !strings.Contains(err.Error(), "cell comments") || !strings.Contains(err.Error(), `"flame"`) {
					t.Fatalf("destination comment conflict = %v", err)
				}
				if after := rowRenameSources(t, root); !reflect.DeepEqual(after, before) {
					t.Fatalf("failed rename changed files: got %#v, want %#v", after, before)
				}
			})
		}
	}
}

func TestRenameRowStaleReferenceLeavesCommentsAndAllSourcesUnchanged(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "effect.cue")
	other := filepath.Join(root, "item.cue")
	writeModuleFile(t, target, `package effect
#PrimaryKey: string
#Row: {link: string}
rows: {burn: {link: "burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}}
_editor: {cellComments: {burn: {link: "self note"}}}`)
	writeModuleFile(t, other, `package item
#PrimaryKey: string
#Row: {link: string}
rows: {sword: {link: "burn"}, stale: {link: "other"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	before := rowRenameSources(t, root)
	_, err := RenameRow(target, "burn", "flame", []RowReference{
		{Path: target, RowID: "burn", Column: "link"},
		{Path: other, RowID: "sword", Column: "link"},
		{Path: other, RowID: "stale", Column: "link"},
	})
	if err == nil || !strings.Contains(err.Error(), "no longer references") {
		t.Fatalf("stale reference error = %v", err)
	}
	if after := rowRenameSources(t, root); !reflect.DeepEqual(after, before) {
		t.Fatalf("failed rename changed files: got %#v, want %#v", after, before)
	}
}

func rowRenameSources(t *testing.T, root string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, entry := range entries {
		source, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		sources[entry.Name()] = string(source)
	}
	return sources
}

func TestRenameRowRejectsMalformedCommentContainersWithoutWrites(t *testing.T) {
	for _, editor := range []string{`"bad"`, `{cellComments:"bad"}`, `{cellComments:{burn:"bad"}}`} {
		t.Run(editor, func(t *testing.T) {
			path := writeSource(t, "effect.cue", `package effect
#PrimaryKey:string
#Row:{name:string}
rows:{burn:{name:"Burn"}}
_tool:{table:{namespace:"rpg",name:"effect"}}
_editor:`+editor)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := RenameRow(path, "burn", "flame", nil); err == nil || !strings.Contains(err.Error(), "must be a struct") {
				t.Fatalf("malformed metadata accepted: %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(before) != string(after) {
				t.Fatal("failed rename changed malformed metadata")
			}
		})
	}
}
