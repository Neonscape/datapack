package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Files stores the destination bytes; Moves restores directory and filename
// positions without copying or discarding unrelated sidecars. This is process
// local compensating history, not a crash or external-concurrency transaction.
type historyEntry struct {
	Files map[string][]byte
	Moves []renameMove
}

func reverseRenameMoves(moves []renameMove) []renameMove {
	if len(moves) == 0 {
		return nil
	}
	result := make([]renameMove, 0, len(moves))
	for i := len(moves) - 1; i >= 0; i-- {
		result = append(result, renameMove{from: moves[i].to, to: moves[i].from})
	}
	return result
}

// Per-call operations allow reported write failures to be tested without
// process-wide hooks, changing permissions, or relying on platform races.
type snapshotRestoreOps struct {
	write    func(string, []byte) error
	remove   func(string) error
	mkdirAll func(string, os.FileMode) error
}

func restoreSnapshotWithOps(snapshot map[string][]byte, ops snapshotRestoreOps) error {
	paths := snapshotPaths(snapshot)
	before, err := snapshotRenameFiles(paths...)
	if err != nil {
		return fmt.Errorf("capture current history sources: %w", err)
	}
	apply := func(path string, content []byte) error {
		if content == nil {
			if err := ops.remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		if err := ops.mkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return ops.write(path, content)
	}
	var committed []string
	for _, path := range paths {
		if err := apply(path, snapshot[path]); err != nil {
			failures := []error{fmt.Errorf("restore history source %q: %w", path, err)}
			for i := len(committed) - 1; i >= 0; i-- {
				path := committed[i]
				if err := apply(path, before[path]); err != nil {
					failures = append(failures, fmt.Errorf("rollback history source %q: %w", path, err))
				}
			}
			return errors.Join(failures...)
		}
		committed = append(committed, path)
	}
	return nil
}
