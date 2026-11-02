package workspace

import "datapack/internal/document"

// CellDelta is returned only by the independent-literal source-patch path,
// after persistence and history publication. Other edits retain the full DTO.
type CellDelta struct {
	Path         string        `json:"path"`
	ProjectEpoch string        `json:"projectEpoch"`
	BaseRevision int           `json:"baseRevision"`
	Revision     int           `json:"revision"`
	RowIndex     int           `json:"rowIndex"`
	RowID        string        `json:"rowID"`
	Column       string        `json:"column"`
	Cell         document.Cell `json:"cell"`
}
type CellEditResult struct {
	Table     *document.Table `json:"table,omitempty"`
	Delta     *CellDelta      `json:"delta,omitempty"`
	Workspace View            `json:"workspace"`
}

func compactCellResult(result CellUpdateResult, err error) (CellEditResult, error) {
	if err != nil {
		return CellEditResult{}, err
	}
	if result.delta != nil {
		return CellEditResult{Delta: result.delta, Workspace: result.Workspace}, nil
	}
	return CellEditResult{Table: &result.Table, Workspace: result.Workspace}, nil
}
func (s *Service) UpdateCellDelta(update CellUpdate) (CellEditResult, error) {
	return compactCellResult(s.UpdateCell(update))
}
func (s *Service) SetCellTextDelta(update CellUpdate) (CellEditResult, error) {
	return compactCellResult(s.SetCellText(update))
}

func makeCellDelta(table document.Table, view View, base int, path, rowID, column string) *CellDelta {
	for index, row := range table.Rows {
		if row.ID == rowID {
			return &CellDelta{Path: path, ProjectEpoch: view.ProjectEpoch, BaseRevision: base, Revision: view.Revision, RowIndex: index, RowID: rowID, Column: column, Cell: row.Values[column]}
		}
	}
	return nil
}
func compactHistoryResult(result HistoryResult, err error) (CellEditResult, error) {
	if err != nil {
		return CellEditResult{}, err
	}
	return CellEditResult{Delta: result.delta, Workspace: result.Workspace}, nil
}
func (s *Service) UndoDelta() (CellEditResult, error) { return compactHistoryResult(s.Undo()) }
func (s *Service) RedoDelta() (CellEditResult, error) { return compactHistoryResult(s.Redo()) }
