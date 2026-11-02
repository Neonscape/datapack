package workspace

import (
	"datapack/internal/compiler"
	"datapack/internal/projecttree"
	"fmt"
)

func (s *Service) refreshExternalCompileInputs() error {
	if s.literalInputFingerprint == "" {
		return nil
	}
	fingerprint, err := s.sourceFingerprint()
	if err == nil && fingerprint == s.literalInputFingerprint {
		return nil
	}
	project, err := LoadProject(s.current.Project.Path)
	if err != nil {
		return fmt.Errorf("reload changed project: %w", err)
	}
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		return fmt.Errorf("rescan changed sources: %w", err)
	}
	view := buildView(project, s.current.Settings, tree)
	s.markEdited()
	s.publishView(&view)
	s.literalTables = nil
	s.literalInputFingerprint, err = s.sourceFingerprint()
	return err
}

// ResearchHistoryStats counts retained payload, not Go object overhead/RSS.
func (s *Service) ResearchHistoryStats() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := map[string]int64{"undo_entries": int64(len(s.undoStack)), "redo_entries": int64(len(s.redoStack))}
	for _, stack := range [][]historyEntry{s.undoStack, s.redoStack} {
		for _, entry := range stack {
			for _, source := range entry.Files {
				result["retained_full_source_bytes"] += int64(len(source))
			}
			if entry.Cell != nil {
				result["cell_patch_entries"]++
				result["cell_expression_bytes"] += int64(len(entry.Cell.Before) + len(entry.Cell.After))
				result["cell_digest_bytes"] += int64(len(entry.Cell.BeforeSHA) + len(entry.Cell.AfterSHA))
			}
		}
	}
	return result
}

type successfulCompile struct {
	fingerprint string
	result      compiler.Result
}

// Imported dependencies outside project roots are intentionally not cached by
// this prototype. The conservative keyword check may reject harmless comments.
func (s *Service) compileFingerprint() (string, bool) {
	snapshot, err := s.captureInputs()
	return snapshot.fingerprint, err == nil && snapshot.cacheable
}
