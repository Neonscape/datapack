package workspace

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"datapack/internal/document"
)

func renameFixture(t *testing.T) (*Service, string, string) {
	t.Helper()
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"],"tables":{"rpg.effect":true,"rpg.item":false}}}`)
	target := filepath.Join(root, "tables", "rpg", "effect", "effect.cue")
	writeFile(t, target, `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {}}`)
	ref := filepath.Join(root, "tables", "rpg", "item", "item.cue")
	writeFile(t, ref, `package item
#PrimaryKey: string
#Row: {effectID: string, effectIDs: [...string]}
rows: {sword: {effectID: "burn", effectIDs: ["burn", "burn"]}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}, effectIDs: {table: "rpg.effect"}}}`)
	writeFile(t, filepath.Join(root, "tables", "rpg", "skill", "skill.cue"), `package skill
#PrimaryKey: string
#Row: {effectID: string}
rows: {fireball: {effectID: "burn"}}
_tool: {table: {namespace: "rpg", name: "skill"}, foreignKeys: {effectID: {table: "rpg.effect"}}}`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(projectPath); err != nil {
		t.Fatal(err)
	}
	return service, target, ref
}

func fileTree(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files[relative] = string(content)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func assertFileTree(t *testing.T, root string, want map[string]string) {
	t.Helper()
	if got := fileTree(t, root); !reflect.DeepEqual(got, want) {
		t.Fatalf("file tree changed: got %#v, want %#v", got, want)
	}
}

func TestRenameRowUndoRedoRestoresAllReferences(t *testing.T) {
	s, target, _ := renameFixture(t)
	root := s.current.Project.Root
	original := fileTree(t, root)
	if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
		t.Fatal(err)
	}
	beforeRename := fileTree(t, root)
	result, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Workspace.CanUndo || result.Workspace.CanRedo || len(result.Workspace.ForeignKeys.Diagnostics) != 0 {
		t.Fatalf("rename history/FK state = %#v", result.Workspace)
	}
	afterRename := fileTree(t, root)
	undo, err := s.Undo()
	if err != nil || !undo.CanUndo || !undo.CanRedo {
		t.Fatalf("Undo(rename) = %#v, %v", undo, err)
	}
	assertFileTree(t, root, beforeRename)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, root, original)
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, root, beforeRename)
	redo, err := s.Redo()
	if err != nil || !redo.CanUndo || redo.CanRedo {
		t.Fatalf("Redo(rename) = %#v, %v", redo, err)
	}
	assertFileTree(t, root, afterRename)
}

func TestRenameRowClearsRedo(t *testing.T) {
	s, target, _ := renameFixture(t)
	if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	result, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Workspace.CanUndo || result.Workspace.CanRedo {
		t.Fatalf("rename did not replace redo branch: undo=%v redo=%v", result.Workspace.CanUndo, result.Workspace.CanRedo)
	}
	if _, err := s.Redo(); err == nil {
		t.Fatal("Redo succeeded after a new rename")
	}
}

