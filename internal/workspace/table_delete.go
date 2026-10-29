package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"datapack/internal/document"
	"datapack/internal/logging"
)

// Directory history owns bytes, entry kinds, permissions and link text, never
// symlink targets. It is process-local compensation, not crash recovery; mtimes,
// ownership, ACLs and hard-link relationships are not part of editor history.
type directoryEntry struct {
	Mode os.FileMode
	Data []byte
	Link string
}
type directorySnapshot map[string]directoryEntry

type directoryHistory struct {
	Path     string
	Expected directorySnapshot
	Restore  directorySnapshot
}

func (h *directoryHistory) inverse() *directoryHistory {
	if h == nil {
		return nil
	}
	return &directoryHistory{Path: h.Path, Expected: h.Restore, Restore: h.Expected}
}

// Reject symlink ancestors before any read/write: WalkDir alone does not
// protect a root reached through a symlink in one of its parent components.
func safeDirectoryPath(projectRoot, path string) error {
	if !filepath.IsAbs(path) || !isWithin(path, projectRoot) || path == projectRoot {
		return fmt.Errorf("table directory is outside project: %s", path)
	}
	for current := filepath.Dir(path); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect directory parent: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("directory parent is not a real directory: %s", current)
		}
		if current == projectRoot {
			break
		}
	}
	return nil
}

func snapshotDirectory(projectRoot, root string) (directorySnapshot, error) {
	if err := safeDirectoryPath(projectRoot, root); err != nil {
		return nil, err
	}
	info, err := os.Lstat(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("table source directory is not a real directory: %s", root)
	}
	result := directorySnapshot{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := directoryEntry{Mode: info.Mode()}
		switch {
		case info.Mode().IsRegular():
			value.Data, err = os.ReadFile(path)
		case info.Mode()&os.ModeSymlink != 0:
			value.Link, err = os.Readlink(path)
		case info.IsDir():
		default:
			return fmt.Errorf("unsupported attachment kind at %s", path)
		}
		if err != nil {
			return err
		}
		result[relative] = value
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("snapshot table directory: %w", err)
	}
	return result, nil
}

