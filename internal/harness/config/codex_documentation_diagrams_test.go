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
			for _, want := range []string{"Mermaid", "roadmap", "architecture", "discovery", "one option per diagram", "wait for the selection", "markdown_description", "exact saved", "confirmation"} {
				if !strings.Contains(pkg.Universal, want) {
					t.Errorf("installed Codex instructions lack %q", want)
				}
			}
		})
	}
}

func TestCrossClientDocumentationDiagramGuidance(t *testing.T) {
	for _, client := range []string{"opencode", "antigravity"} {
		t.Run(client, func(t *testing.T) {
			instructions, err := Compose(nil, client)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(instructions, "one option per diagram") {
				t.Fatal("installed instructions lack diagram selection")
			}
		})
	}
}

func TestClaudeDocumentationDiagramGuidance(t *testing.T) {
	layers, err := ComposeLayers(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"AskUserQuestion", "one option per diagram", "selected exact block"} {
		if !strings.Contains(layers.Universal, want) {
			t.Errorf("installed Claude instructions lack %q", want)
		}
	}
}
