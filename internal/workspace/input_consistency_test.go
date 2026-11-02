package workspace

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCompileInputChangeRejectedWithAndWithoutCache(t *testing.T) {
	for _, imports := range []bool{false, true} {
		t.Run(map[bool]string{false: "cacheable", true: "import_fallback"}[imports], func(t *testing.T) {
			s, root, _ := dependencyFixture(t, imports, false)
			s.successfulCompile = nil
			s.inputHooks = &inputConsistencyHooks{afterCompile: func() { writeFile(t, filepath.Join(root, "constraints.cue"), "package item\n#Row: {price: <=5}\n") }}
			result, err := s.Compile()
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Workspace.Compile.Diagnostics) == 0 {
				t.Fatal("mixed input compile published as clean")
			}
		})
	}
}
func TestExportChangeAfterCacheHitCannotPublish(t *testing.T) {
	s, root, _ := dependencyFixture(t, false, false)
	if _, err := s.Export(); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "exports", "rpg.item.csv")
	before, _ := os.ReadFile(output)
	s.inputHooks = &inputConsistencyHooks{afterCacheHit: func() { writeFile(t, filepath.Join(root, "constraints.cue"), "package item\n#Row: {price: <=5}\n") }}
	if _, err := s.Export(); err == nil {
		t.Fatal("changed input cache result exported")
	}
	after, _ := os.ReadFile(output)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected cache hit replaced output")
	}
}
func TestExportCommitFenceWorksForImportFallback(t *testing.T) {
	for _, imports := range []bool{false, true} {
		t.Run(map[bool]string{false: "cacheable", true: "import_fallback"}[imports], func(t *testing.T) {
			s, root, _ := dependencyFixture(t, imports, false)
			if _, err := s.Export(); err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(root, "exports", "rpg.item.csv")
			before, _ := os.ReadFile(output)
			s.inputHooks = &inputConsistencyHooks{beforeExportCommit: func() { writeFile(t, filepath.Join(root, "constraints.cue"), "package item\n#Row: {price: <=5}\n") }}
			if _, err := s.Export(); err == nil {
				t.Fatal("input changed before export commit was not rejected")
			}
			after, _ := os.ReadFile(output)
			if !bytes.Equal(before, after) {
				t.Fatal("export fence replaced previous valid data")
			}
		})
	}
}
