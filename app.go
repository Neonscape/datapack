package main

import (
	"context"
	"fmt"
	"os"

	"datapack/internal/document"
	"datapack/internal/logging"
	"datapack/internal/navigation"
	"datapack/internal/schemamodel"
	"datapack/internal/workspace"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	workspace *workspace.Service
	logger    *logging.Logger
	logPath   string
}

// NewApp creates a new App application struct
func NewApp() *App {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Sprintf("resolve user config directory: %v", err))
	}
	logger, logPath, err := logging.NewUserLogger(configRoot)
	if err != nil {
		panic(fmt.Sprintf("create application logger: %v", err))
	}
	service, err := workspace.NewUserService()
	if err != nil {
		panic(fmt.Sprintf("create workspace service: %v", err))
	}
	service.SetLogger(logger)
	logger.Info("application created", map[string]any{"log_path": logPath})
	return &App{workspace: service, logger: logger, logPath: logPath}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.AddSink(func(entry logging.Entry) { runtime.EventsEmit(ctx, "log:entry", entry) })
	a.logger.Info("application started", nil)
}

func (a *App) LogPath() string { return a.logPath }

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

// GetPublicSchema returns the structured representation of one public Schema
// file in the currently open project.
func (a *App) GetPublicSchema(path string) (document.PublicSchema, error) {
	return a.workspace.GetPublicSchema(path)
}

func (a *App) CreateSchemaDefinition(create workspace.SchemaDefinitionCreate) (workspace.SchemaEditResult, error) {
	return a.workspace.CreateSchemaDefinition(create)
}

func (a *App) SetSchemaDefinition(set workspace.SchemaDefinitionSet) (workspace.SchemaEditResult, error) {
	return a.workspace.SetSchemaDefinition(set)
}

func (a *App) RenameSchemaDefinition(rename workspace.SchemaDefinitionRename) (workspace.SchemaEditResult, error) {
	return a.workspace.RenameSchemaDefinition(rename)
}

func (a *App) DeleteSchemaDefinition(delete workspace.SchemaDefinitionDelete) (workspace.SchemaEditResult, error) {
	return a.workspace.DeleteSchemaDefinition(delete)
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

// Compile runs one full compile and returns the refreshed workspace view.
func (a *App) Compile() (workspace.CompileResult, error) {
	return a.workspace.Compile()
}

// SetCompileMode switches between auto and manual compilation and persists
// the user-level editor settings.
func (a *App) SetCompileMode(mode string) (workspace.View, error) {
	return a.workspace.SetCompileMode(mode)
}

// Navigate resolves one semantic jump to a document and location.
func (a *App) Navigate(request navigation.Request) (navigation.Target, error) {
	return a.workspace.Navigate(request)
}

// PrintSchema serializes a structured Schema node back into CUE.
func (a *App) PrintSchema(node schemamodel.Node) (string, error) {
	return schemamodel.Print(node)
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
