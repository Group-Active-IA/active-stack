package config

import (
	"strings"
	"testing"
)

func TestCodexDocumentationDiagramGuidance(t *testing.T) {
	for _, tc := range []struct {
		name    string
		toggles []string
	}{
		{name: "default"},
		{name: "governed", toggles: []string{"governance"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pkg, err := ComposeCodexPackage(tc.toggles)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"Mermaid", "roadmap", "architecture", "markdown_description", "exact saved", "confirmation"} {
				if !strings.Contains(pkg.Universal, want) {
					t.Errorf("installed Codex instructions lack %q", want)
				}
			}
		})
	}
}
