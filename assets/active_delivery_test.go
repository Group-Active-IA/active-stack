package assets

import (
	"io/fs"
	"strings"
	"testing"
)

func TestActiveDeliveryWorkflowContract(t *testing.T) {
	skill, err := fs.ReadFile(SkillsFS, "skills/active-delivery/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := fs.ReadFile(SkillsFS, "skills/active-delivery/references/handoff.md")
	if err != nil {
		t.Fatal(err)
	}
	content := string(skill)
	for _, heading := range []string{"## Choose the delivery mode", "## Continue or resume ODD", "## Promote ODD to SDD"} {
		if !strings.Contains(content, heading) {
			t.Fatalf("workflow missing %q", heading)
		}
	}
	choose := strings.SplitN(strings.SplitN(content, "## Choose the delivery mode", 2)[1], "## Continue or resume ODD", 2)[0]
	resume := strings.SplitN(strings.SplitN(content, "## Continue or resume ODD", 2)[1], "## Promote ODD to SDD", 2)[0]
	promote := strings.SplitN(content, "## Promote ODD to SDD", 2)[1]
	cases := []struct {
		name, source string
		expected     []string
	}{
		{"route selection", choose, []string{"ODD", "SDD", "explicit", "governance"}},
		{"workspace readiness", choose, []string{"CLI is unavailable", "ask for confirmation", "openspec init --tools codex", "Only after the workspace is ready"}},
		{"ODD resume", resume, []string{"OpenSpec", "ID"}},
		{"handoff review", promote, []string{"confirm", "workspace readiness", "openspec list --json", "openspec status --change", "existing"}},
		{"recoverable evidence", string(handoff), []string{"knowledge-base/entrega/<id>.md", "git rev-parse HEAD", "Engram", "decision", "verification"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			for _, expected := range tt.expected {
				if !strings.Contains(strings.ToLower(tt.source), strings.ToLower(expected)) {
					t.Errorf("missing %q in installed workflow contract", expected)
				}
			}
		})
	}
}
