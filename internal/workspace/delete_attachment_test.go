package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestDeleteUndoRestoresAttachedFiles(t *testing.T) {
	s, _, table := renameFixture(t)
	asset := filepath.Join(filepath.Dir(table), "assets", "only-copy.bin")
	writeFile(t, asset, "original attached data")
	if _, err := s.DeleteTable(TableDelete{Path: table}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(table); err != nil {
		t.Fatalf("main source not restored: %v", err)
	}
	data, err := os.ReadFile(asset)
	if err != nil || string(data) != "original attached data" {
		t.Fatalf("Undo restored the main source but lost its attached file: %v", err)
	}
}

// Inspect entry kinds/link text without following links; binary files, empty
// directories and permissions are observable in the same comparison.
func assertDirectory(t *testing.T, projectRoot, path string, want directorySnapshot) {
	t.Helper()
	got, err := snapshotDirectory(projectRoot, path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("directory not restored: got %#v want %#v", got, want)
	}
}
func assertNoDeleteStaging(t *testing.T, root string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, ".datapack-table-history-*"))
	if err != nil || len(paths) != 0 {
		t.Fatalf("staging leaked: %v %v", paths, err)
	}
}
func attachmentFixture(t *testing.T) (*Service, string, directorySnapshot) {
	t.Helper()
	s, _, table := renameFixture(t)
	directory := filepath.Dir(table)
	t.Cleanup(func() {
		_ = filepath.WalkDir(s.current.Project.Root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return os.Chmod(path, 0700)
			}
			return nil
		})
	})
	writeFile(t, filepath.Join(directory, "assets", "only-copy.bin"), "\x00\xffattached data")
	writeFile(t, filepath.Join(directory, "assets", "empty.bin"), "")
	if err := os.MkdirAll(filepath.Join(directory, "empty"), 0750); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(filepath.Join(directory, "assets", "only-copy.bin"), 0640); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Join(directory, "assets"), 0550); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(directory, 0750); err != nil {
			t.Fatal(err)
		}
	}
	before, err := snapshotDirectory(s.current.Project.Root, directory)
	if err != nil {
		t.Fatal(err)
	}
	return s, table, before
}
func TestDeleteTableRepeatedUndoRedoRestoresAttachmentsAndPermissions(t *testing.T) {
	s, table, before := attachmentFixture(t)
	root, directory := s.current.Project.Root, filepath.Dir(table)
	config, err := os.ReadFile(s.current.Project.Path)
	if err != nil {
		t.Fatal(err)
	}
	for cycle := 0; cycle < 3; cycle++ {
		if cycle == 0 {
			if _, err := s.DeleteTable(TableDelete{Path: table}); err != nil {
				t.Fatal(err)
			}
		} else if _, err := s.Redo(); err != nil {
			t.Fatal(err)
		}
		assertDirectory(t, root, directory, nil)
		assertNoDeleteStaging(t, root)
		if _, exists := s.current.Project.Config.Export.Tables["rpg.item"]; exists {
			t.Fatal("export selection retained deleted table")
		}
		state := historyTestState(s)
		if _, err := s.DeleteTable(TableDelete{Path: table}); err == nil {
			t.Fatal("deleted table accepted twice")
		}
		if historyTestState(s) != state {
			t.Fatal("duplicate deletion changed history")
		}
		if _, err := s.Undo(); err != nil {
			t.Fatal(err)
		}
		assertDirectory(t, root, directory, before)
		assertNoDeleteStaging(t, root)
		got, err := os.ReadFile(s.current.Project.Path)
		if err != nil || !bytes.Equal(got, config) {
			t.Fatalf("original config not restored: %v", err)
		}
	}
}
func TestDeleteTableLinksNeverTouchTheirTargets(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires Windows privilege")
	}
	s, _, table := renameFixture(t)
	root, directory := s.current.Project.Root, filepath.Dir(table)
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "only-copy.bin"), "outside secret-like fixture")
	for name, target := range map[string]string{"external": outside, "relative": "item.cue", "dangling": "missing.bin"} {
		if err := os.Symlink(target, filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
	}
	before, err := snapshotDirectory(root, directory)
	if err != nil {
		t.Fatal(err)
	}
	outsideBefore := fileTree(t, outside)
	for i := 0; i < 2; i++ {
		if i == 0 {
			_, err = s.DeleteTable(TableDelete{Path: table})
		} else {
			_, err = s.Redo()
		}
		if err != nil {
			t.Fatal(err)
		}
		assertDirectory(t, root, directory, nil)
		assertFileTree(t, outside, outsideBefore)
		if _, err = s.Undo(); err != nil {
			t.Fatal(err)
		}
		assertDirectory(t, root, directory, before)
		assertFileTree(t, outside, outsideBefore)
	}
}
func TestDeleteTableFailuresCompensateFilesConfigAndBothHistoryStacks(t *testing.T) {
	for _, failure := range []string{"move", "save-partial", "refresh"} {
		t.Run(failure, func(t *testing.T) {
			s, table, tree := attachmentFixture(t)
			// Ensure there is a redo entry that failed deletion must retain.
			if _, err := s.UpdateCell(CellUpdate{Path: table, RowID: "sword", Column: "effectID", Expression: `"burn"`}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			project := s.current.Project
			original, err := os.ReadFile(project.Path)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected " + failure)
			ops := tableDeleteOps{rename: os.Rename, saveProject: SaveProject}
			switch failure {
			case "move":
				ops.rename = func(string, string) error { return injected }
			case "save-partial":
				ops.saveProject = func(project Project) error {
					if err := os.WriteFile(project.Path, []byte("{"), 0600); err != nil {
						t.Fatal(err)
					}
					return injected
				}
			case "refresh":
				// Actual refresh error, with export save still running normally.
				if err := os.Rename(filepath.Join(project.Root, "schema"), filepath.Join(project.Root, "schema-away")); err != nil {
					t.Fatal(err)
				}
			}
			state, before := historyTestState(s), fileTree(t, project.Root)
			_, err = s.deleteTable(TableDelete{Path: table}, ops)
			if err == nil || failure != "refresh" && !errors.Is(err, injected) {
				t.Fatalf("wrong failure: %v", err)
			}
			assertFileTree(t, project.Root, before)
			assertDirectory(t, project.Root, filepath.Dir(table), tree)
			assertNoDeleteStaging(t, project.Root)
			if historyTestState(s) != state {
				t.Fatal("failed delete changed workspace/config map/history")
			}
			got, err := os.ReadFile(project.Path)
			if err != nil || !bytes.Equal(got, original) {
				t.Fatalf("failed delete did not restore config: %v", err)
			}
			if failure == "refresh" {
				if err := os.Rename(filepath.Join(project.Root, "schema-away"), filepath.Join(project.Root, "schema")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.DeleteTable(TableDelete{Path: table}); err != nil {
				t.Fatalf("cannot retry failed delete: %v", err)
			}
			if len(s.redoStack) != 0 {
				t.Fatal("successful retry retained stale redo")
			}
		})
	}
}
func TestDeleteTableHistoryRefreshFailureCanRetry(t *testing.T) {
	for _, redo := range []bool{false, true} {
		t.Run(fmt.Sprint("redo=", redo), func(t *testing.T) {
			s, table, _ := attachmentFixture(t)
			project := s.current.Project
			if _, err := s.DeleteTable(TableDelete{Path: table}); err != nil {
				t.Fatal(err)
			}
			if redo {
				if _, err := s.Undo(); err != nil {
					t.Fatal(err)
				}
			}
			// The config is part of this history entry; damage a separate schema root.
			if err := os.Rename(filepath.Join(project.Root, "schema"), filepath.Join(project.Root, "schema-away")); err != nil {
				t.Fatal(err)
			}
			before, state := fileTree(t, project.Root), historyTestState(s)
			tree, err := snapshotDirectory(project.Root, filepath.Dir(table))
			if err != nil {
				t.Fatal(err)
			}
			operation := s.Undo
			if redo {
				operation = s.Redo
			}
			if _, err := operation(); err == nil {
				t.Fatal("history succeeded with invalid schema root")
			}
			assertFileTree(t, project.Root, before)
			assertDirectory(t, project.Root, filepath.Dir(table), tree)
			assertNoDeleteStaging(t, project.Root)
			if historyTestState(s) != state {
				t.Fatal("failed history consumed stacks or changed view")
			}
			if err := os.Rename(filepath.Join(project.Root, "schema-away"), filepath.Join(project.Root, "schema")); err != nil {
				t.Fatal(err)
			}
			if _, err := operation(); err != nil {
				t.Fatalf("cannot retry history: %v", err)
			}
		})
	}
}
func TestDeleteTableHistoryConflictsPreserveUserFiles(t *testing.T) {
	for _, kind := range []string{"undo-directory", "undo-file", "undo-symlink", "redo-extra-attachment", "redo-changed-attachment"} {
		t.Run(kind, func(t *testing.T) {
			s, table, _ := attachmentFixture(t)
			root, directory := s.current.Project.Root, filepath.Dir(table)
			if _, err := s.DeleteTable(TableDelete{Path: table}); err != nil {
				t.Fatal(err)
			}
			operation := s.Undo
			switch kind {
			case "undo-directory":
				writeFile(t, filepath.Join(directory, "user.bin"), "unrelated new data")
			case "undo-file":
				writeFile(t, directory, "unrelated new file")
			case "undo-symlink":
				if runtime.GOOS == "windows" {
					t.Skip("symlink creation requires Windows privilege")
				}
				if err := os.Symlink(t.TempDir(), directory); err != nil {
					t.Fatal(err)
				}
			default:
				if _, err := s.Undo(); err != nil {
					t.Fatal(err)
				}
				if kind == "redo-extra-attachment" {
					writeFile(t, filepath.Join(directory, "user.bin"), "new data")
				} else {
					writeFile(t, filepath.Join(directory, "item.cue"), "user edit")
				}
				operation = s.Redo
			}
			state := historyTestState(s)
			before, err := snapshotDirectory(filepath.Dir(root), root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := operation(); err == nil {
				t.Fatal("history destroyed conflicting data")
			}
			if historyTestState(s) != state {
				t.Fatal("conflict consumed history")
			}
			assertDirectory(t, filepath.Dir(root), root, before)
			assertNoDeleteStaging(t, root)
		})
	}
}
func TestDeleteTableRejectsSymlinkAncestorsAndSource(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires Windows privilege")
	}
	for _, kind := range []string{"directory", "ancestor", "source"} {
		t.Run(kind, func(t *testing.T) {
			s, _, table := renameFixture(t)
			root := s.current.Project.Root
			from := filepath.Dir(table)
			if kind == "ancestor" {
				from = filepath.Dir(from)
			}
			if kind == "source" {
				from = table
			}
			to := filepath.Join(t.TempDir(), "outside")
			if err := os.Rename(from, to); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(to, from); err != nil {
				t.Fatal(err)
			}
			outside := fileTree(t, filepath.Dir(to))
			state := historyTestState(s)
			if _, err := s.DeleteTable(TableDelete{Path: table}); err == nil {
				t.Fatal("delete accepted symlink source path")
			}
			assertFileTree(t, filepath.Dir(to), outside)
			if historyTestState(s) != state {
				t.Fatal("unsafe delete changed history")
			}
			assertNoDeleteStaging(t, root)
		})
	}
}
func TestDirectoryInstallFailureAndRollbackFailureReported(t *testing.T) {
	s, table, before := attachmentFixture(t)
	root, directory := s.current.Project.Root, filepath.Dir(table)
	if _, err := s.DeleteTable(TableDelete{Path: table}); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("install failed")
	h := &directoryHistory{Path: directory, Restore: before}
	if _, err := changeDirectory(s.current.Project, h, func(string, string) error { return injected }); !errors.Is(err, injected) {
		t.Fatalf("install error lost: %v", err)
	}
	assertDirectory(t, root, directory, nil)
	assertNoDeleteStaging(t, root)
	// A rollback rename failure retains the original tree in an identified
	// recovery directory and reports it, rather than deleting the only copy.
	h = &directoryHistory{Path: directory, Restore: before}
	transaction, err := changeDirectory(s.current.Project, h, os.Rename)
	if err != nil {
		t.Fatal(err)
	}
	transaction.rename = func(string, string) error { return injected }
	if err := transaction.rollback(); !errors.Is(err, injected) || !strings.Contains(err.Error(), transaction.stage) {
		t.Fatalf("rollback recovery location/error lost: %v", err)
	}
	assertDirectory(t, root, directory, before)
	transaction.rename = os.Rename
	if err := transaction.rollback(); err != nil {
		t.Fatal(err)
	}
	assertDirectory(t, root, directory, nil)
	assertNoDeleteStaging(t, root)
}

func TestDeleteTableSnapshotReadFailureLeavesEverythingUntouched(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission semantics")
	}
	s, _, table := renameFixture(t)
	asset := filepath.Join(filepath.Dir(table), "unreadable.bin")
	writeFile(t, asset, "must not remove")
	if err := os.Chmod(asset, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(asset, 0600)
	if _, err := os.ReadFile(asset); err == nil {
		t.Skip("test user can read mode 0000")
	}
	state := historyTestState(s)
	config, err := os.ReadFile(s.current.Project.Path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteTable(TableDelete{Path: table}); err == nil {
		t.Fatal("delete discarded an unreadable attachment")
	}
	if historyTestState(s) != state {
		t.Fatal("failed snapshot changed history")
	}
	got, err := os.ReadFile(s.current.Project.Path)
	if err != nil || !bytes.Equal(got, config) {
		t.Fatalf("config changed: %v", err)
	}
	if _, err := os.Stat(table); err != nil {
		t.Fatal(err)
	}
	assertNoDeleteStaging(t, s.current.Project.Root)
}
func TestDeleteTableCannotOwnSourceRootOrAnotherTable(t *testing.T) {
	for _, kind := range []string{"source-root", "nested-table"} {
		t.Run(kind, func(t *testing.T) {
			s, _, table := renameFixture(t)
			if kind == "source-root" {
				s.current.Project.Config.TableRoots = append(s.current.Project.Config.TableRoots, filepath.Join("tables", "rpg", "item"))
			} else {
				source, err := os.ReadFile(table)
				if err != nil {
					t.Fatal(err)
				}
				nested := filepath.Join(filepath.Dir(table), "nested", "nested.cue")
				writeFile(t, nested, strings.ReplaceAll(string(source), `name: "item"`, `name: "nested"`))
				if _, err := s.OpenProject(s.current.Project.Path); err != nil {
					t.Fatal(err)
				}
			}
			root := s.current.Project.Root
			before, state := fileTree(t, root), historyTestState(s)
			if _, err := s.DeleteTable(TableDelete{Path: table}); err == nil {
				t.Fatal("delete accepted ownership of another source")
			}
			assertFileTree(t, root, before)
			if historyTestState(s) != state {
				t.Fatal("unsafe deletion changed history")
			}
			assertNoDeleteStaging(t, root)
		})
	}
}
