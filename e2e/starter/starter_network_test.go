//go:build e2e_network

package starter

// C-32: Network-gated E2E arm — tasks 9.1, 9.2, 9.3
//
// IMPORTANT: This file has `//go:build e2e_network` as its FIRST non-blank line.
// It is NEVER compiled by `go test ./...` and is NEVER run in default CI.
// It runs ONLY when explicitly opted in:
//   go test -tags e2e_network ./e2e/starter/...
//
// Its purpose is drift-detection: clone the real upstream repos and verify that
// the same FS post-conditions hold. Degraded real third-party skills are reported
// (not required to install successfully).
//
// Design note (D1): a build tag is the Go-idiomatic gate for "expensive/external,
// off by default". Unlike -short or env-var gates, the code doesn't even compile
// into the default binary, so it cannot run by accident.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/agents/claude"
	"github.com/JuanCruzRobledo/jr-stack/internal/catalog"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
)

// ─────────────────────────────────────────────────────────────────
// Task 9.2/9.3: Network test — clone real upstream, run starter add,
// assert FS post-conditions.
// ─────────────────────────────────────────────────────────────────

// TestNetwork_StarterBase_RealUpstream clones the real upstream skill repo
// (JuanCruzRobledo/jr-skills) and runs `starter add base`, asserting FS
// post-conditions against real content.
//
// Degraded real third-party skills (the 6 stale skills) are reported but not
// required to install successfully (C-32 scope: make degradation visible,
// not heal it).
func TestNetwork_StarterBase_RealUpstream(t *testing.T) {
	requireGit(t)

	// Load the real embedded catalog — same as production.
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	// Resolve the "base" starter to get its harness set.
	harnesses, err := cat.ResolveStarter("base")
	if err != nil {
		t.Fatalf("ResolveStarter(\"base\") error = %v", err)
	}

	// Collect skill-type harness IDs.
	var skillIDs []string
	for _, h := range harnesses {
		if h.Type == model.HarnessSkill {
			skillIDs = append(skillIDs, h.ID)
		}
	}
	if len(skillIDs) == 0 {
		t.Skip("no skill harnesses in starter 'base'")
	}

	reg := &networkRegistry{adapter: claude.NewAdapter()}
	homeDir := t.TempDir()

	var out bytes.Buffer
	exitCode := headless.RunHeadless(headless.ParsedFlags{
		Yes:           true,
		HomeDir:       homeDir,
		NoSelfInstall: true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeCustom,
			Custom: skillIDs,
		},
	}, cat, reg, &out)

	// The run must exit 0 even if some best-effort harnesses degrade.
	if exitCode != 0 {
		t.Errorf("network run exited %d; output:\n%s", exitCode, out.String())
	}

	t.Logf("network run output:\n%s", out.String())

	output := out.String()
	skillsDir := claudeSkillsDir(homeDir)

	// Assert FS post-conditions for each skill harness.
	// Harnesses marked as best-effort may degrade — log them but don't fail.
	for _, h := range harnesses {
		if h.Type != model.HarnessSkill {
			continue
		}
		if h.BestEffort {
			// Degraded real third-party skills are reported, not required to install.
			// The output should mention their degraded state if they failed.
			if strings.Contains(output, h.ID) && strings.Contains(output, "Degraded") {
				t.Logf("DEGRADED (expected): %s — third-party skill degraded on real upstream", h.ID)
			}
			continue
		}
		// Non-best-effort skills: assert SKILL.md exists.
		assertSKILLmdExists(t, skillsDir, h.ID)
	}
}

// networkRegistry satisfies install.Registry using the real claude adapter.
type networkRegistry struct {
	adapter install.AgentAdapter
}

func (r *networkRegistry) Get(a model.Agent) (install.AgentAdapter, bool) {
	if a == model.AgentClaude {
		return r.adapter, true
	}
	return nil, false
}
