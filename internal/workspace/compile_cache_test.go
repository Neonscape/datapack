package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResearchCompileCacheInvalidatesAndManualOpenIsPending(t *testing.T) {
	t.Setenv("DATAPACK_COMPILE_CACHE", "1")
	t.Setenv("DATAPACK_DEFER_OPEN_COMPILE", "1")
	s, path, _ := literalProject(t)
	view, ok := s.Current()
	if !ok || !view.Compile.Pending || view.Compile.CompiledAt != "" {
		t.Fatal("open pretended to have current validation")
	}
	if _, err := s.Compile(); err != nil {
		t.Fatal(err)
	}
	first := s.successfulCompile
	if first == nil {
		t.Fatal("success not cached")
	}
	if _, err := s.Compile(); err != nil {
		t.Fatal(err)
	}
	if s.successfulCompile != first {
		t.Fatal("unchanged compile did not reuse immutable package")
	}
	if _, err := s.UpdateCell(CellUpdate{path, "first", "price", `"invalid int"`}); err != nil {
		t.Fatal(err)
	}
	result, err := s.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Workspace.Compile.Diagnostics) == 0 {
		t.Fatal("edited invalid data used stale clean result")
	}
	if _, err := s.Export(); err == nil {
		t.Fatal("invalid edited data exported")
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	raw = []byte(strings.Replace(string(raw), "price: 1,", "price: 9,", 1))
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Compile(); err != nil {
		t.Fatal(err)
	}
	if s.successfulCompile == first {
		t.Fatal("external changed bytes used stale package")
	}
}

func TestResearchCompileDiscoversNewExternalFiles(t *testing.T) {
	t.Setenv("DATAPACK_COMPILE_CACHE", "1")
	s, path, _ := literalProject(t)
	if _, err := s.Compile(); err != nil {
		t.Fatal(err)
	}
	newPath := filepath.Join(filepath.Dir(path), "new-damaged.cue")
	if err := os.WriteFile(newPath, []byte("package broken\nthis is invalid {"), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := s.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Workspace.Compile.Diagnostics) == 0 {
		t.Fatal("new external damaged file was skipped by cache/tree")
	}
}

func TestResearchSharedViewKeepsTableLoadDiagnostics(t *testing.T) {
	s, path, _ := literalProject(t)
	raw, _ := os.ReadFile(path)
	raw = []byte(strings.Replace(string(raw), "name: string,", "name?: string,", 1))
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatal(err)
	}
	view, _ := s.Current()
	reopened, err := s.OpenProject(view.Project.Path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.ForeignKeys.Diagnostics) == 0 || !strings.Contains(reopened.ForeignKeys.Diagnostics[0].Message, "cannot load reference source") {
		t.Fatal("shared document collection lost reference-index load failure")
	}
}
