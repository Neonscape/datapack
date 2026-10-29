package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"datapack/internal/logging"
)

// Imports and the requested edit share one original snapshot. Failed edits or
// refreshes compensate source changes before returning; only a complete edit
// is published to history. This is not a crash/concurrent filesystem transaction.
func editSource[T any](s *Service, path string, operation func() (T, error)) (result T, view View, err error) {
	started := time.Now()
	before, err := os.ReadFile(path)
	if err != nil {
		return result, View{}, fmt.Errorf("record edit source: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return result, View{}, fmt.Errorf("record edit permissions: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, restoreSourceFile(path, before, info.Mode().Perm()))
		}
	}()
	result, err = operation()
	if err != nil {
		s.log(logging.Error, "source edit failed", map[string]any{"path": path, "error": err.Error()})
		var zero T
		return zero, View{}, err
	}
	if err = preserveSourceMode(path, info.Mode().Perm()); err != nil {
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
	if info.Mode().Perm() != mode {
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("preserve source permissions %q: %w", path, err)
		}
	}
	return nil
}
