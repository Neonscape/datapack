package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRenameAndDeleteRefreshFailurePreservesSourcesAndHistory(t *testing.T) {
	for _, operation := range []string{"row rename", "public rename", "local rename", "public delete", "local delete"} {
		t.Run(operation, func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			data, _ := os.ReadFile(table)
			writeFile(t, table, strings.ReplaceAll(string(data), "Existing", "LocalOnly"))
			// A config changed on disk since the last view must not publish a
			// successful source edit when refreshing the view fails.
			writeFile(t, s.current.Project.Path, "invalid project config")
			writeFile(t, filepath.Join(filepath.Dir(table), "attachment.bin"), "binary\x00attachment")
			beforeModes := map[string]os.FileMode{}
			filepath.WalkDir(s.current.Project.Root, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					t.Fatal(err)
				}
				info, err := entry.Info()
				if err != nil {
					t.Fatal(err)
				}
				beforeModes[path] = info.Mode()
				return nil
			})
			before := fileTree(t, s.current.Project.Root)
			state := sourceEditState(s)
			var err error
			switch operation {
			case "row rename":
				_, err = s.RenameRow(RowRename{Path: table, OldID: "one", NewID: "renamed"})
			case "public rename":
				_, err = s.RenameSchemaDefinition(SchemaDefinitionRename{Path: rules, OldName: "Existing", NewName: "Renamed"})
			case "local rename":
				_, err = s.RenameSchemaDefinition(SchemaDefinitionRename{Path: table, OldName: "LocalOnly", NewName: "Renamed"})
			case "public delete":
				_, err = s.DeleteSchemaDefinition(SchemaDefinitionDelete{Path: rules, Name: "Existing"})
			case "local delete":
				_, err = s.DeleteSchemaDefinition(SchemaDefinitionDelete{Path: table, Name: "LocalOnly"})
			}
			if err == nil {
				t.Fatal("refresh unexpectedly succeeded")
			}
			assertFileTree(t, s.current.Project.Root, before)
			afterModes := map[string]os.FileMode{}
			filepath.WalkDir(s.current.Project.Root, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					t.Fatal(err)
				}
				info, err := entry.Info()
				if err != nil {
					t.Fatal(err)
				}
				afterModes[path] = info.Mode()
				return nil
			})
			if !reflect.DeepEqual(beforeModes, afterModes) {
				t.Fatal("failed command changed file/directory permissions")
			}
			if sourceEditState(s) != state {
				t.Fatal("failed operation changed history/revision/compile state")
			}
		})
	}
}

