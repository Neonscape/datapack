package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func historyTestState(s *Service) string {
	return fmt.Sprintf("%#v|%#v|%#v|%d|%d|%#v", s.undoStack, s.redoStack, *s.current, s.revision, s.compiledRevision, s.compileState)
}

func TestHistoryRefreshFailurePreservesFilesAndBothStacks(t *testing.T) {
	for _, redo := range []bool{false, true} {
		t.Run(fmt.Sprint("redo=", redo), func(t *testing.T) {
			s, target, _ := renameFixture(t)
			if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
				t.Fatal(err)
			}
			if redo {
				if _, err := s.Undo(); err != nil {
					t.Fatal(err)
				}
			}
			projectPath := s.current.Project.Path
			valid, err := os.ReadFile(projectPath)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, projectPath, "{")
			before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
			operation := s.Undo
			if redo {
				operation = s.Redo
			}
			if _, err := operation(); err == nil {
				t.Fatal("history succeeded with invalid project")
			}
			assertFileTree(t, s.current.Project.Root, before)
			if historyTestState(s) != state {
				t.Fatal("failed refresh consumed history or changed workspace state")
			}
			writeFile(t, projectPath, string(valid))
			if _, err := operation(); err != nil {
				t.Fatalf("cannot retry history after repairing project: %v", err)
			}
		})
	}
}

func TestHistoryReadFailureDoesNotMistakeDirectoryForMissingFile(t *testing.T) {
	s, target, _ := renameFixture(t)
	if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
		t.Fatal(err)
	}
	hot, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "only-copy.bin"), "preserve this source")
	before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
	if _, err := s.Undo(); err == nil {
		t.Fatal("history accepted unreadable current file")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if historyTestState(s) != state {
		t.Fatal("failed snapshot consumed history")
	}
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}
	writeFile(t, target, string(hot))
	if _, err := s.Undo(); err != nil {
		t.Fatalf("cannot retry after repairing path: %v", err)
	}
}

func TestHistoryDoesNotTreatDanglingSymlinkAsMissingFile(t *testing.T) {
	s, target, _ := renameFixture(t)
	if _, err := s.UpdateCell(CellUpdate{Path: target, RowID: "burn", Column: "name", Expression: `"Hot"`}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-source", target); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks unavailable")
		}
		t.Fatal(err)
	}
	state := historyTestState(s)
	if _, err := s.Undo(); err == nil {
		t.Fatal("history overwrote unreadable existing link")
	}
	if link, err := os.Readlink(target); err != nil || link != "missing-source" {
		t.Fatalf("dangling link changed: %q %v", link, err)
	}
	if historyTestState(s) != state {
		t.Fatal("failed snapshot consumed history")
	}
}