func materializeDirectory(root string, snapshot directorySnapshot) error {
	if entry, ok := snapshot["."]; !ok || !entry.Mode.IsDir() {
		return fmt.Errorf("directory history has no real root")
	}
	paths := make([]string, 0, len(snapshot))
	for path := range snapshot {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, relative := range paths {
		if !filepath.IsLocal(relative) {
			return fmt.Errorf("invalid directory history path: %s", relative)
		}
		entry := snapshot[relative]
		if relative != "." {
			if parent, ok := snapshot[filepath.Dir(relative)]; !ok || !parent.Mode.IsDir() {
				return fmt.Errorf("directory history parent is not a directory: %s", relative)
			}
		}
		path := filepath.Join(root, relative)
		switch {
		case entry.Mode.IsDir():
			if err := os.Mkdir(path, 0700); err != nil {
				return err
			}
		case entry.Mode.IsRegular():
			if err := os.WriteFile(path, entry.Data, 0600); err != nil {
				return err
			}
		case entry.Mode&os.ModeSymlink != 0:
			if err := os.Symlink(entry.Link, path); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported directory history kind: %s", relative)
		}
	}
	// Apply permissions last, children first, so read-only directories do not
	// obstruct reconstruction. Never chmod through an attachment symlink.
	for i := len(paths) - 1; i >= 0; i-- {
		entry := snapshot[paths[i]]
		if entry.Mode&os.ModeSymlink == 0 {
			mode := entry.Mode & (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)
			if err := os.Chmod(filepath.Join(root, paths[i]), mode); err != nil {
				return err
			}
		}
	}
	return nil
}

type directoryTransaction struct {
	path, stage            string
	hadOriginal, installed bool
	rename                 func(string, string) error
}

// Staging lives outside configured source roots, so a refresh cannot discover
// a second copy of the deleted/restored table. Each operation owns a fresh slot.
func changeDirectory(project Project, history *directoryHistory, rename func(string, string) error) (_ *directoryTransaction, err error) {
	current, err := snapshotDirectory(project.Root, history.Path)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(current, history.Expected) {
		return nil, fmt.Errorf("table directory changed since history was recorded: %s", history.Path)
	}
	stage, err := os.MkdirTemp(project.Root, ".datapack-table-history-*")
	if err != nil {
		return nil, fmt.Errorf("create table history staging: %w", err)
	}
	t := &directoryTransaction{path: history.Path, stage: stage, rename: rename}
	defer func() {
		if err != nil {
			err = errors.Join(err, t.rollback())
		}
	}()
	for _, root := range append(append([]string{}, project.Config.TableRoots...), project.Config.SchemaRoots...) {
		if pathsOverlap(stage, filepath.Join(project.Root, root)) {
			return nil, fmt.Errorf("table history staging overlaps source root")
		}
	}
	if history.Restore != nil {
		if err := materializeDirectory(filepath.Join(stage, "restored"), history.Restore); err != nil {
			return nil, fmt.Errorf("prepare table history: %w", err)
		}
	}
	if current != nil {
		if err := rename(history.Path, filepath.Join(stage, "original")); err != nil {
			return nil, fmt.Errorf("stage table directory: %w", err)
		}
		t.hadOriginal = true
	}
	if history.Restore != nil {
		if err := rename(filepath.Join(stage, "restored"), history.Path); err != nil {
			return nil, fmt.Errorf("install table history: %w", err)
		}
		t.installed = true
	}
	return t, nil
}
func (t *directoryTransaction) rollback() error {
	if t.installed {
		if err := t.rename(t.path, filepath.Join(t.stage, "restored")); err != nil {
			return fmt.Errorf("rollback restored directory (recovery at %s): %w", t.stage, err)
		}
		t.installed = false
	}
	if t.hadOriginal {
		if _, err := os.Lstat(t.path); !os.IsNotExist(err) {
			return fmt.Errorf("rollback directory destination occupied or unreadable; recovery at %s", t.stage)
		}
		if err := t.rename(filepath.Join(t.stage, "original"), t.path); err != nil {
			return fmt.Errorf("rollback table directory (recovery at %s): %w", t.stage, err)
		}
		t.hadOriginal = false
	}
	return t.cleanup()
}
func (t *directoryTransaction) cleanup() error {
	// Remove only our fresh staging slot. Make its directories writable before
	// cleanup; original read-only permissions remain in the in-memory history.
	// WalkDir and RemoveAll never follow stored attachment links.
	if err := filepath.WalkDir(t.stage, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(path, 0700)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("prepare history staging cleanup %s: %w", t.stage, err)
	}
	if err := os.RemoveAll(t.stage); err != nil {
		return fmt.Errorf("clean table history staging %s: %w", t.stage, err)
	}
	return nil
}

type tableDeleteOps struct {
	rename      func(string, string) error
	saveProject func(Project) error
}

func (s *Service) deleteTable(delete TableDelete, ops tableDeleteOps) (result TableDeleteResult, err error) {
	if s.current == nil {
		return result, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, delete.Path) {
		return result, s.userError("workspace.tableNotInProject", delete.Path)
	}
	directory := filepath.Dir(delete.Path)
	project := s.current.Project
	if err := safeDirectoryPath(project.Root, directory); err != nil {
		return result, err
	}
	if info, err := os.Lstat(directory); err != nil || !info.IsDir() {
		return result, fmt.Errorf("table directory is missing or is not a real directory: %s", directory)
	}
	if info, err := os.Lstat(delete.Path); err != nil || !info.Mode().IsRegular() {
		return result, fmt.Errorf("table source is missing or is not a regular file: %s", delete.Path)
	}
	source, err := document.LoadTable(delete.Path)
	if err != nil {
		return result, err
	}
	name := logicalTableName(source)
	// A delete owns exactly one table directory, never a configured source root
	// or a directory containing another registered table/schema/project file.
	for _, root := range project.Config.SchemaRoots {
		if pathsOverlap(filepath.Join(project.Root, root), directory) {
			return result, fmt.Errorf("table directory overlaps a schema root")
		}
	}
	for _, root := range project.Config.TableRoots {
		if isWithin(filepath.Join(project.Root, root), directory) {
			return result, fmt.Errorf("table directory contains a configured source root")
		}
	}
	for _, path := range []string{project.Path, filepath.Join(project.Root, "cue.mod"), filepath.Join(project.Root, ".git")} {
		if pathsOverlap(directory, path) {
			return result, fmt.Errorf("table directory overlaps protected project files")
		}
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	for _, table := range tables {
		if table.Path == delete.Path {
			continue
		}
		if isWithin(table.Path, directory) {
			return result, fmt.Errorf("table directory contains another table: %s", table.Path)
		}
		for _, key := range table.ForeignKeys {
			if key.Table == name {
				return result, s.userError("workspace.tableReferenced", name, logicalTableName(table)+"."+key.Column)
			}
		}
	}
	tree, err := snapshotDirectory(project.Root, directory)
	if err != nil {
		return result, err
	}
	relative, err := filepath.Rel(directory, delete.Path)
	if err != nil {
		return result, err
	}
	if entry, ok := tree[relative]; !ok || !entry.Mode.IsRegular() {
		return result, fmt.Errorf("table source is not a regular file")
	}
	files, err := snapshotRenameFiles(project.Path)
	if err != nil {
		return result, err
	}
	fileTransaction, err := newTableRenameTransaction(files, os.Rename)
	if err != nil {
		return result, err
	}
	project.Config.Export.Tables = maps.Clone(project.Config.Export.Tables)
	history := &directoryHistory{Path: directory, Expected: tree}
	transaction, err := changeDirectory(project, history, ops.rename)
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, transaction.rollback(), fileTransaction.rollback())
		} else if cleanupErr := transaction.cleanup(); cleanupErr != nil {
			s.log(logging.Warn, "table deletion staging cleanup failed", map[string]any{"error": cleanupErr.Error()})
		}
	}()
	if removeExportTableSelection(&project.Config, name) {
		if err := ops.saveProject(project); err != nil {
			return result, err
		}
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return result, err
	}
	s.recordHistory(historyEntry{Files: files, Directory: history.inverse()})
	view.CanUndo, view.CanRedo = true, false
	s.current = &view
	s.log(logging.Info, "table deleted", map[string]any{"path": delete.Path, "name": name})
	return TableDeleteResult{Workspace: view}, nil
}
