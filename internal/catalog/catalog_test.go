package catalog

import (
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

func TestLoad_EmbeddedCatalogIsValid(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error on embedded catalog: %v", err)
	}
	if len(c.Harnesses) == 0 {
		t.Fatal("embedded catalog has no harnesses")
	}
}

func TestLoad_KnownHarnessesPresent(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	for _, id := range []string{"openspec", "engram", "sdd-orchestrator", "jr-orchestrator", "kb-creator"} {
		if _, ok := c.ByID(id); !ok {
			t.Errorf("expected harness %q in catalog, not found", id)
		}
	}
}

func TestForMode_LiteIsSubsetOfFull(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	lite := c.ForMode(model.ModeLite)
	full := c.ForMode(model.ModeFull)

	if len(lite) == 0 {
		t.Fatal("lite mode has no harnesses")
	}
	if len(full) <= len(lite) {
		t.Errorf("expected full (%d) to have more harnesses than lite (%d)", len(full), len(lite))
	}

	inFull := make(map[string]bool)
	for _, h := range full {
		inFull[h.ID] = true
	}
	for _, h := range lite {
		if !inFull[h.ID] {
			t.Errorf("lite harness %q is not included in full mode", h.ID)
		}
	}
}

func TestForMode_CustomReturnsAll(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if got := len(c.ForMode(model.ModeCustom)); got != len(c.Harnesses) {
		t.Errorf("custom mode returned %d harnesses, want all %d", got, len(c.Harnesses))
	}
}

func TestForAgent_RespectsAgentScope(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	// sdd-orchestrator is scoped to claude+opencode, so gemini must NOT get it.
	for _, h := range c.ForAgent(model.AgentGemini) {
		if h.ID == "sdd-orchestrator" {
			t.Error("gemini should not receive claude/opencode-scoped sdd-orchestrator")
		}
	}
	// claude must get it.
	var claudeHasOrchestrator bool
	for _, h := range c.ForAgent(model.AgentClaude) {
		if h.ID == "sdd-orchestrator" {
			claudeHasOrchestrator = true
		}
	}
	if !claudeHasOrchestrator {
		t.Error("claude should receive sdd-orchestrator")
	}
}

func TestSkillHarnesses_HaveMethod(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	for _, h := range c.Harnesses {
		if h.Type != model.HarnessSkill {
			continue
		}
		if h.Source == nil {
			t.Errorf("skill harness %q has nil source", h.ID)
			continue
		}
		if h.Source.Method == "" {
			t.Errorf("skill harness %q has empty source.method after catalog load", h.ID)
		}
		switch h.Source.Method {
		case "clone", "npx", "embed":
			// valid
		default:
			t.Errorf("skill harness %q has unknown source.method %q", h.ID, h.Source.Method)
		}
	}
}

func TestMethodInference_ThirdParty(t *testing.T) {
	// A skill harness with third_party:true and no method should infer "npx".
	yaml := `harnesses:
  - id: x
    name: X
    type: skill
    third_party: true
    source: { repo: some/repo }
    install_modes: [full]`
	c, err := parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse(): %v", err)
	}
	h, _ := c.ByID("x")
	if h.Source.Method != "npx" {
		t.Errorf("expected method %q for third_party skill, got %q", "npx", h.Source.Method)
	}
}

func TestMethodInference_OwnSkill(t *testing.T) {
	// A skill harness without third_party and no method should infer "clone".
	yaml := `harnesses:
  - id: x
    name: X
    type: skill
    source: { repo: some/repo }
    install_modes: [full]`
	c, err := parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse(): %v", err)
	}
	h, _ := c.ByID("x")
	if h.Source.Method != "clone" {
		t.Errorf("expected method %q for own skill, got %q", "clone", h.Source.Method)
	}
}

func TestMethodInference_ExplicitOverride(t *testing.T) {
	// Explicit method must NOT be overwritten by inference.
	yaml := `harnesses:
  - id: x
    name: X
    type: skill
    source: { repo: some/repo, method: embed }
    install_modes: [full]`
	c, err := parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse(): %v", err)
	}
	h, _ := c.ByID("x")
	if h.Source.Method != "embed" {
		t.Errorf("expected explicit method %q to be preserved, got %q", "embed", h.Source.Method)
	}
}

func TestValidate_RejectsBadCatalogs(t *testing.T) {
	tests := map[string]struct {
		yaml string
		want string
	}{
		"duplicate id": {
			yaml: `harnesses:
  - {id: dup, name: A, type: config, install_modes: [lite]}
  - {id: dup, name: B, type: config, install_modes: [lite]}`,
			want: "duplicate",
		},
		"invalid type": {
			yaml: `harnesses:
  - {id: x, name: X, type: bogus, install_modes: [lite]}`,
			want: "invalid type",
		},
		"no install modes": {
			yaml: `harnesses:
  - {id: x, name: X, type: config, install_modes: []}`,
			want: "no install_modes",
		},
		"invalid mode": {
			yaml: `harnesses:
  - {id: x, name: X, type: config, install_modes: [turbo]}`,
			want: "invalid mode",
		},
		"skill without source": {
			yaml: `harnesses:
  - {id: x, name: X, type: skill, install_modes: [full]}`,
			want: "source.repo",
		},
		"external without method": {
			yaml: `harnesses:
  - {id: x, name: X, type: external, install_modes: [lite]}`,
			want: "external.method",
		},
		"skill with unknown method": {
			yaml: `harnesses:
  - id: x
    name: X
    type: skill
    source: { repo: some/repo, method: ftp }
    install_modes: [full]`,
			want: "unknown source.method",
		},
		"unknown dependency": {
			yaml: `harnesses:
  - {id: x, name: X, type: config, install_modes: [lite], depends_on: [ghost]}`,
			want: "unknown harness",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parse([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}