func TestTableRenameHistoryMovesAllSidecars(t *testing.T) {
	s, target, _ := renameFixture(t)
	oldDirectory := filepath.Dir(target)
	asset := filepath.Join(oldDirectory, "assets", "binary.bin")
	writeFile(t, asset, "\x00\xff\x01asset")
	if err := os.Chmod(asset, 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(oldDirectory, "empty"), 0755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(oldDirectory, "relative-link")
	linkErr := os.Symlink(filepath.Join("assets", "binary.bin"), link)
	if linkErr != nil && runtime.GOOS != "windows" {
		t.Fatal(linkErr)
	}
	before := fileTree(t, s.current.Project.Root)
	renamed, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"})
	if err != nil {
		t.Fatal(err)
	}
	newDirectory := filepath.Dir(renamed.Table.Path)
	after := fileTree(t, s.current.Project.Root)
	for _, action := range []struct {
		run   func() (HistoryResult, error)
		dir   string
		files map[string]string
	}{
		{s.Undo, oldDirectory, before}, {s.Redo, newDirectory, after}, {s.Undo, oldDirectory, before},
	} {
		if _, err := action.run(); err != nil {
			t.Fatal(err)
		}
		assertFileTree(t, s.current.Project.Root, action.files)
		if info, err := os.Stat(filepath.Join(action.dir, "empty")); err != nil || !info.IsDir() {
			t.Fatalf("empty directory lost: %v", err)
		}
		if info, err := os.Stat(filepath.Join(action.dir, "assets", "binary.bin")); err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != 0640 {
			t.Fatalf("sidecar permissions changed: %v", err)
		}
		if value, err := os.Readlink(filepath.Join(action.dir, "relative-link")); linkErr == nil && (err != nil || value != filepath.Join("assets", "binary.bin")) {
			t.Fatalf("sidecar symlink changed: %q %v", value, err)
		}
	}
}

func TestTableRenameHistoryRefreshFailureRollsBackMovesAndBytes(t *testing.T) {
	s, target, _ := renameFixture(t)
	writeFile(t, filepath.Join(filepath.Dir(target), "sidecar.bin"), "asset")
	if _, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"}); err != nil {
		t.Fatal(err)
	}
	schemaRoot := filepath.Join(s.current.Project.Root, "schema")
	if err := os.Rename(schemaRoot, schemaRoot+".away"); err != nil {
		t.Fatal(err)
	}
	before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
	if _, err := s.Undo(); err == nil {
		t.Fatal("history succeeded without configured schema root")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if historyTestState(s) != state {
		t.Fatal("failed table history changed workspace or stacks")
	}
	if err := os.Rename(schemaRoot+".away", schemaRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatalf("cannot retry table history: %v", err)
	}
}

func TestSnapshotFailureCompensatesAllEarlierWritesAndRemovals(t *testing.T) {
	for _, failure := range []string{"write", "remove"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			a, b, c := filepath.Join(root, "a"), filepath.Join(root, "b"), filepath.Join(root, "c")
			writeFile(t, a, "original")
			writeFile(t, c, "")
			if err := os.Chmod(a, 0640); err != nil {
				t.Fatal(err)
			}
			before := fileTree(t, root)
			injected := errors.New("injected " + failure + " failure")
			ops := snapshotRestoreOps{write: writeFileAtomic, remove: os.Remove, mkdirAll: os.MkdirAll}
			if failure == "write" {
				ops.write = func(path string, content []byte) error {
					if path == b {
						return injected
					}
					return writeFileAtomic(path, content)
				}
			}
			if failure == "remove" {
				ops.remove = func(path string) error {
					if path == c {
						return injected
					}
					return os.Remove(path)
				}
			}
			err := restoreSnapshotWithOps(map[string][]byte{a: []byte("changed"), b: []byte("created"), c: nil}, ops)
			if !errors.Is(err, injected) {
				t.Fatalf("missing original error: %v", err)
			}
			assertFileTree(t, root, before)
			if info, err := os.Stat(a); err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != 0640 {
				t.Fatalf("rollback changed file permissions: %v", err)
			}
		})
	}
}

func TestSnapshotRollbackFailureIsReportedWithOriginalError(t *testing.T) {
	root := t.TempDir()
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	writeFile(t, a, "original")
	commitErr, rollbackErr := errors.New("commit failed"), errors.New("rollback failed")
	ops := snapshotRestoreOps{remove: os.Remove, mkdirAll: os.MkdirAll, write: func(path string, content []byte) error {
		if path == b {
			return commitErr
		}
		if string(content) == "original" {
			return rollbackErr
		}
		return writeFileAtomic(path, content)
	}}
	err := restoreSnapshotWithOps(map[string][]byte{a: []byte("changed"), b: []byte("new")}, ops)
	if !errors.Is(err, commitErr) || !errors.Is(err, rollbackErr) {
		t.Fatalf("incomplete failure report: %v", err)
	}
}

func TestTableRenameHistoryCollisionPreservesFilesAndCanRetry(t *testing.T) {
	s, target, _ := renameFixture(t)
	oldDirectory := filepath.Dir(target)
	writeFile(t, filepath.Join(oldDirectory, "sidecar.bin"), "asset")
	renamed, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"})
	if err != nil {
		t.Fatal(err)
	}
	newDirectory := filepath.Dir(renamed.Table.Path)
	writeFile(t, filepath.Join(oldDirectory, "new-user-file.bin"), "must not overwrite")
	before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
	if _, err := s.Undo(); err == nil {
		t.Fatal("Undo overwrote an existing destination directory")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if historyTestState(s) != state {
		t.Fatal("failed Undo consumed history")
	}
	if err := os.RemoveAll(oldDirectory); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	// The directory move succeeds, then the source filename move collides
	// with this new sidecar. Both moves must be compensated without data loss.
	writeFile(t, filepath.Join(oldDirectory, "spell.cue"), "user source")
	before, state = fileTree(t, s.current.Project.Root), historyTestState(s)
	if _, err := s.Redo(); err == nil {
		t.Fatal("Redo overwrote a filename collision")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if historyTestState(s) != state {
		t.Fatal("failed Redo consumed history")
	}
	if _, err := os.Stat(newDirectory); !os.IsNotExist(err) {
		t.Fatalf("failed Redo left moved directory: %v", err)
	}
	if err := os.Remove(filepath.Join(oldDirectory, "spell.cue")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Redo(); err != nil {
		t.Fatalf("cannot retry after removing collision: %v", err)
	}
}
