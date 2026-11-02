package workspace

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"datapack/internal/document"
	"datapack/internal/perftrace"
	"datapack/internal/sourcefile"
)

type literalCache struct {
	snapshot    *document.LiteralSnapshot
	fingerprint string
}
type cellHistory struct{ Path, RowID, Column, Before, After, BeforeSHA, AfterSHA string }

func fastEnabled() bool           { return os.Getenv("DATAPACK_FAST_EDIT") != "0" }
func digest(source []byte) string { v := sha256.Sum256(source); return hex.EncodeToString(v[:]) }

func (s *Service) rememberLiteral(path string, table document.Table) {
	if !fastEnabled() || len(table.Formulas) != 0 || len(table.ForeignKeys) != 0 {
		return
	}
	if _, independent := document.SchemaCapsule(table); !independent {
		return
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return
	}
	table.Columns = append([]document.Column(nil), table.Columns...)
	snapshot, err := document.NewLiteralSnapshot(path, source, table)
	if err != nil {
		return
	}
	fingerprint, err := s.sourceFingerprint()
	if err != nil || fingerprint != s.literalInputFingerprint {
		return
	}
	if s.literalTables == nil {
		s.literalTables = map[string]*literalCache{}
	}
	s.literalTables[path] = &literalCache{snapshot, fingerprint}
}
func (s *Service) cachedLiteral(path string) (document.Table, bool) {
	if !fastEnabled() || s.current.Settings.Build.Mode != "manual" {
		return document.Table{}, false
	}
	cached := s.literalTables[path]
	if cached == nil {
		return document.Table{}, false
	}
	fingerprint, err := s.sourceFingerprint()
	if err != nil || fingerprint != cached.fingerprint {
		delete(s.literalTables, path)
		return document.Table{}, false
	}
	table := cached.snapshot.Table
	table.Columns = append([]document.Column(nil), table.Columns...)
	return table, true
}

func (s *Service) tryFastLiteral(update CellUpdate) (CellUpdateResult, bool, error) {
	if s.sourceUncertain {
		return CellUpdateResult{}, true, fmt.Errorf("source state uncertain; reopen project before editing")
	}
	if !fastEnabled() || s.current.Settings.Build.Mode != "manual" {
		return CellUpdateResult{}, false, nil
	}
	cache := s.literalTables[update.Path]
	if cache == nil {
		return CellUpdateResult{}, false, nil
	}
	fingerprint, err := s.sourceFingerprint()
	if err != nil {
		return CellUpdateResult{}, true, err
	}
	if fingerprint != cache.fingerprint {
		delete(s.literalTables, update.Path)
		return CellUpdateResult{}, true, fmt.Errorf("project sources changed externally; reload before editing")
	}
	patch, err := cache.snapshot.Patch(update.Path, update.RowID, update.Column, update.Expression, false)
	if err != nil {
		return CellUpdateResult{}, false, nil
	}
	baseRevision := s.revision
	beforeSHA := digest(cache.snapshot.Source)
	table, view, err := s.writeLiteralPatch(update.Path, cache, patch)
	if err != nil {
		return CellUpdateResult{}, true, err
	}
	s.recordHistory(historyEntry{Cell: &cellHistory{update.Path, update.RowID, update.Column, patch.Before, patch.After, beforeSHA, digest(patch.Source)}})
	view.CanUndo, view.CanRedo = true, false
	s.current = &view
	delta := makeCellDelta(table, view, baseRevision, update.Path, update.RowID, update.Column)
	return CellUpdateResult{Table: table, Workspace: view, delta: delta}, true, nil
}

type literalWriteOps struct {
	refresh func() (string, error)
	restore func(string, []byte, os.FileMode) error
	stage   func(string, []byte, os.FileMode) (string, error)
	rename  func(string, string) error
}

