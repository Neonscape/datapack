package workspace

import (
	"reflect"
	"strings"
	"testing"

	"datapack/internal/document"
)

func rowRenameMetadataFixture(t *testing.T) (*Service, string) {
	t.Helper()
	s, target, _ := renameFixture(t)
	writeFile(t, target, `package effect
#PrimaryKey: string
#Row: {name: string, parentID: string, parentIDs: [...string]}
rows: {
    burn: {name: "Burn", parentID: "burn", parentIDs: ["burn", "other", "burn"]}
    other: {name: "Other", parentID: "burn", parentIDs: ["burn"]}
}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {
    parentID: {table: "rpg.effect"}
    parentIDs: {table: "rpg.effect"}
}}
_editor: {cellComments: {
    burn: {name: "renamed row note", parentID: "self-reference note"}
    other: {name: "unchanged row note"}
}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	return s, target
}

func TestRenameRowUndoRedoRestoresSelfReferencesAndCellComments(t *testing.T) {
	s, target := rowRenameMetadataFixture(t)
	root := s.current.Project.Root
	before := fileTree(t, root)
	result, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"})
	if err != nil {
		t.Fatal(err)
	}
	wantComments := map[string]map[string]string{
		"flame": {"name": "renamed row note", "parentID": "self-reference note"},
		"other": {"name": "unchanged row note"},
	}
	if !reflect.DeepEqual(result.Table.CellComments, wantComments) || len(result.Workspace.ForeignKeys.Diagnostics) != 0 {
		t.Fatalf("renamed comments = %#v, diagnostics = %#v", result.Table.CellComments, result.Workspace.ForeignKeys.Diagnostics)
	}
	if len(result.Workspace.Compile.Diagnostics) != 0 {
		t.Fatalf("renamed source no longer compiles: %#v", result.Workspace.Compile.Diagnostics)
	}
	if result.Table.Rows[0].ID != "flame" || result.Table.Rows[0].Values["parentID"].Display != "flame" || result.Table.Rows[0].Values["parentIDs"].Expression != `["flame", "other", "flame"]` {
		t.Fatalf("renamed self references = %#v", result.Table.Rows)
	}
	after := fileTree(t, root)
	undo, err := s.Undo()
	if err != nil || undo.CanUndo || !undo.CanRedo || len(undo.Workspace.ForeignKeys.Diagnostics) != 0 || len(undo.Workspace.Compile.Diagnostics) != 0 {
		t.Fatalf("Undo(rename) = %#v, %v", undo, err)
	}
	assertFileTree(t, root, before)
	redo, err := s.Redo()
	if err != nil || !redo.CanUndo || redo.CanRedo || len(redo.Workspace.ForeignKeys.Diagnostics) != 0 || len(redo.Workspace.Compile.Diagnostics) != 0 {
		t.Fatalf("Redo(rename) = %#v, %v", redo, err)
	}
	assertFileTree(t, root, after)
	table, err := document.LoadTable(target)
	if err != nil || !reflect.DeepEqual(table.CellComments, wantComments) {
		t.Fatalf("redo comments = %#v, %v", table.CellComments, err)
	}
}

func TestRenameRowCommentCollisionPreservesFilesAndHistory(t *testing.T) {
	s, target := rowRenameMetadataFixture(t)
	// Destination metadata can outlive its row. It must not be overwritten
	// or silently attached to the renamed row, even if the old row has notes.
	if _, err := s.SetCellComment(CellCommentSet{Path: target, RowID: "flame", Column: "name", Comment: "orphan note"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	root := s.current.Project.Root
	before := fileTree(t, root)
	viewBefore, undoBefore, redoBefore := s.current, s.undoStack, s.redoStack
	_, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"})
	if err == nil || !strings.Contains(err.Error(), "cell comments") {
		t.Fatalf("comment collision error = %v", err)
	}
	assertFileTree(t, root, before)
	if !reflect.DeepEqual(s.current, viewBefore) || !reflect.DeepEqual(s.undoStack, undoBefore) || !reflect.DeepEqual(s.redoStack, redoBefore) {
		t.Fatal("failed row rename changed workspace or undo/redo history")
	}
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	table, err := document.LoadTable(target)
	if err != nil || table.Rows[0].ID != "burn" || table.Rows[0].Values["name"].Display != "Hot" || table.CellComments["flame"]["name"] != "orphan note" {
		t.Fatalf("existing redo did not survive failed rename: %#v, %v", table, err)
	}
}
