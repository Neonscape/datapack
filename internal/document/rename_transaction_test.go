package document

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	paths, err := filepath.Glob(filepath.Join(root, ".datapack-*"))
	if err != nil || len(paths) != 0 {
		t.Fatalf("rename leaked staged files = %v, %v", paths, err)
	}
}

func TestReplaceAllPreservesSourcePermissions(t *testing.T) {
	root := t.TempDir()
	first, second := filepath.Join(root, "first.cue"), filepath.Join(root, "second.cue")
	if err := os.WriteFile(first, []byte("first"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := replaceAll([]string{first, second}, map[string][]byte{first: []byte("changed first"), second: []byte("changed second")}, map[string][]byte{first: []byte("first"), second: []byte("second")}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{first: 0640, second: 0644} {
		info, err := os.Stat(path)
		if err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != want {
			t.Fatalf("mode %s: %v %v, want %v", path, info, err, want)
		}
	}
}

func TestReplaceAllReportsRollbackFailureAndRetainsOriginalBytes(t *testing.T) {
	root := t.TempDir()
	first, second := filepath.Join(root, "first.cue"), filepath.Join(root, "second.cue")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte("original"), 0640); err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	ops := sourceReplaceOps{rename: func(from, to string) error {
		count++
		if count == 2 || count == 3 {
			return errors.New("injected rename failure")
		}
		return os.Rename(from, to)
	}}
	err := replaceAllWithOps([]string{first, second}, map[string][]byte{first: []byte("candidate"), second: []byte("candidate")}, map[string][]byte{first: []byte("original"), second: []byte("original")}, ops)
	if err == nil || !strings.Contains(err.Error(), "recovery bytes retained at") {
		t.Fatalf("rollback error: %v", err)
	}
	recovery, globErr := filepath.Glob(filepath.Join(root, ".datapack-stage-*"))
	if globErr != nil || len(recovery) != 1 {
		t.Fatalf("recovery copies: %v %v", recovery, globErr)
	}
	original, readErr := os.ReadFile(recovery[0])
	if readErr != nil || string(original) != "original" {
		t.Fatalf("recovery: %q %v", original, readErr)
	}
	current, _ := os.ReadFile(second)
	if string(current) != "original" {
		t.Fatal("uncommitted source changed")
	}
}
