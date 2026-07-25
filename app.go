package main

import (
	"context"
	"fmt"

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
