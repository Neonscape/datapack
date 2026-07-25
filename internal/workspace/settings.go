package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const defaultAutoCompileDebounceSeconds = 10

type EditorSettings struct {
	Version int           `json:"version"`
	Build   BuildSettings `json:"build"`
}

type BuildSettings struct {
	Mode                       string `json:"mode"`
	AutoCompileDebounceSeconds int    `json:"autoCompileDebounceSeconds,omitempty"`
}

func DefaultEditorSettings() EditorSettings {
	return EditorSettings{
		Version: 1,
		Build: BuildSettings{
			Mode:                       "auto",
			AutoCompileDebounceSeconds: defaultAutoCompileDebounceSeconds,
		},
	}
}

func SettingsPath(configRoot string) string {
	return filepath.Join(configRoot, "datapack", "settings.json")
}

func LoadEditorSettings(configRoot string) (EditorSettings, error) {
	path := SettingsPath(configRoot)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultEditorSettings(), nil
	}
	if err != nil {
		return EditorSettings{}, fmt.Errorf("open editor settings: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var settings EditorSettings
	if err := decoder.Decode(&settings); err != nil {
		return EditorSettings{}, fmt.Errorf("decode editor settings: %w", err)
	}
	if err := validateEditorSettings(settings); err != nil {
		return EditorSettings{}, err
	}
	return settings, nil
}

func validateEditorSettings(settings EditorSettings) error {
	if settings.Version != 1 {
		return fmt.Errorf("unsupported editor settings version %d", settings.Version)
	}
	switch settings.Build.Mode {
	case "auto":
		if settings.Build.AutoCompileDebounceSeconds <= 0 {
			return fmt.Errorf("auto compile debounce must be greater than zero")
		}
	case "manual":
		return nil
	default:
		return fmt.Errorf("unsupported build mode %q", settings.Build.Mode)
	}
	return nil
}
