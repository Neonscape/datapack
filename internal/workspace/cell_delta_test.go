package workspace

import (
	"bytes"
	"datapack/internal/document"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestCommittedCellDeltaAndHistoryMatchFullDocument(t *testing.T) {
	s, path, _ := literalProject(t)
	updates := []CellUpdate{{path, "first", "price", "123456"}, {path, "middle", "enabled", "true"}, {path, "last", "category", `"b"`}}
	for _, update := range updates {
		before, _ := os.ReadFile(path)
		base := s.revision
		r, err := s.UpdateCellDelta(update)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(r)
		if r.Delta == nil || r.Table != nil || len(encoded) > 4096 {
			t.Fatalf("not compact: %d %v", len(encoded), r.Delta)
		}
		d := r.Delta
		if d.BaseRevision != base || d.Revision != base+1 || d.ProjectEpoch != s.current.ProjectEpoch || r.Workspace.Revision != d.Revision {
			t.Fatal("wrong authority version")
		}
		raw, _ := os.ReadFile(path)
		if bytes.Equal(raw, before) {
			t.Fatal("response preceded persistence")
		}
		fresh, err := document.LoadTable(path)
		if err != nil {
			t.Fatal(err)
		}
		if fresh.Rows[d.RowIndex].ID != d.RowID || !reflect.DeepEqual(fresh.Rows[d.RowIndex].Values[d.Column], d.Cell) {
			t.Fatal("delta differs from saved source")
		}
		u, err := s.UndoDelta()
		if err != nil || u.Delta == nil {
			t.Fatalf("undo: %v", err)
		}
		restored, _ := os.ReadFile(path)
		if !bytes.Equal(before, restored) {
			t.Fatal("delta undo did not persist exact source")
		}
		redo, err := s.RedoDelta()
		if err != nil || redo.Delta == nil || !reflect.DeepEqual(redo.Delta.Cell, d.Cell) {
			t.Fatalf("redo: %v", err)
		}
	}
	r, err := s.SetCellTextDelta(CellUpdate{path, "first", "name", "中文, \"quoted\"\nsecond"})
	if err != nil || r.Delta == nil {
		t.Fatalf("text delta: %v", err)
	}
	if r.Delta.Cell.Display != "中文, \"quoted\"\nsecond" {
		t.Fatal("text delta changed escaping")
	}
	if _, err := s.SetCellFormula(FormulaSet{Path: path, RowID: "first", Column: "price", Expression: "1 + 2"}); err != nil {
		t.Fatal(err)
	}
	full, err := s.UpdateCellDelta(CellUpdate{path, "middle", "price", "3"})
	if err != nil || full.Delta != nil || full.Table == nil {
		t.Fatalf("formula table must use hydrated fallback: %v", err)
	}
}
