package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"datapack/internal/logging"
)

// Imports and the requested edit share one original snapshot. Failed edits or
// refreshes compensate source changes before returning; only a complete edit
// is published to history. This is not a crash/concurrent filesystem transaction.
func editSource[T any](s *Service, path string, operation func() (T, error)) (result T, view View, err error) {
	started := time.Now()
	info, err := os.Lstat(path)
	if err != nil {
		return result, View{}, fmt.Errorf("record edit permissions: %w", err)
	}
	if !info.Mode().IsRegular() {
		return result, View{}, fmt.Errorf("source %q is not a regular file; symlinks are unsupported", path)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		return result, View{}, fmt.Errorf("record edit source: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, restoreSourceFile(path, before, sourcePermissionMode(info.Mode())))
		}
	}()
	result, err = operation()
	if err != nil {
		s.log(logging.Error, "source edit failed", map[string]any{"path": path, "error": err.Error()})
		var zero T
		return zero, View{}, err
	}
	if err = preserveSourceMode(path, sourcePermissionMode(info.Mode())); err != nil {
		var zero T
		return zero, View{}, err
	}
	view, err = s.refreshAfterEdit()
	if err != nil {
		s.log(logging.Error, "workspace refresh failed after edit", map[string]any{"path": path, "error": err.Error()})
		var zero T
		return zero, View{}, err
	}
	s.recordEdit(map[string][]byte{path: before})
	view.CanUndo, view.CanRedo = true, false
	s.current = &view
	s.log(logging.Info, "source edit completed", map[string]any{"path": path, "duration_ms": time.Since(started).Milliseconds()})
	return result, view, nil
}

func restoreSourceFile(path string, before []byte, mode os.FileMode) error {
	current, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, current) {
		if err := writeFileAtomic(path, before); err != nil {
			return fmt.Errorf("rollback edit source %q: %w", path, err)
		}
	}
	if err := preserveSourceMode(path, mode); err != nil {
		return fmt.Errorf("rollback edit permissions: %w", err)
	}
	return nil
}

func preserveSourceMode(path string, mode os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("inspect source permissions %q: %w", path, err)
	}
	if sourcePermissionMode(info.Mode()) != sourcePermissionMode(mode) {
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("preserve source permissions %q: %w", path, err)
		}
	}
	return nil
}

// editSources publishes a multi-file command only after all source writes and
// workspace refresh succeed. On reported failure it restores bytes and modes;
// the Service lock remains owned by the command. No crash/concurrent FS claim.
func editSources[T any](s *Service, paths []string, operation func() (T, error)) (result T, view View, err error) {
	return editSourcesWithOps(s, paths, operation, sourceCommandOps{refresh: s.refreshAfterEdit, restore: writeFileAtomic})
}

type sourceCommandOps struct {
	refresh func() (View, error)
	restore func(string, []byte) error
}

func editSourcesWithOps[T any](s *Service, paths []string, operation func() (T, error), ops sourceCommandOps) (result T, view View, err error) {
	snapshot, err := snapshotRenameFiles(paths...)
	if err != nil {
		return result, View{}, err
	}
	transaction, err := newTableRenameTransaction(snapshot, os.Rename)
	if err != nil {
		return result, View{}, err
	}
	transaction.restore = ops.restore
	defer func() {
		if err != nil {
			err = errors.Join(err, transaction.rollback())
		}
	}()
	result, err = operation()
	if err != nil {
		var zero T
		return zero, View{}, err
	}
	for path, mode := range transaction.modes {
		if err = preserveSourceMode(path, mode); err != nil {
			var zero T
			return zero, View{}, err
		}
	}
	view, err = ops.refresh()
	if err != nil {
		var zero T
		return zero, View{}, err
	}
	s.recordEdit(snapshot)
	view.CanUndo, view.CanRedo = true, false
	s.current = &view
	return result, view, nil
}

// Missing directories are owned only until the command succeeds. os.Remove
// refuses a nonempty directory, so compensation cannot remove another file.
func missingParentDirectories(directory string) ([]string, error) {
	var missing []string
	for current := directory; ; current = filepath.Dir(current) {
		info, err := os.Stat(current)
		if err == nil {
			if !info.IsDir() {
				return nil, fmt.Errorf("source parent %q is not a directory", current)
			}
			return missing, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		missing = append(missing, current)
		if filepath.Dir(current) == current {
			return nil, fmt.Errorf("source parent has no existing ancestor")
		}
	}
}

func removeCreatedDirectories(paths []string) error {
	var failures []error
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			failures = append(failures, fmt.Errorf("rollback created directory %q: %w", path, err))
		}
	}
	return errors.Join(failures...)
}

func sourcePermissionMode(mode os.FileMode) os.FileMode {
	return mode & (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)
}
