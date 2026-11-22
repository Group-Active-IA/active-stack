package catalog

import (
	"regexp"
	"testing"
)

// [EQ] Fase 7 (R-20) T-F7-3: los `ref` de los tres repos externos propios son
// SHA completos de 40 caracteres y coinciden con el merge aprobado de cada uno.
var fullSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

func refOf(t *testing.T, id string) (repo, ref string) {
	t.Helper()
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	h, ok := c.ByID(id)
	if !ok {
		t.Fatalf("harness %q missing", id)
	}
	if h.Source == nil {
		t.Fatalf("harness %q has no source", id)
	}
	return h.Source.Repo, h.Source.Ref
}

func TestExternalRefsAreFullSHAs(t *testing.T) {
	for _, id := range []string{"active-orchestrator", "roadmap-generator", "agent-instruction"} {
		t.Run(id, func(t *testing.T) {
			_, ref := refOf(t, id)
			if !fullSHA.MatchString(ref) {
				t.Fatalf("ref %q is not a full 40-char lowercase SHA", ref)
			}
		})
	}
}

func TestActiveOrchestratorRefIsMemoryPolicyMerge(t *testing.T) {
	repo, ref := refOf(t, "active-orchestrator")
	if repo != "Group-Active-IA/active-orchestrator" || ref != "5d758fceb3b560b934c3348eaa20e944f79503a6" {
		t.Fatalf("got %s@%s", repo, ref)
	}
}
