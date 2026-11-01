package workspace

import "datapack/internal/document"

type DisplayNameSet struct {
	Path        string `json:"path"`
	Definition  string `json:"definition,omitempty"`
	DisplayName string `json:"displayName"`
}

// SetDisplayName preserves file identity and uses the same edit/history protocol
// as other source-backed editor metadata.
func (s *Service) SetDisplayName(update DisplayNameSet) (SchemaEditResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return SchemaEditResult{}, s.userError("workspace.noProject")
	}
	scope, err := s.schemaDefinitionScope(update.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	if scope == "local" {
		table, view, err := s.edit(update.Path, func() (document.Table, error) {
			if err := document.SetDisplayName(update.Path, update.Definition, update.DisplayName); err != nil {
				return document.Table{}, err
			}
			return document.LoadTable(update.Path)
		})
		if err != nil {
			return SchemaEditResult{}, err
		}
		return SchemaEditResult{Table: withTableMetadata(table, view), Workspace: view}, nil
	}
	schema, view, err := editSource(s, update.Path, func() (document.PublicSchema, error) {
		if err := document.SetDisplayName(update.Path, update.Definition, update.DisplayName); err != nil {
			return document.PublicSchema{}, err
		}
		return document.LoadPublicSchema(update.Path)
	})
	if err != nil {
		return SchemaEditResult{}, err
	}
	return SchemaEditResult{Schema: schema, Workspace: view}, nil
}
