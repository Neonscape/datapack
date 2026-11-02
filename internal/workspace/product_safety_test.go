package workspace

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestLiteralRefreshFailureRollsBackWithoutPublishing(t *testing.T) {
	s, path, original := literalProject(t)
	before := sourceEditState(s)
	s.literalOps = &literalWriteOps{refresh: func() (string, error) { return "", errors.New("injected refresh") }}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "99"}); err == nil {
		t.Fatal("failure hidden")
	}
	actual, _ := os.ReadFile(path)
	if !bytes.Equal(original, actual) || sourceEditState(s) != before {
		t.Fatal("failed refresh changed bytes/revision/history")
	}
}
func TestLiteralRollbackFailureLocksAuthorityUntilSuccessfulReopen(t *testing.T) {
	s, path, _ := literalProject(t)
	project := s.current.Project.Path
	s.literalOps = &literalWriteOps{refresh: func() (string, error) { return "", errors.New("refresh") }, restore: func(string, []byte, os.FileMode) error { return errors.New("rollback") }}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "99"}); err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("wrong result: %v", err)
	}
	actual, _ := os.ReadFile(path)
	if !bytes.Contains(actual, []byte("price: 99")) {
		t.Fatal("fixture must expose already committed bytes")
	}
	if !s.current.SourceUncertain || len(s.undoStack) != 0 {
		t.Fatal("uncertain source published as success")
	}
	if _, err := s.Undo(); err == nil {
		t.Fatal("undo accepted uncertain source")
	}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "100"}); err == nil {
		t.Fatal("edit accepted uncertain source")
	}
	if _, err := s.Compile(); err == nil {
		t.Fatal("compile accepted uncertain source")
	}
	if _, err := s.Export(); err == nil {
		t.Fatal("export accepted uncertain source")
	}
	if _, err := s.OpenProject(project + ".missing"); err == nil || !s.sourceUncertain {
		t.Fatal("failed reopen unlocked source")
	}
	s.literalOps = nil
	if _, err := s.OpenProject(project); err != nil {
		t.Fatal(err)
	}
	if s.sourceUncertain {
		t.Fatal("successful reopen did not recover authority")
	}
}
func TestManualOpenAndStaleDiagnosticsHaveExplicitRevision(t *testing.T) {
	s, path, _ := literalProject(t)
	if !s.current.Compile.Pending || s.current.Compile.CompiledAt != "" || s.current.Compile.InputRevision != -1 {
		t.Fatal("manual open claimed validation")
	}
	first, err := s.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if first.Workspace.Compile.Pending || len(first.Workspace.Compile.Diagnostics) > 0 {
		t.Fatal("initial compile")
	}
	// An invalid value is persisted in manual mode; only explicit compile/export validates it.
	invalid, err := s.UpdateCell(CellUpdate{path, "first", "price", `"bad"`})
	if err != nil {
		t.Fatal(err)
	}
	if !invalid.Workspace.Compile.Pending || invalid.Workspace.Compile.InputRevision >= invalid.Workspace.Revision {
		t.Fatal("stale success not marked")
	}
	failed, err := s.Compile()
	if err != nil || len(failed.Workspace.Compile.Diagnostics) == 0 {
		t.Fatalf("invalid compile: %v", err)
	}
	repaired, err := s.UpdateCell(CellUpdate{path, "first", "price", "2"})
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.Workspace.Compile.Pending || len(repaired.Workspace.Compile.Diagnostics) == 0 || repaired.Workspace.Compile.InputRevision >= repaired.Workspace.Revision {
		t.Fatal("old errors lost their revision")
	}
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	if s.current.Compile.Pending || len(s.current.Compile.Diagnostics) > 0 {
		t.Fatal("export did not validate current input")
	}
}
func TestOperationProgressReadableAndConflictingEditsRejected(t *testing.T) {
	s, path, _ := literalProject(t)
	finish, err := s.beginOperation("compile")
	if err != nil {
		t.Fatal(err)
	}
	s.operationProgress("validate", 1, 3)
	state := s.GetOperationState()
	if !state.Running || state.Completed != 1 || state.Total != 3 {
		t.Fatal(state)
	}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "2"}); err == nil {
		t.Fatal("conflicting edit accepted")
	}
	if _, err := s.Compile(); err == nil {
		t.Fatal("duplicate operation accepted")
	}
	finish(errors.New("injected"))
	if state = s.GetOperationState(); state.Running || state.Phase != "failed" {
		t.Fatal(state)
	}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "2"}); err != nil {
		t.Fatal("failure did not unlock", err)
	}
}

func TestHistoryRefreshFailureDoesNotConsumeUndoOrRedo(t *testing.T) {
	for _, undo := range []bool{true, false} {
		t.Run(map[bool]string{true: "undo", false: "redo"}[undo], func(t *testing.T) {
			s, path, _ := literalProject(t)
			if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "2"}); err != nil {
				t.Fatal(err)
			}
			if !undo {
				if _, err := s.Undo(); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := os.ReadFile(path)
			state := sourceEditState(s)
			s.literalOps = &literalWriteOps{refresh: func() (string, error) { return "", errors.New("history refresh") }}
			var err error
			if undo {
				_, err = s.Undo()
			} else {
				_, err = s.Redo()
			}
			if err == nil {
				t.Fatal("history failure hidden")
			}
			after, _ := os.ReadFile(path)
			if !bytes.Equal(before, after) || sourceEditState(s) != state {
				t.Fatal("failed history changed source/revision/stacks")
			}
		})
	}
}
