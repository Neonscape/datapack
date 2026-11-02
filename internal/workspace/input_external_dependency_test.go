package workspace

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestTransitiveInputsRejectUnsupportedExternalAttributes(t *testing.T) {
	for _, attribute := range []string{`@embed("data.json")`, `@extern("javascript")`} {
		t.Run(attribute, func(t *testing.T) {
			s, root, _ := dependencyFixture(t, true, false)
			if _, err := s.Export(); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(root, "shared", "base", "base.cue"), "package base\nlimit: 20 "+attribute+"\n")
			_, err := s.captureInputs()
			if err == nil || !strings.Contains(err.Error(), "cannot establish complete external input dependencies") {
				t.Fatalf("transitive marker was not rejected by complete-input policy: %v", err)
			}
			if _, err := s.Export(); err == nil {
				t.Fatal("unverifiable transitive external input exported")
			}
		})
	}
}
