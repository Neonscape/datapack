package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"datapack/internal/document"
)

// Keep filesystem operations local to a table rename, so failures at each
// boundary can be tested without process-wide hooks or permissions tricks.
type tableRenameOps struct {
	rewrite     func(string, string, string, []document.TableReference) (document.Table, error)
	rename      func(string, string) error
	mkdirAll    func(string, os.FileMode) error
	saveProject func(Project) error
}

func defaultTableRenameOps() tableRenameOps {
	return tableRenameOps{
		rewrite: document.RenameTable, rename: os.Rename,
		mkdirAll: os.MkdirAll, saveProject: SaveProject,
	}
}

// Unlike the legacy snapshot helper, history and rename must not mistake an
// unreadable existing file for a missing file and later remove it on rollback.
func snapshotRenameFiles(paths ...string) (map[string][]byte, error) {
	snapshot := map[string][]byte{}
	for _, path := range paths {
		if _, exists := snapshot[path]; exists {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("snapshot rename source %q: %w", path, err)
		}
		if os.IsNotExist(err) {
			if _, inspectErr := os.Lstat(path); inspectErr == nil {
				return nil, fmt.Errorf("snapshot source %q exists but cannot be read: %w", path, err)
			} else if !os.IsNotExist(inspectErr) {
				return nil, fmt.Errorf("inspect snapshot source %q: %w", path, inspectErr)
			}
		}
		snapshot[path] = content
	}
	return snapshot, nil
}

type renameMove struct{ from, to string }

// This is compensating rollback for reported I/O failures, not a crash-safe or
// concurrently observable multi-file filesystem transaction. Moves are undone
// before restoring bytes, preserving sibling assets in the moved directory.
type tableRenameTransaction struct {
	snapshot map[string][]byte
	modes    map[string]os.FileMode
	rename   func(string, string) error
	moves    []renameMove
}

func newTableRenameTransaction(snapshot map[string][]byte, rename func(string, string) error) (*tableRenameTransaction, error) {
	t := &tableRenameTransaction{snapshot: snapshot, modes: map[string]os.FileMode{}, rename: rename}
	for path, content := range snapshot {
		if content == nil {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("snapshot rename permissions %q: %w", path, err)
		}
		t.modes[path] = info.Mode().Perm()
	}
	return t, nil
}

func (t *tableRenameTransaction) move(from, to string) error {
	if err := t.rename(from, to); err != nil {
		return err
	}
	t.moves = append(t.moves, renameMove{from, to})
	return nil
}

func (t *tableRenameTransaction) rollback() error {
	var failures []error
	for index := len(t.moves) - 1; index >= 0; index-- {
		move := t.moves[index]
		if err := t.rename(move.to, move.from); err != nil {
			failures = append(failures, fmt.Errorf("rollback rename %q to %q: %w", move.to, move.from, err))
			// Do not restore/delete snapshot paths against a directory layout
			// that could not be restored: that could destroy the only copy.
			return errors.Join(failures...)
		}
	}
	for _, path := range snapshotPaths(t.snapshot) {
		before := t.snapshot[path]
		current, err := os.ReadFile(path)
		if before == nil {
			if os.IsNotExist(err) {
				continue
			}
			err = os.Remove(path)
		} else {
			if err != nil || !bytes.Equal(current, before) {
				err = writeFileAtomic(path, before)
			}
			if err == nil {
				err = os.Chmod(path, t.modes[path])
			}
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("rollback rename source %q: %w", path, err))
		}
	}
	return errors.Join(failures...)
}
