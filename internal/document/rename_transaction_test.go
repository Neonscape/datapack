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
	first, second := filepath.Join(root, "first.cue"), filepath.Join(root, "second.cue")
	originals := map[string][]byte{first: []byte("first original"), second: []byte("second original")}
	candidates := map[string][]byte{first: []byte("first candidate"), second: []byte("second candidate")}
	modes := map[string]os.FileMode{first: 0640, second: 0644}
	for path, data := range originals {
		if err := os.WriteFile(path, data, modes[path]); err != nil {
			t.Fatal(err)
		}
	}
	calls := 0
	firstCommitted := false
	ops := sourceReplaceOps{rename: func(from, to string) error {
		calls++
		if calls == 2 {
			return errors.New("injected second replacement failure")
		}
		err := os.Rename(from, to)
		if calls == 1 && err == nil {
			bytes, _ := os.ReadFile(first)
			firstCommitted = string(bytes) == string(candidates[first])
		}
		return err
	}}
	if err := replaceAllWithOps([]string{first, second}, candidates, originals, ops); err == nil {
		t.Fatal("second replacement succeeded")
	}
	if calls != 3 || !firstCommitted {
		t.Fatalf("not a committed-then-restored transaction: calls=%d committed=%v", calls, firstCommitted)
	}
	for path, want := range originals {
		bytes, err := os.ReadFile(path)
		if err != nil || string(bytes) != string(want) {
			t.Fatalf("source %s: %q %v", path, bytes, err)
		}
		info, err := os.Stat(path)
		if err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != modes[path] {
			t.Fatalf("mode %s: %v %v", path, info, err)
		}
	}
	assertNoRenameTemporaryFiles(t, root)
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

func TestReplaceAllKeepsSpecialPermissionBits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.cue")
	if err := os.WriteFile(path, []byte("original"), 0640); err != nil {
		t.Fatal(err)
	}
	wanted := os.FileMode(0640) | os.ModeSetgid | os.ModeSticky
	if err := os.Chmod(path, wanted); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	mask := os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky
	if info.Mode()&mask != wanted {
		t.Skip("filesystem does not retain tested special mode")
	}
	if err := replaceAll([]string{path}, map[string][]byte{path: []byte("candidate")}, map[string][]byte{path: []byte("original")}); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(path)
	if err != nil || info.Mode()&mask != wanted {
		t.Fatalf("special mode lost: %v %v", info, err)
	}
}
