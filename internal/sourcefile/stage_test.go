package sourcefile

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type failingFile struct {
	*os.File
	phase string
}

func (f failingFile) Chmod(mode os.FileMode) error {
	if f.phase == "chmod" {
		return errors.New("injected chmod")
	}
	return f.File.Chmod(mode)
}
func (f failingFile) Write(data []byte) (int, error) {
	if f.phase == "write" {
		return 0, errors.New("injected write")
	}
	return f.File.Write(data)
}
func (f failingFile) Sync() error {
	if f.phase == "sync" {
		return errors.New("injected sync")
	}
	return f.File.Sync()
}
func (f failingFile) Close() error {
	err := f.File.Close()
	if f.phase == "close" {
		return errors.Join(err, errors.New("injected close"))
	}
	return err
}

func TestStageFailuresLeaveOriginalAndNoTemporarySource(t *testing.T) {
	for _, phase := range []string{"create", "chmod", "write", "sync", "close"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "original.cue")
			if err := os.WriteFile(path, []byte("original bytes"), 0640); err != nil {
				t.Fatal(err)
			}
			_, err := stage(path, []byte("replacement"), 0644, func(dir, pattern string) (stagedFile, error) {
				if phase == "create" {
					return nil, errors.New("injected create")
				}
				file, err := os.CreateTemp(dir, pattern)
				if err != nil {
					return nil, err
				}
				return failingFile{File: file, phase: phase}, nil
			})
			if err == nil || !strings.Contains(err.Error(), "injected "+phase) {
				t.Fatalf("phase error: %v", err)
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != "original bytes" {
				t.Fatalf("original changed: %q %v", data, err)
			}
			staged, err := filepath.Glob(filepath.Join(root, ".datapack-stage-*"))
			if err != nil || len(staged) != 0 {
				t.Fatalf("temporary sources: %v %v", staged, err)
			}
		})
	}
}

func TestStagePreservesArbitraryBytesAndModeWithoutReplacingSource(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "original.cue")
	original := []byte("original")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range [][]byte{{}, {0, 1, 2}, []byte("invalid CUE {"), []byte("rows: {one: {invalidBusinessValue: 5}}")} {
		staged, err := Stage(path, candidate, 0640)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(staged)
		if err != nil || string(data) != string(candidate) {
			t.Fatalf("raw bytes: %v %v", data, err)
		}
		info, err := os.Stat(staged)
		if err != nil || runtime.GOOS != "windows" && info.Mode().Perm() != 0640 {
			t.Fatalf("mode: %v %v", info, err)
		}
		source, err := os.ReadFile(path)
		if err != nil || string(source) != string(original) {
			t.Fatal("staging replaced source")
		}
		if err := os.Remove(staged); err != nil {
			t.Fatal(err)
		}
	}
}
