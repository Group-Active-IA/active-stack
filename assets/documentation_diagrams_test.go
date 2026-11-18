package assets

import (
	"io/fs"
	"strings"
	"testing"
)

func TestDocumentationDiagramEntryPoints(t *testing.T) {
	for _, tc := range []struct {
		skill   string
		heading string
		want    []string
	}{
		{"active-delivery", "Documentation diagrams", []string{"fenced Mermaid", "architecture", "saved block", "source is unavailable", "discovery", "one option per diagram", "OpenCode", "wait for the selection"}},
		{"active-roadmap", "Diagramas de documentación", []string{"bloque Mermaid editable", "markdown_description", "text_content", "bloque guardado exacto", "vista previa y confirmación", "una opción por diagrama", "esperá la selección"}},
	} {
		t.Run(tc.skill, func(t *testing.T) {
			data, err := fs.ReadFile(SkillsFS, "skills/"+tc.skill+"/SKILL.md")
			if err != nil {
				t.Fatal(err)
			}
			_, section, found := strings.Cut(string(data), "## "+tc.heading+"\n")
			if !found {
				t.Fatal("embedded skill lacks diagram guidance")
			}
			for _, want := range tc.want {
				if !strings.Contains(section, want) {
					t.Errorf("diagram guidance lacks %q", want)
				}
			}
		})
	}
}
