package workspace

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotHistoryRejectsExternalSourceChanges(t *testing.T) {
	for _, redo := range []bool{false, true} {
		t.Run(fmt.Sprintf("redo=%v", redo), func(t *testing.T) {
			s, target, reference := renameFixture(t)
			if _, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"}); err != nil {
				t.Fatal(err)
			}
			operation := s.Undo
			if redo {
				if _, err := s.Undo(); err != nil {
					t.Fatal(err)
				}
				operation = s.Redo
			}
			committed, err := os.ReadFile(reference)
			if err != nil {
				t.Fatal(err)
			}
			external := append(bytes.Clone(committed), []byte("\n// external work that must survive history\n")...)
			writeFile(t, reference, string(external))
			before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
			if _, err := operation(); err == nil {
				t.Fatal("snapshot history accepted externally modified dependency")
			}
			assertFileTree(t, s.current.Project.Root, before)
			if historyTestState(s) != state {
				t.Fatal("rejected history changed workspace or stacks")
			}
			writeFile(t, reference, string(committed))
			if _, err := operation(); err != nil {
				t.Fatalf("history cannot retry after conflict is removed: %v", err)
			}
		})
	}
}

func TestSnapshotHistoryMissingAndRecreatedSources(t *testing.T) {
	t.Run("deleted dependency", func(t *testing.T) {
		s, target, reference := renameFixture(t)
		if _, err := s.RenameRow(RowRename{Path: target, OldID: "burn", NewID: "flame"}); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(reference); err != nil {
			t.Fatal(err)
		}
		before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
		if _, err := s.Undo(); err == nil {
			t.Fatal("history recreated externally deleted dependency")
		}
		assertFileTree(t, s.current.Project.Root, before)
		if historyTestState(s) != state {
			t.Fatal("conflict consumed history")
		}
	})
	t.Run("empty file at absent path", func(t *testing.T) {
		s, _, _ := renameFixture(t)
		created, err := s.CreateSchemaFile("fresh")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Undo(); err != nil {
			t.Fatal(err)
		}
		writeFile(t, created.Schema.Path, "")
		before, state := fileTree(t, s.current.Project.Root), historyTestState(s)
		if _, err := s.Redo(); err == nil {
			t.Fatal("history overwrote new empty file at absent path")
		}
		assertFileTree(t, s.current.Project.Root, before)
		if historyTestState(s) != state {
			t.Fatal("conflict consumed history")
		}
		if err := os.Remove(created.Schema.Path); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Redo(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSnapshotHistoryMovedSourcesAndUnrelatedChanges(t *testing.T) {
	s, target, _ := renameFixture(t)
	result, err := s.RenameTable(TableRename{Path: target, Namespace: "magic", Name: "spell"})
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(result.Table.Path)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, result.Table.Path, string(committed)+"\n// external source edit\n")
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.Undo(); err == nil {
		t.Fatal("history accepted externally edited moved source")
	}
	assertFileTree(t, s.current.Project.Root, before)
	writeFile(t, result.Table.Path, string(committed))
	// Unrelated files must not block history or be restored from an old snapshot.
	sidecar := filepath.Join(filepath.Dir(result.Table.Path), "new-sidecar.bin")
	writeFile(t, sidecar, "external asset")
	unrelated := filepath.Join(s.current.Project.Root, "schema", "external.cue")
	writeFile(t, unrelated, "package external\n#Value: int\n")
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(filepath.Dir(target), "new-sidecar.bin")); err != nil || string(got) != "external asset" {
		t.Fatalf("sidecar lost: %q %v", got, err)
	}
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(sidecar); err != nil || string(got) != "external asset" {
		t.Fatalf("sidecar lost after redo: %q %v", got, err)
	}
	if got, err := os.ReadFile(unrelated); err != nil || string(got) != "package external\n#Value: int\n" {
		t.Fatalf("unrelated source lost: %q %v", got, err)
	}
}
