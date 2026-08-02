package main

import (
	"context"
	"fmt"

	"datapack/internal/document"
	"datapack/internal/workspace"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	workspace *workspace.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	service, err := workspace.NewUserService()
	if err != nil {
		panic(fmt.Sprintf("create workspace service: %v", err))
	}
	return &App{workspace: service}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectProject lets the user choose a data project and returns its source tree.
func (a *App) SelectProject() (workspace.View, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Datapack Project",
		Filters: []runtime.FileFilter{{
			DisplayName: "Datapack Project (*.dtproj)",
			Pattern:     "*.dtproj",
		}},
	})
	if err != nil {
		return workspace.View{}, fmt.Errorf("select project: %w", err)
	}
	if path == "" {
		return workspace.View{}, nil
	}
	return a.workspace.OpenProject(path)
}

// OpenProject opens a project by path. It is also used by tests and future
// recent-project entries that already know the path.
func (a *App) OpenProject(path string) (workspace.View, error) {
	return a.workspace.OpenProject(path)
}

func (a *App) CurrentWorkspace() (workspace.View, bool) {
	return a.workspace.Current()
}

// GetTable returns the structured, read-only representation of a table in the
// currently open project.
func (a *App) GetTable(path string) (document.Table, error) {
	return a.workspace.GetTable(path)
}

func (a *App) CreateTable(create workspace.TableCreate) (workspace.TableCreateResult, error) {
	return a.workspace.CreateTable(create)
}

// UpdateCell writes a valid CUE expression to an existing table cell.
func (a *App) UpdateCell(update workspace.CellUpdate) (workspace.CellUpdateResult, error) {
	return a.workspace.UpdateCell(update)
}

// SetCellFormula assigns an expr-lang/expr formula to one table cell.
func (a *App) SetCellFormula(set workspace.FormulaSet) (workspace.FormulaSetResult, error) {
	return a.workspace.SetCellFormula(set)
}

func (a *App) CreateRow(create workspace.RowCreate) (workspace.RowCreateResult, error) {
	return a.workspace.CreateRow(create)
}

func (a *App) ClearCell(update workspace.CellUpdate) (workspace.CellUpdateResult, error) {
	return a.workspace.ClearCell(update)
}

func (a *App) Paste(paste workspace.Paste) (workspace.HistoryResult, error) {
	return a.workspace.Paste(paste)
}

func (a *App) Undo(path string) (workspace.HistoryResult, error) {
	return a.workspace.Undo(path)
}

func (a *App) Redo(path string) (workspace.HistoryResult, error) {
	return a.workspace.Redo(path)
}

// AddRow inserts a new table row using a complete CUE struct expression.
func (a *App) AddRow(add workspace.RowAdd) (workspace.RowAddResult, error) {
	return a.workspace.AddRow(add)
}

// DeleteRow removes an existing table row.
func (a *App) DeleteRow(delete workspace.RowDelete) (workspace.RowDeleteResult, error) {
	return a.workspace.DeleteRow(delete)
}

// RenameRow changes a row ID and updates all declared foreign-key references.
func (a *App) RenameRow(rename workspace.RowRename) (workspace.RowRenameResult, error) {
	return a.workspace.RenameRow(rename)
}

// Export writes the current project's configured CSV output.
func (a *App) Export() (workspace.ExportResult, error) {
	return a.workspace.Export()
}

// AddColumn adds a column and initializes its value for every existing row.
func (a *App) AddColumn(add workspace.ColumnAdd) (workspace.ColumnAddResult, error) {
	return a.workspace.AddColumn(add)
}

func (a *App) SetColumnSchema(set workspace.ColumnSchemaSet) (workspace.ColumnSchemaSetResult, error) {
	return a.workspace.SetColumnSchema(set)
}

// DeleteColumn removes a column and its associated row values and metadata.
func (a *App) DeleteColumn(delete workspace.ColumnDelete) (workspace.ColumnDeleteResult, error) {
	return a.workspace.DeleteColumn(delete)
}

// RenameColumn changes a column and updates its same-table metadata.
func (a *App) RenameColumn(rename workspace.ColumnRename) (workspace.ColumnRenameResult, error) {
	return a.workspace.RenameColumn(rename)
}

// MoveColumn changes a column's position in the table Schema.
func (a *App) MoveColumn(move workspace.ColumnMove) (workspace.ColumnMoveResult, error) {
	return a.workspace.MoveColumn(move)
}