func (s *Service) writeLiteralPatch(path string, cache *literalCache, patch document.LiteralPatch) (document.Table, View, error) {
	defer perftrace.Start("fast.writeLiteralPatch")()
	info, err := os.Lstat(path)
	if err != nil {
		return document.Table{}, View{}, err
	}
	if !info.Mode().IsRegular() {
		return document.Table{}, View{}, fmt.Errorf("fast patch requires regular source")
	}
	ops := literalWriteOps{stage: sourcefile.Stage, rename: os.Rename, restore: restoreSourceFile, refresh: s.sourceFingerprint}
	if s.literalOps != nil {
		if s.literalOps.stage != nil {
			ops.stage = s.literalOps.stage
		}
		if s.literalOps.rename != nil {
			ops.rename = s.literalOps.rename
		}
		if s.literalOps.refresh != nil {
			ops.refresh = s.literalOps.refresh
		}
		if s.literalOps.restore != nil {
			ops.restore = s.literalOps.restore
		}
	}
	staged, err := ops.stage(path, patch.Source, info.Mode())
	if err != nil {
		return document.Table{}, View{}, err
	}
	defer os.Remove(staged)
	// Check again after staging, before the same-filesystem atomic replacement.
	current, err := os.ReadFile(path)
	if err != nil {
		return document.Table{}, View{}, err
	}
	fingerprint, err := s.sourceFingerprint()
	if err != nil {
		return document.Table{}, View{}, err
	}
	if !bytes.Equal(current, cache.snapshot.Source) || fingerprint != cache.fingerprint {
		return document.Table{}, View{}, fmt.Errorf("source changed while staging; edit not applied")
	}
	if err := ops.rename(staged, path); err != nil {
		return document.Table{}, View{}, err
	}
	fingerprint, err = ops.refresh()
	if err != nil {
		rollback := ops.restore(path, cache.snapshot.Source, sourcePermissionMode(info.Mode()))
		if rollback != nil {
			s.markSourceUncertain()
			return document.Table{}, View{}, errors.Join(fmt.Errorf("source may already be saved; rollback failed; reopen project before editing: %w", err), rollback)
		}
		return document.Table{}, View{}, errors.Join(fmt.Errorf("cache refresh failed; edit rolled back: %w", err), rollback)
	}
	cache.snapshot.Commit(patch)
	cache.fingerprint = fingerprint
	s.literalInputFingerprint = fingerprint
	s.markEdited()
	view := *s.current
	s.publishView(&view)
	if view.Settings.Build.Mode == "auto" {
		s.compile()
		view = *s.current
	}
	table := cache.snapshot.Table
	table.Columns = append([]document.Column(nil), table.Columns...)
	return withTableMetadata(table, view), view, nil
}

func (s *Service) restoreCellHistory(undo bool, entry historyEntry) (HistoryResult, error) {
	h := entry.Cell
	cache := s.literalTables[h.Path]
	if cache == nil {
		table, err := document.LoadTable(h.Path)
		if err != nil {
			return HistoryResult{}, err
		}
		s.rememberLiteral(h.Path, table)
		cache = s.literalTables[h.Path]
	}
	if cache == nil {
		return HistoryResult{}, fmt.Errorf("literal history requires reloading the original fixed-value table")
	}
	expected, expression := h.AfterSHA, h.Before
	if !undo {
		expected, expression = h.BeforeSHA, h.After
	}
	raw, err := os.ReadFile(h.Path)
	if err != nil {
		return HistoryResult{}, err
	}
	if digest(raw) != expected {
		return HistoryResult{}, fmt.Errorf("source changed outside history; refusing to overwrite")
	}
	patch, err := cache.snapshot.Patch(h.Path, h.RowID, h.Column, expression, true)
	if err != nil {
		return HistoryResult{}, err
	}
	baseRevision := s.revision
	table, view, err := s.writeLiteralPatch(h.Path, cache, patch)
	if err != nil {
		return HistoryResult{}, err
	}
	if undo {
		s.undoStack = s.undoStack[:len(s.undoStack)-1]
		s.redoStack = append(s.redoStack, entry)
	} else {
		s.redoStack = s.redoStack[:len(s.redoStack)-1]
		s.undoStack = append(s.undoStack, entry)
	}
	view.CanUndo, view.CanRedo = len(s.undoStack) > 0, len(s.redoStack) > 0
	s.current = &view
	return HistoryResult{Workspace: view, CanUndo: view.CanUndo, CanRedo: view.CanRedo, delta: makeCellDelta(table, view, baseRevision, h.Path, h.RowID, h.Column)}, nil
}
