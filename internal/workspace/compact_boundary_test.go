package workspace

import (
	"bytes"
	"datapack/internal/document"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Unlike hydrated numeric lifecycle tests, every boundary edit/history below
// must return a genuine compact delta. Stack sizes and byte-exact source restore
// are observed inside the real Service, including crossing file/cell history.
func TestCompactBoundaryHistorySourceAndStacks(t *testing.T) {
	t.Setenv("DATAPACK_FAST_EDIT", "1")
	root := t.TempDir()
	tablePath := filepath.Join(root, "tables", "literal.cue")
	writeFile(t, filepath.Join(root, "literal.dtproj"), `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0755); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile("../../test/compactdelta/literal.cue")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, tablePath, string(original))
	s := NewService(t.TempDir())
	if _, err = s.OpenProject(filepath.Join(root, "literal.dtproj")); err != nil {
		t.Fatal(err)
	}
	if _, err = s.GetTable(tablePath); err != nil {
		t.Fatal(err)
	}
	updates := []CellUpdate{
		{tablePath, "alpha", "n", "9007199254740991"},
		{tablePath, "alpha", "n", "9007199254740993"},
		{tablePath, "beta", "f", "-0.12345678901234567890123456789"},
		{tablePath, "alpha", "f", "1e400"},
		{tablePath, "alpha", "f", "-1e-400"},
		{tablePath, "beta", "nullable", "9007199254740993.0"},
		{tablePath, "beta", "nullable", "null"},
	}
	source := func() []byte {
		b, e := os.ReadFile(tablePath)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	check := func(r CellEditResult, base int) {
		t.Helper()
		if r.Delta == nil || r.Table != nil {
			t.Fatalf("expected compact only: %+v", r)
		}
		d := r.Delta
		if d.BaseRevision != base || d.Revision != base+1 || r.Workspace.Revision != d.Revision || d.ProjectEpoch != s.current.ProjectEpoch {
			t.Fatal("version contract")
		}
		fresh, e := document.LoadTable(tablePath)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(fresh.Rows[d.RowIndex].Values[d.Column], d.Cell) {
			t.Fatal("delta vs persisted source")
		}
		b, e := json.Marshal(r)
		if e != nil {
			t.Fatal(e)
		}
		var wire map[string]json.RawMessage
		if e = json.Unmarshal(b, &wire); e != nil {
			t.Fatal(e)
		}
		if _, ok := wire["table"]; ok {
			t.Fatal("compact includes table")
		}
	}
	states := [][]byte{original}
	for i, u := range updates {
		base := s.revision
		r, e := s.UpdateCellDelta(u)
		if e != nil {
			t.Fatal(e)
		}
		check(r, base)
		states = append(states, source())
		if len(s.undoStack) != i+1 || len(s.redoStack) != 0 || !r.Workspace.CanUndo || r.Workspace.CanRedo {
			t.Fatal("edit stack mismatch")
		}
	}
	// Two hydrated entries above the compact history, then back to literal source.
	beforeFormula := source()
	if _, e := s.SetCellFormula(FormulaSet{Path: tablePath, RowID: "alpha", Column: "spare", Expression: "1 + 2"}); e != nil {
		t.Fatal(e)
	}
	formulaSource := source()
	full, e := s.UpdateCellDelta(CellUpdate{tablePath, "alpha", "n", "9007199254740995"})
	if e != nil || full.Delta != nil || full.Table == nil {
		t.Fatalf("hydrated edit: %v", e)
	}
	for i, want := range [][]byte{formulaSource, beforeFormula} {
		base := s.revision
		r, e := s.UndoDelta()
		if e != nil {
			t.Fatal(e)
		}
		if r.Delta != nil || r.Table != nil || r.Workspace.Revision != base+1 {
			t.Fatal("file history must request refresh")
		}
		if !bytes.Equal(source(), want) {
			t.Fatalf("hydrated undo source %d", i)
		}
	}
	for i := len(updates) - 1; i >= 0; i-- {
		base := s.revision
		r, e := s.UndoDelta()
		if e != nil {
			t.Fatal(e)
		}
		check(r, base)
		if !bytes.Equal(source(), states[i]) {
			t.Fatalf("undo lost source %d", i)
		}
		if len(s.undoStack) != i || len(s.redoStack) != len(updates)-i+2 || r.Workspace.CanUndo != (i > 0) || !r.Workspace.CanRedo {
			t.Fatalf("undo stack mismatch %d", i)
		}
	}
	stableRevision := s.revision
	stableSource := source()
	if _, e = s.UndoDelta(); e == nil {
		t.Fatal("empty undo accepted")
	}
	if s.revision != stableRevision || !bytes.Equal(source(), stableSource) {
		t.Fatal("empty undo mutated state")
	}
	for i := 0; i < len(updates); i++ {
		base := s.revision
		r, e := s.RedoDelta()
		if e != nil {
			t.Fatal(e)
		}
		check(r, base)
		if !bytes.Equal(source(), states[i+1]) {
			t.Fatalf("redo lost source %d", i)
		}
		if len(s.undoStack) != i+1 || len(s.redoStack) != len(updates)-i+1 {
			t.Fatal("redo stack mismatch")
		}
	}
	// A real new compact edit clears only the redo branch, preserving all undo.
	base := s.revision
	r, e := s.UpdateCellDelta(CellUpdate{tablePath, "beta", "n", "9007199254740997"})
	if e != nil {
		t.Fatal(e)
	}
	check(r, base)
	if len(s.undoStack) != len(updates)+1 || len(s.redoStack) != 0 || r.Workspace.CanRedo {
		t.Fatal("redo branch retained after edit")
	}
	stableRevision = s.revision
	stableSource = source()
	if _, e = s.RedoDelta(); e == nil {
		t.Fatal("empty redo accepted")
	}
	if s.revision != stableRevision || !bytes.Equal(source(), stableSource) {
		t.Fatal("empty redo mutated state")
	}
}
