package assets

import (
	"io/fs"
	"strings"
	"testing"
)

func TestActiveDeliveryRoadmapReconciliation(t *testing.T) {
	paths := []string{
		"skills/active-delivery/SKILL.md",
		"skills/active-delivery/references/roadmap-reconciliation.md",
	}
	contents := make(map[string]string, len(paths))
	for _, path := range paths {
		data, err := fs.ReadFile(SkillsFS, path)
		if err != nil {
			t.Error(err)
		}
		contents[path] = string(data)
	}
	skill := contents[paths[0]]
	reference := contents[paths[1]]
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{"lifecycle hooks for either mode", skill, []string{"references/roadmap-reconciliation.md", "resume, delivery and closure", "ODD and SDD"}},
		{"local default has no remote lookup", reference, []string{"local by default", ".active-stack/roadmap.json", "no remote lookup"}},
		{"authorities and persisted observation", reference, []string{"CHANGES.md", "scope and dependencies", "status and assignee", ".active-stack/roadmap-state.json", "last confirmed status", "link", "timestamp"}},
		{"offline retains evidence and marks pending", reference, []string{"preserve the last confirmed", "pending", "continue local work"}},
		{"preview and confirmation required for remote writes", reference, []string{"read-only", "preview", "explicit confirmation", "active-roadmap"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, expected := range tc.want {
				if !strings.Contains(tc.source, expected) {
					t.Errorf("expected %q in the embedded contract", expected)
				}
			}
		})
	}
}