func TestRenameTableUpdatesSelfReferences(t *testing.T) {
	s, target, ref := renameFixture(t)
	writeFile(t, target, `package effect
#PrimaryKey: string
#Row: {name: string, parentID: string}
rows: {burn: {name: "Burn", parentID: "burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {parentID: {table: "rpg.effect"}}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	root := s.current.Project.Root
	before := fileTree(t, root)
	result, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{result.Table.Path, ref} {
		table, err := document.LoadTable(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range table.ForeignKeys {
			if key.Table != "magic.spell" {
				t.Fatalf("%s foreign key = %#v", path, key)
			}
		}
	}
	if len(result.Workspace.ForeignKeys.Diagnostics) != 0 {
		t.Fatalf("foreign-key diagnostics = %#v", result.Workspace.ForeignKeys.Diagnostics)
	}
	after := fileTree(t, root)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, root, before)
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, root, after)
}

func TestRenameTableFailedMoveLeavesAllFilesUnchanged(t *testing.T) {
	s, target, _ := renameFixture(t)
	root := s.current.Project.Root
	writeFile(t, filepath.Join(filepath.Dir(target), "sidecar.bin"), "preserve source asset")
	writeFile(t, filepath.Join(root, "tables", "magic", "spell", "sentinel.bin"), "preserve destination asset")
	before := fileTree(t, root)
	projectBefore, _ := os.ReadFile(s.current.Project.Path)
	if _, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"}); err == nil {
		t.Fatal("RenameTable unexpectedly replaced a nonempty destination")
	}
	assertFileTree(t, root, before)
	projectAfter, _ := os.ReadFile(s.current.Project.Path)
	if !bytes.Equal(projectAfter, projectBefore) || !s.current.Project.Config.Export.Tables["rpg.effect"] {
		t.Fatal("failed rename changed export selection")
	}
	if len(s.undoStack) != 0 || len(s.redoStack) != 0 {
		t.Fatal("failed rename changed history")
	}
}

func TestRenameTableRollsBackEveryFailureBoundary(t *testing.T) {
	injected := errors.New("injected rename failure")
	for _, failure := range []string{"partial source write", "parent mkdir", "directory move", "filename move", "project write"} {
		t.Run(failure, func(t *testing.T) {
			s, target, ref := renameFixture(t)
			root := s.current.Project.Root
			writeFile(t, filepath.Join(filepath.Dir(target), "assets", "sidecar.bin"), "preserve source asset")
			// Keep both stacks populated to verify failure does not consume or
			// invalidate earlier history.
			for _, name := range []string{`"Hot"`, `"Hotter"`} {
				if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: name}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			modes := map[string]os.FileMode{target: 0o640, ref: 0o604, s.current.Project.Path: 0o600}
			for path, mode := range modes {
				if err := os.Chmod(path, mode); err != nil {
					t.Fatal(err)
				}
				info, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				modes[path] = info.Mode().Perm()
			}
			before := fileTree(t, root)
			viewBefore := s.current
			undoBefore, redoBefore := s.undoStack, s.redoStack
			ops := defaultTableRenameOps()
			switch failure {
			case "partial source write":
				ops.rewrite = func(path, namespace, name string, refs []document.TableReference) (document.Table, error) {
					// Replace the source only, then fail before FK files commit.
					if _, err := document.RenameTable(path, namespace, name, nil); err != nil {
						t.Fatal(err)
					}
					return document.Table{}, injected
				}
			case "parent mkdir":
				ops.mkdirAll = func(string, os.FileMode) error { return injected }
			case "directory move", "filename move":
				calls := 0
				failAt := 1
				if failure == "filename move" {
					failAt = 2
				}
				ops.rename = func(from, to string) error {
					calls++
					if calls == failAt {
						return injected
					}
					return os.Rename(from, to)
				}
			case "project write":
				ops.saveProject = func(project Project) error {
					// Include rollback after a partially successful final step.
					if err := SaveProject(project); err != nil {
						t.Fatal(err)
					}
					return injected
				}
			}
			_, err := s.renameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"}, ops)
			if !errors.Is(err, injected) {
				t.Fatalf("RenameTable error = %v, want injected failure", err)
			}
			assertFileTree(t, root, before)
			for path, mode := range modes {
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm() != mode {
					t.Fatalf("rollback did not preserve permissions for %s: %v", path, err)
				}
			}
			if s.current != viewBefore || !s.current.Project.Config.Export.Tables["rpg.effect"] || s.current.Project.Config.Export.Tables["magic.spell"] {
				t.Fatal("failed rename changed in-memory project")
			}
			if !reflect.DeepEqual(s.undoStack, undoBefore) || !reflect.DeepEqual(s.redoStack, redoBefore) {
				t.Fatal("failed rename changed history")
			}
			if _, err := s.Redo(); err != nil {
				t.Fatalf("previous redo no longer works: %v", err)
			}
		})
	}
}

func TestRenameTableDoesNotOverwriteSiblingSource(t *testing.T) {
	s, target, _ := renameFixture(t)
	root := s.current.Project.Root
	writeFile(t, filepath.Join(filepath.Dir(target), "spell.cue"), "// preserve this sibling source")
	before := fileTree(t, root)
	if _, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"}); err == nil {
		t.Fatal("rename overwrote a sibling source")
	}
	assertFileTree(t, root, before)
}

func TestRenameTableReportsRollbackFailureWithoutDeletingMovedAssets(t *testing.T) {
	for _, failure := range []string{"filename rollback", "directory rollback"} {
		t.Run(failure, func(t *testing.T) {
			s, target, _ := renameFixture(t)
			root := s.current.Project.Root
			writeFile(t, filepath.Join(filepath.Dir(target), "sidecar.bin"), "only copy")
			failedWrite := errors.New("project write failed")
			failedRollback := errors.New("rollback move failed")
			ops := defaultTableRenameOps()
			ops.saveProject = func(Project) error { return failedWrite }
			failAt := 3
			sourceName := "spell.cue"
			if failure == "directory rollback" {
				failAt = 4
				sourceName = "effect.cue"
			}
			calls := 0
			ops.rename = func(from, to string) error {
				calls++
				if calls == failAt {
					return failedRollback
				}
				return os.Rename(from, to)
			}
			_, err := s.renameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"}, ops)
			if !errors.Is(err, failedWrite) || !errors.Is(err, failedRollback) || !strings.Contains(err.Error(), "rollback") {
				t.Fatalf("both operation and rollback errors must be reported: %v", err)
			}
			for _, name := range []string{sourceName, "sidecar.bin"} {
				if _, err := os.Stat(filepath.Join(root, "tables", "magic", "spell", name)); err != nil {
					t.Fatalf("failed rollback lost moved file %s: %v", name, err)
				}
			}
			if len(s.undoStack) != 0 || len(s.redoStack) != 0 {
				t.Fatal("failed rollback recorded successful edit")
			}
		})
	}
}

func TestSnapshotRenameFilesRejectsReadErrors(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.cue")
	writeFile(t, path, "source")
	missing := filepath.Join(root, "missing.cue")
	snapshot, err := snapshotRenameFiles(path, path, missing)
	if err != nil || string(snapshot[path]) != "source" || snapshot[missing] != nil || len(snapshot) != 2 {
		t.Fatalf("snapshot = %#v, %v", snapshot, err)
	}
	// Reading a directory fails on all supported platforms, without requiring
	// chmod or depending on whether the tests run as root.
	if _, err := snapshotRenameFiles(root); err == nil {
		t.Fatal("snapshot treated a read error as an absent file")
	}
}

func TestFailedRenameRowLeavesHistoryUnchanged(t *testing.T) {
	s, target, _ := renameFixture(t)
	if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "not valid"}); err == nil {
		t.Fatal("invalid row rename succeeded")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if len(s.undoStack) != 0 || len(s.redoStack) != 1 {
		t.Fatal("failed row rename changed history")
	}
}

func TestRenameTableWithinSameDirectory(t *testing.T) {
	s, target, _ := renameFixture(t)
	root := s.current.Project.Root
	directory := filepath.Join(root, "tables", "magic", "spell")
	if err := os.MkdirAll(filepath.Dir(directory), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Dir(target), directory); err != nil {
		t.Fatal(err)
	}
	target = filepath.Join(directory, filepath.Base(target))
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	before := fileTree(t, root)
	result, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Table.Path != filepath.Join(directory, "spell.cue") || len(result.Workspace.ForeignKeys.Diagnostics) != 0 {
		t.Fatalf("same-directory rename = %#v", result.Table)
	}
	after := fileTree(t, root)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, root, before)
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, root, after)
}
