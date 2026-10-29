package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceAllCleansStagedFilesWhenStagingFails(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.cue")
	second := filepath.Join(root, "missing", "second.cue")
	if err := os.WriteFile(first, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := replaceAll([]string{first, second}, map[string][]byte{first: []byte("candidate"), second: []byte("candidate")}, map[string][]byte{first: []byte("original")}); err == nil {
		t.Fatal("staging in a missing directory succeeded")
	}
	assertNoRenameTemporaryFiles(t, root)
	if source, err := os.ReadFile(first); err != nil || string(source) != "original" {
		t.Fatalf("first source = %q, %v", source, err)
	}
}

func TestReplaceAllCleansStagedFilesAndRestoresCommittedSources(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.cue")
	second := filepath.Join(root, "second.cue")
	if err := os.WriteFile(first, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Replacing a directory with the staged file fails after first commits.
	if err := os.Mkdir(second, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceAll([]string{first, second}, map[string][]byte{first: []byte("candidate"), second: []byte("candidate")}, map[string][]byte{first: []byte("original")}); err == nil {
		t.Fatal("replacement over a directory succeeded")
	}
	assertNoRenameTemporaryFiles(t, root)
	if source, err := os.ReadFile(first); err != nil || string(source) != "original" {
		t.Fatalf("first source = %q, %v", source, err)
	}
	if info, err := os.Stat(second); err != nil || !info.IsDir() {
		t.Fatalf("original directory was changed: %v", err)
	}
}

func assertNoRenameTemporaryFiles(t *testing.T, root string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, ".datapack-rename-*"))
	if err != nil || len(paths) != 0 {
		t.Fatalf("rename leaked staged files = %v, %v", paths, err)
	}
}
