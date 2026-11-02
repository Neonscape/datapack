package workspace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"datapack/internal/document"
	"datapack/internal/sourcefile"
)

func literalProject(t *testing.T) (*Service, string, []byte) {
	t.Helper()
	t.Setenv("DATAPACK_FAST_EDIT", "1")
	root := t.TempDir()
	config := t.TempDir()
	writeFile(t, filepath.Join(root, "literal.dtproj"), `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "tables", "literal.cue")
	content := []byte(`package literal
// Keep this comment and spacing exactly.
#PrimaryKey: string
#Row: {name: string, price: int, enabled: bool, category: "a" | "b"}
rows: {
 first: {name: "中文,\"quoted\"", price: 1, enabled: true, category: "a"} // first comment
 middle: {name: "middle", price: 9007199254740993, enabled: false, category: "b"}
 last: {name: "last", price: -1, enabled: true, category: "a"}
}

_tool: {table: {namespace: "test", name: "literal"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "name", "price", "enabled", "category"]}}
`)
	writeFile(t, path, string(content))
	settings := DefaultEditorSettings()
	settings.Build.Mode = "manual"
	if err := SaveEditorSettings(config, settings); err != nil {
		t.Fatal(err)
	}
	s := NewService(config)
	if _, err := s.OpenProject(filepath.Join(root, "literal.dtproj")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetTable(path); err != nil {
		t.Fatal(err)
	}
	if s.literalTables[path] == nil {
		t.Fatal("snapshot not prepared")
	}
	return s, path, content
}
func TestResearchLiteralPersistHistoryAndCopyOnWrite(t *testing.T) {
	s, path, original := literalProject(t)
	oldTable, _ := s.GetTable(path)
	updates := []CellUpdate{{path, "first", "price", "123456"}, {path, "middle", "enabled", "true"}, {path, "last", "category", `"b"`}, {path, "first", "price", "2"}}
	for _, update := range updates {
		r, err := s.UpdateCell(update)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Table.Rows) != 3 || !r.Workspace.Compile.Pending {
			t.Fatal("wrong snapshot state")
		}
		if len(s.undoStack[len(s.undoStack)-1].Files) != 0 || s.undoStack[len(s.undoStack)-1].Cell == nil {
			t.Fatal("expected bounded cell history")
		}
	}
	if oldTable.Rows[0].Values["price"].Expression != "1" {
		t.Fatal("published table mutated after edit")
	}
	if _, err := s.SetCellText(CellUpdate{path, "last", "name", "新名字, \"引号\"\n第二行"}); err != nil {
		t.Fatal(err)
	}
	actual, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if actual.Rows[2].Values["name"].Display != "新名字, \"引号\"\n第二行" {
		t.Fatalf("bad saved string: %q", actual.Rows[2].Values["name"].Display)
	}
	for i := 0; i < 5; i++ {
		if _, err := s.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	restored, _ := os.ReadFile(path)
	if !bytes.Equal(restored, original) {
		t.Fatal("undo failed to restore all original source bytes")
	}
	for i := 0; i < 5; i++ {
		if _, err := s.Redo(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetCompileMode("auto"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal("undo after switching mode:", err)
	}
}
func TestResearchLiteralExternalChangesAreNotOverwritten(t *testing.T) {
	s, path, original := literalProject(t)
	external := append(append([]byte{}, original...), []byte("\n// EXTERNAL EDIT\n")...)
	if err := os.WriteFile(path, external, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "2"}); err == nil {
		t.Fatal("external change silently overwritten")
	}
	actual, _ := os.ReadFile(path)
	if !bytes.Equal(actual, external) {
		t.Fatal("external source lost")
	}
}
func TestResearchLiteralInvalidInputAndHistoryConflict(t *testing.T) {
	s, path, original := literalProject(t)
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "{broken"}); err == nil {
		t.Fatal("invalid input accepted")
	}
	actual, _ := os.ReadFile(path)
	if !bytes.Equal(actual, original) {
		t.Fatal("invalid expression modified source")
	}
	// Reload after the slow path invalidates the literal cache.
	if _, err := s.GetTable(path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "8"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	raw = append(raw, []byte("\n// newer external content\n")...)
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err == nil {
		t.Fatal("undo overwrote external edit")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, raw) {
		t.Fatal("external history content lost")
	}
}
func TestResearchLiteralLongSequenceKeepsAllUndo(t *testing.T) {
	s, path, original := literalProject(t)
	for i := 0; i < 100; i++ {
		if _, err := s.UpdateCell(CellUpdate{path, "first", "price", fmt.Sprint(i + 10)}); err != nil {
			t.Fatal(err)
		}
	}
	if len(s.undoStack) != 100 {
		t.Fatal("history truncated")
	}
	for i := 0; i < 100; i++ {
		if _, err := s.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(after, original) {
		t.Fatal("100 undo operations lost source")
	}
}

func TestResearchLiteralFailedWritesPreserveSourceAndHistory(t *testing.T) {
	for _, stage := range []string{"stage", "rename"} {
		t.Run(stage, func(t *testing.T) {
			s, path, original := literalProject(t)
			ops := literalWriteOps{stage: sourcefile.Stage, rename: os.Rename}
			failure := fmt.Errorf("injected %s failure", stage)
			if stage == "stage" {
				ops.stage = func(string, []byte, os.FileMode) (string, error) { return "", failure }
			} else {
				ops.rename = func(string, string) error { return failure }
			}
			s.literalOps = &ops
			if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "99"}); err == nil {
				t.Fatal("write failure hidden")
			}
			actual, _ := os.ReadFile(path)
			if !bytes.Equal(actual, original) {
				t.Fatal("failed write modified source")
			}
			if len(s.undoStack) != 0 || len(s.redoStack) != 0 {
				t.Fatal("failed write consumed history")
			}
			s.literalOps = nil
			if _, err := s.UpdateCell(CellUpdate{path, "first", "price", "99"}); err != nil {
				t.Fatal("safe retry failed:", err)
			}
		})
	}
}

func TestResearchLiteralPublishedTablesRemainSafeDuringConcurrentMarshalling(t *testing.T) {
	s, path, _ := literalProject(t)
	var wg sync.WaitGroup
	failures := make(chan error, 4)
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				table, err := s.GetTable(path)
				if err != nil {
					failures <- err
					return
				}
				if _, err := json.Marshal(table); err != nil {
					failures <- err
					return
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			r, err := s.UpdateCell(CellUpdate{path, "first", "price", fmt.Sprint(i + 1)})
			if err != nil {
				failures <- err
				return
			}
			if _, err := json.Marshal(r); err != nil {
				failures <- err
				return
			}
		}
	}()
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}
