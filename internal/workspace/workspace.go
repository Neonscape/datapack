package workspace

import (
	"fmt"
	"os"

	"datapack/internal/projecttree"
)

type View struct {
	Project  Project          `json:"project"`
	Settings EditorSettings   `json:"settings"`
	Tree     projecttree.Tree `json:"tree"`
}

type Service struct {
	configRoot string
	current    *View
}

func NewService(configRoot string) *Service {
	return &Service{configRoot: configRoot}
}

func NewUserService() (*Service, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	return NewService(configRoot), nil
}

func (s *Service) OpenProject(path string) (View, error) {
	project, err := LoadProject(path)
	if err != nil {
		return View{}, err
	}
	settings, err := LoadEditorSettings(s.configRoot)
	if err != nil {
		return View{}, err
	}
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		return View{}, err
	}
	view := View{Project: project, Settings: settings, Tree: tree}
	s.current = &view
	return view, nil
}

func (s *Service) Current() (View, bool) {
	if s.current == nil {
		return View{}, false
	}
	return *s.current, true
}