func TestCreateSchemaRefreshFailureRemovesOnlyItsNewDirectories(t *testing.T) {
	s, _, _ := sourceEditFixture(t)
	writeFile(t, s.current.Project.Path, "invalid project config")
	before := fileTree(t, s.current.Project.Root)
	state := sourceEditState(s)
	if _, err := s.CreateSchemaFile("fresh"); err == nil {
		t.Fatal("refresh unexpectedly succeeded")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if sourceEditState(s) != state {
		t.Fatal("failed create changed state")
	}
	if _, err := os.Stat(s.current.Project.Root + "/schema/fresh"); !os.IsNotExist(err) {
		t.Fatalf("left created directory: %v", err)
	}
}

func TestSnapshotSourcesDistinguishesEmptyMissingAndUnreadable(t *testing.T) {
	root := t.TempDir()
	empty, missing := root+"/empty.cue", root+"/missing.cue"
	writeFile(t, empty, "")
	snapshot, err := snapshotRenameFiles(empty, missing)
	if err != nil || snapshot[empty] == nil || len(snapshot[empty]) != 0 || snapshot[missing] != nil {
		t.Fatalf("empty/missing snapshot: %#v %v", snapshot, err)
	}
	if _, err := snapshotRenameFiles(root); err == nil {
		t.Fatal("unreadable source directory accepted")
	}
	link := root + "/dangling.cue"
	if err := os.Symlink(missing, link); err == nil {
		if _, err := snapshotRenameFiles(link); err == nil {
			t.Fatal("existing unreadable symlink treated as missing")
		}
	}
}

func TestUnseenSchemaRenameIsIncludedInHistoryAndRefreshRollback(t *testing.T) {
	for _, failRefresh := range []bool{false, true} {
		t.Run(map[bool]string{true: "failed refresh", false: "history"}[failRefresh], func(t *testing.T) {
			s, _, rules := sourceEditFixture(t)
			target := filepath.Join(filepath.Dir(rules), "unseen.cue")
			source := `package rules
#Fresh: string
_schema: {definitions: {Fresh: {kind: "type",type: "string"}}}`
			writeFile(t, target, source)
			if failRefresh {
				writeFile(t, s.current.Project.Path, "invalid project")
			}
			before := fileTree(t, s.current.Project.Root)
			state := sourceEditState(s)
			_, err := s.RenameSchemaDefinition(SchemaDefinitionRename{Path: target, OldName: "Fresh", NewName: "NewName"})
			if failRefresh {
				if err == nil {
					t.Fatal("refresh succeeded")
				}
				assertFileTree(t, s.current.Project.Root, before)
				if sourceEditState(s) != state {
					t.Fatal("failed rename changed state")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			renamed, _ := os.ReadFile(target)
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			undone, _ := os.ReadFile(target)
			if string(undone) != source {
				t.Fatalf("unseen target missing from undo: %s", undone)
			}
			if _, err := s.Redo(); err != nil {
				t.Fatal(err)
			}
			redone, _ := os.ReadFile(target)
			if string(redone) != string(renamed) {
				t.Fatal("redo differs")
			}
		})
	}
}

func TestSourceCommandsKeepSpecialModeAndRejectReadableSymlinks(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	want := os.FileMode(0640) | os.ModeSetgid | os.ModeSticky
	if err := os.Chmod(table, want); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(table)
	if err != nil {
		t.Fatal(err)
	}
	if sourcePermissionMode(info.Mode()) == want {
		state := sourceEditState(s)
		before, _ := os.ReadFile(table)
		if _, err := s.RenameRow(RowRename{Path: table, OldID: "one", NewID: "invalid id"}); err == nil {
			t.Fatal("invalid row accepted")
		}
		after, _ := os.ReadFile(table)
		info, _ = os.Stat(table)
		if string(before) != string(after) || sourcePermissionMode(info.Mode()) != want || sourceEditState(s) != state {
			t.Fatal("rejected command changed bytes/mode/state")
		}
		if _, err := s.RenameRow(RowRename{Path: table, OldID: "one", NewID: "renamed"}); err != nil {
			t.Fatal(err)
		}
		info, _ = os.Stat(table)
		if sourcePermissionMode(info.Mode()) != want {
			t.Fatal("successful rename lost special mode")
		}
	}
	original, _ := os.ReadFile(table)
	backup := table + ".original"
	if err := os.WriteFile(backup, original, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(table); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(backup, table); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	state := sourceEditState(s)
	if _, err := snapshotRenameFiles(table); err == nil {
		t.Fatal("readable symlink snapshot accepted")
	}
	if _, err := s.RenameRow(RowRename{Path: table, OldID: "renamed", NewID: "another"}); err == nil {
		t.Fatal("source symlink accepted")
	}
	target, err := os.Readlink(table)
	if err != nil || target != backup {
		t.Fatalf("link changed: %s %v", target, err)
	}
	after, _ := os.ReadFile(backup)
	if string(after) != string(original) || sourceEditState(s) != state {
		t.Fatal("symlink rejection changed source/history")
	}
}

func TestWorkspaceRefreshCompensationFailureRetainsOriginalBytes(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	before, _ := os.ReadFile(table)
	undo, redo := len(s.undoStack), len(s.redoStack)
	_, _, err := editSourcesWithOps(s, []string{table}, func() (bool, error) { return true, os.WriteFile(table, []byte("changed"), 0644) }, sourceCommandOps{
		refresh: func() (View, error) { return View{}, fmt.Errorf("injected refresh failure") },
		restore: func(path string, data []byte) error {
			return writeFileAtomicWithRename(path, data, func(string, string) error { return fmt.Errorf("injected rollback rename failure") })
		},
	})
	if err == nil || !strings.Contains(err.Error(), "injected refresh failure") || !strings.Contains(err.Error(), "recovery bytes retained at") {
		t.Fatalf("lost compensation error: %v", err)
	}
	_, tail, found := strings.Cut(err.Error(), `recovery bytes retained at "`)
	recovery, _, closed := strings.Cut(tail, `"`)
	if !found || !closed {
		t.Fatalf("missing recovery location: %v", err)
	}
	data, readErr := os.ReadFile(recovery)
	if readErr != nil || string(data) != string(before) {
		t.Fatalf("original recovery: %s %v", data, readErr)
	}
	if !s.sourceUncertain || !s.current.SourceUncertain || len(s.undoStack) != undo || len(s.redoStack) != redo {
		t.Fatal("rollback failure must mark authority uncertain without consuming history")
	}
}

func TestSingleSourceEditRejectsNonRegularBeforeReading(t *testing.T) {
	s, _, _ := sourceEditFixture(t)
	directory := t.TempDir()
	called := false
	_, _, err := editSource(s, directory, func() (bool, error) { called = true; return true, nil })
	if called || err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("nonregular edit: called=%v err=%v", called, err)
	}
}
