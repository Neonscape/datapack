package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManualSettingsRejectInvalidLocaleOrThemeWithoutReplacingFile(t *testing.T) {
	for _, field := range []string{"language", "theme"} {
		t.Run(field, func(t *testing.T) {
			root := t.TempDir()
			valid := DefaultEditorSettings()
			valid.Build = BuildSettings{Mode: "manual"}
			if err := SaveEditorSettings(root, valid); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(SettingsPath(root))
			if err != nil {
				t.Fatal(err)
			}
			invalid := valid
			if field == "language" {
				invalid.Language = "unsupported"
			} else {
				invalid.Theme = "unsupported"
			}
			if err := SaveEditorSettings(root, invalid); err == nil {
				t.Error("invalid manual settings accepted")
			}
			after, _ := os.ReadFile(SettingsPath(root))
			if string(after) != string(before) {
				t.Error("rejected save changed the settings file")
			}
			raw := `{"version":1,"build":{"mode":"manual"},"language":"en","theme":"unsupported"}`
			if field == "language" {
				raw = `{"version":1,"build":{"mode":"manual"},"language":"unsupported","theme":"light"}`
			}
			if err := os.WriteFile(SettingsPath(root), []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadEditorSettings(root); err == nil {
				t.Error("invalid manual settings loaded")
			}
		})
	}
}
func TestManualLegacySettingsKeepDefaultsAndZeroDebounce(t *testing.T) {
	root := t.TempDir()
	path := SettingsPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"build":{"mode":"manual"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	value, err := LoadEditorSettings(root)
	if err != nil {
		t.Fatal(err)
	}
	if value.Language != DefaultLanguage || value.Theme != DefaultTheme || value.Build.AutoCompileDebounceSeconds != 0 {
		t.Fatalf("legacy defaults: %#v", value)
	}
}
