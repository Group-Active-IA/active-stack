package catalog

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// C-19: best-effort harness flag
// ─────────────────────────────────────────────────────────────────────────────

// TestBestEffortFlag_FindSkillAndSkillCreatorAreBestEffort asserts that the
// embedded catalog parses find-skill and skill-creator with BestEffort == true,
// while all other harnesses default to BestEffort == false.
func TestBestEffortFlag_FindSkillAndSkillCreatorAreBestEffort(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	bestEffortIDs := []string{"find-skill", "skill-creator"}
	for _, id := range bestEffortIDs {
		h, ok := c.ByID(id)
		if !ok {
			t.Errorf("harness %q not found in catalog", id)
			continue
		}
		if !h.BestEffort {
			t.Errorf("harness %q: BestEffort = false, want true", id)
		}
	}
}

// TestBestEffortFlag_NonBestEffortHarnessesDefaultToFalse asserts that harnesses
// without best_effort set default to BestEffort == false.
func TestBestEffortFlag_NonBestEffortHarnessesDefaultToFalse(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	nonBestEffortIDs := []string{"openspec", "engram", "sdd-orchestrator", "active-orchestrator", "permissions"}
	for _, id := range nonBestEffortIDs {
		h, ok := c.ByID(id)
		if !ok {
			t.Errorf("harness %q not found in catalog", id)
			continue
		}
		if h.BestEffort {
			t.Errorf("harness %q: BestEffort = true, want false (should default to false)", id)
		}
	}
}

// TestBestEffortFlag_ParseFromYAML asserts that catalog.parse() correctly reads
// best_effort: true from raw YAML, and that absence of the field defaults to false.
func TestBestEffortFlag_ParseFromYAML(t *testing.T) {
	yaml := `harnesses:
  - id: be-skill
    name: Best Effort
    type: skill
    best_effort: true
    source: { repo: some/repo, method: clone }
    install_modes: [full]
  - id: normal-skill
    name: Normal Skill
    type: skill
    source: { repo: some/repo, method: clone }
    install_modes: [full]`

	c, err := parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	beH, ok := c.ByID("be-skill")
	if !ok {
		t.Fatal("be-skill not found")
	}
	if !beH.BestEffort {
		t.Errorf("be-skill: BestEffort = false, want true")
	}

	normalH, ok := c.ByID("normal-skill")
	if !ok {
		t.Fatal("normal-skill not found")
	}
	if normalH.BestEffort {
		t.Errorf("normal-skill: BestEffort = true, want false (no best_effort in YAML)")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// C-22: find-skill / skill-creator install via clone from a repo subdir
// ─────────────────────────────────────────────────────────────────────────────

// TestC22_ThirdPartySkillsUseCloneWithPath asserts that find-skill and
// skill-creator abandoned npx in favor of clone + a non-empty Source.Path
// (the subdir inside the upstream monorepo where their SKILL.md lives).
func TestC22_ThirdPartySkillsUseCloneWithPath(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, id := range []string{"find-skill", "skill-creator"} {
		h, ok := c.ByID(id)
		if !ok {
			t.Errorf("harness %q not found in catalog", id)
			continue
		}
		if h.Source == nil {
			t.Errorf("harness %q has nil source", id)
			continue
		}
		if h.Source.Method != "clone" {
			t.Errorf("harness %q: Source.Method = %q, want %q", id, h.Source.Method, "clone")
		}
		if h.Source.Path == "" {
			t.Errorf("harness %q: Source.Path is empty, want a subdir path", id)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Part 1 — Engram brew tap formula
// ─────────────────────────────────────────────────────────────────────────────

// TestEngram_BrewTapFormula asserts that the embedded catalog specifies the
// full Homebrew TAP path for engram ("gentleman-programming/tap/engram"),
// NOT the bare "engram" that does not exist in homebrew-core.
// filepath.Base of the tap path must still resolve to "engram" (the binary name).
func TestEngram_BrewTapFormula(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	h, ok := c.ByID("engram")
	if !ok {
		t.Fatal("engram harness not found in catalog")
	}
	if h.External == nil {
		t.Fatal("engram: External is nil")
	}
	const wantPkg = "gentleman-programming/tap/engram"
	if h.External.Pkg != wantPkg {
		t.Errorf("engram External.Pkg = %q, want %q (bare 'engram' is not in homebrew-core; the tap formula is required)", h.External.Pkg, wantPkg)
	}
	// Binary name must still be "engram" (filepath.Base of the tap path).
	// This is what runBrew uses to resolve the installed binary after brew install.
	if got := filepath.Base(h.External.Pkg); got != "engram" {
		t.Errorf("filepath.Base(Pkg) = %q, want %q", got, "engram")
	}
}

// TestEngram_HasMCPEntry asserts that the embedded catalog engram harness
// declares an MCP stdio entry with name="engram", command="engram", args=["mcp"].
// This is required so active-stack can register the Engram MCP server into each
// agent's config after installing the binary — without this, the binary is
// on disk but the agent cannot communicate with it.
func TestEngram_HasMCPEntry(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	h, ok := c.ByID("engram")
	if !ok {
		t.Fatal("engram harness not found in catalog")
	}
	if h.External == nil {
		t.Fatal("engram: External is nil")
	}
	if h.External.MCP == nil {
		t.Fatal("engram External.MCP is nil; engram requires an MCP stdio entry to be registered into agent configs")
	}
	m := h.External.MCP
	if m.Name != "engram" {
		t.Errorf("MCP.Name = %q, want %q", m.Name, "engram")
	}
	if m.Command != "engram" {
		t.Errorf("MCP.Command = %q, want %q", m.Command, "engram")
	}
	if len(m.Args) != 1 || m.Args[0] != "mcp" {
		t.Errorf("MCP.Args = %v, want [mcp]", m.Args)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Part 2c — Catalog validation: External.MCP is validated on Load
// ─────────────────────────────────────────────────────────────────────────────

// TestCatalogValidation_ExternalMCP_EmptyCommandFails asserts that a catalog
// with an external harness whose mcp block has an empty command fails Load.
// A malformed catalog is a loud fail (build/release error — project convention).
func TestCatalogValidation_ExternalMCP_EmptyCommandFails(t *testing.T) {
	malformed := `harnesses:
  - id: bad-external
    name: Bad External
    type: external
    external:
      method: homebrew
      pkg: some/tap/tool
      mcp:
        name: tool
        command: ""
    install_modes: [lite]`

	_, err := parse([]byte(malformed))
	if err == nil {
		t.Fatal("expected error for external harness with mcp.command empty, got nil")
	}
	if !strings.Contains(err.Error(), "command") {
		t.Errorf("error %q should mention 'command', the failing field", err.Error())
	}
}

// TestCatalogValidation_ExternalMCP_WellFormedPasses asserts that an external
// harness with a valid mcp block loads cleanly.
func TestCatalogValidation_ExternalMCP_WellFormedPasses(t *testing.T) {
	valid := `harnesses:
  - id: good-external
    name: Good External
    type: external
    external:
      method: homebrew
      pkg: some/tap/tool
      mcp:
        name: tool
        command: tool
        args:
          - mcp
    install_modes: [lite]`

	if _, err := parse([]byte(valid)); err != nil {
		t.Errorf("unexpected error for valid external mcp block: %v", err)
	}
}

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
	for _, id := range []string{"openspec", "engram", "sdd-orchestrator", "active-orchestrator", "kb-creator"} {
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

	// C-32: no starter-only harness must appear in lite or full.
	for _, h := range lite {
		if h.IsStarterOnly() {
			t.Errorf("lite mode: harness %q is starter-only and must not appear in the global install plan", h.ID)
		}
	}
	for _, h := range full {
		if h.IsStarterOnly() {
			t.Errorf("full mode: harness %q is starter-only and must not appear in the global install plan", h.ID)
		}
	}
}

func TestForMode_ActiveOrchestratorIsFullOnly(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	for _, h := range c.ForMode(model.ModeLite) {
		if h.ID == "active-orchestrator" {
			t.Fatal("active-orchestrator must not be in lite: it orchestrates full-only skills (kb-creator, roadmap-generator, agent-instruction, find-skill)")
		}
	}

	var inFull bool
	for _, h := range c.ForMode(model.ModeFull) {
		if h.ID == "active-orchestrator" {
			inFull = true
		}
	}
	if !inFull {
		t.Fatal("active-orchestrator must be in full")
	}
}

// TestForMode_CustomReturnsOnlyGlobal asserts that custom mode returns only the
// 13 foundation-global harnesses — NOT the 30 starter-only C-30 skills.
// C-32: replaces TestForMode_CustomReturnsAll (which incorrectly expected all 43).
func TestForMode_CustomReturnsOnlyGlobal(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	const wantFoundationGlobal = 13
	got := c.ForMode(model.ModeCustom)
	if len(got) != wantFoundationGlobal {
		t.Errorf("custom mode returned %d harnesses, want %d foundation-global", len(got), wantFoundationGlobal)
	}
	for _, h := range got {
		if h.IsStarterOnly() {
			t.Errorf("custom mode: harness %q is starter-only and must not appear in custom mode", h.ID)
		}
	}
}

// TestForMode_ExcludesStarterOnlyHarnesses asserts that ForMode excludes
// starter-only harnesses from all modes (lite, full, custom).
// C-32: core invariant test — covers all three modes with count assertions.
func TestForMode_ExcludesStarterOnlyHarnesses(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	cases := []struct {
		mode      model.InstallMode
		wantCount int
	}{
		{model.ModeLite, 6},
		{model.ModeFull, 13},
		{model.ModeCustom, 13},
	}

	for _, tc := range cases {
		harnesses := c.ForMode(tc.mode)
		// No starter-only harness must appear.
		for _, h := range harnesses {
			if h.IsStarterOnly() {
				t.Errorf("ForMode(%q): harness %q has Scope=starter-only and must not appear in the global install plan", tc.mode, h.ID)
			}
		}
		// Count must match expected foundation-global count.
		if len(harnesses) != tc.wantCount {
			t.Errorf("ForMode(%q): got %d harnesses, want %d", tc.mode, len(harnesses), tc.wantCount)
		}
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
		case "clone", "embed":
			// valid
		default:
			t.Errorf("skill harness %q has unknown source.method %q", h.ID, h.Source.Method)
		}
	}
}

func TestMethodInference_ThirdParty(t *testing.T) {
	// A skill harness with third_party:true and no method should infer "clone"
	// (npx support was removed: third-party skills now clone like first-party).
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
	if h.Source.Method != "clone" {
		t.Errorf("expected method %q for third_party skill, got %q", "clone", h.Source.Method)
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
		"skill with removed npx method": {
			yaml: `harnesses:
  - id: x
    name: X
    type: skill
    source: { repo: some/repo, method: npx }
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

// ─────────────────────────────────────────────────────────────────────────────
// C-31: HarnessCommand catalog entry
// ─────────────────────────────────────────────────────────────────────────────

// TestC31_StarterAddCommand_InEmbeddedCatalog asserts that the embedded catalog
// contains the "starter-add-command" harness with type "command", validates
// successfully, and is scoped to Claude + OpenCode (focused agents only).
// HARD RULE: catalog.Load() must validate without error (invalid catalog = loud fail).
func TestC31_StarterAddCommand_InEmbeddedCatalog(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error on embedded catalog (catalog must be valid): %v", err)
	}

	h, ok := c.ByID("starter-add-command")
	if !ok {
		t.Fatal("harness 'starter-add-command' not found in embedded catalog")
	}

	if h.Type != model.HarnessCommand {
		t.Errorf("harness type = %q, want %q", h.Type, model.HarnessCommand)
	}

	// Must be scoped to the focused agents only (Claude + OpenCode).
	if !h.SupportsAgent(model.AgentClaude) {
		t.Error("starter-add-command must support AgentClaude")
	}
	if !h.SupportsAgent(model.AgentOpenCode) {
		t.Error("starter-add-command must support AgentOpenCode")
	}
	// Must NOT support other agents (focused-only).
	for _, other := range []model.Agent{model.AgentGemini, model.AgentCodex, model.AgentCursor} {
		if h.SupportsAgent(other) {
			t.Errorf("starter-add-command must NOT support %q (focused agents only)", other)
		}
	}
}

// TestC31_CommandHarnessType_IsValidInCatalogYAML asserts that a catalog entry
// with type: command passes validation (HarnessCommand.IsValid() returns true).
func TestC31_CommandHarnessType_IsValidInCatalogYAML(t *testing.T) {
	raw := `harnesses:
  - id: my-command
    name: My Command
    type: command
    install_modes: [lite, full]
    agents: [claude, opencode]`

	_, err := parse([]byte(raw))
	if err != nil {
		t.Errorf("catalog.parse() rejected a valid 'command' type harness: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// C-32: harness-scope-model — catalog.Load validation rules
// ─────────────────────────────────────────────────────────────────────────────

// TestCatalogLoad_StarterOnlyMustBeReferenced asserts that a starter-only harness
// that is NOT referenced by any starter fails catalog.Load with an error naming
// the harness id (Rule 1).
//
// RED: fails because Rule 1 does not exist yet in validate().
func TestCatalogLoad_StarterOnlyMustBeReferenced(t *testing.T) {
	// An orphaned starter-only harness: not referenced by any starter.
	orphanYAML := `harnesses:
  - id: orphan-skill
    name: Orphan Skill
    type: skill
    scope: starter-only
    source: { repo: some/repo, method: clone }
    install_modes: [full]
starters:
  - id: my-starter
    name: My Starter
    harnesses: []`

	_, err := parse([]byte(orphanYAML))
	if err == nil {
		t.Fatal("expected error for orphaned starter-only harness, got nil")
	}
	if !strings.Contains(err.Error(), "orphan-skill") {
		t.Errorf("error %q does not name the offending harness %q", err.Error(), "orphan-skill")
	}

	// Triangulation: a starter-only harness that IS referenced loads clean.
	referencedYAML := `harnesses:
  - id: curated-skill
    name: Curated Skill
    type: skill
    scope: starter-only
    source: { repo: some/repo, method: clone }
    install_modes: [full]
starters:
  - id: my-starter
    name: My Starter
    harnesses: [curated-skill]`

	if _, err := parse([]byte(referencedYAML)); err != nil {
		t.Errorf("unexpected error for referenced starter-only harness: %v", err)
	}
}

// TestCatalogLoad_StarterOnlyWithLiteModeIsInvalid asserts that a harness with
// scope: starter-only and install_modes including lite fails catalog.Load with
// an error naming the harness id (Rule 2).
//
// RED: fails because Rule 2 does not exist yet in validate().
func TestCatalogLoad_StarterOnlyWithLiteModeIsInvalid(t *testing.T) {
	// A starter-only harness that lists lite — contradiction.
	liteStarterOnlyYAML := `harnesses:
  - id: bad-skill
    name: Bad Skill
    type: skill
    scope: starter-only
    source: { repo: some/repo, method: clone }
    install_modes: [lite, full]
starters:
  - id: my-starter
    name: My Starter
    harnesses: [bad-skill]`

	_, err := parse([]byte(liteStarterOnlyYAML))
	if err == nil {
		t.Fatal("expected error for starter-only harness with lite mode, got nil")
	}
	if !strings.Contains(err.Error(), "bad-skill") {
		t.Errorf("error %q does not name the offending harness %q", err.Error(), "bad-skill")
	}

	// Triangulation: a starter-only harness with only [full] loads clean.
	fullOnlyYAML := `harnesses:
  - id: good-skill
    name: Good Skill
    type: skill
    scope: starter-only
    source: { repo: some/repo, method: clone }
    install_modes: [full]
starters:
  - id: my-starter
    name: My Starter
    harnesses: [good-skill]`

	if _, err := parse([]byte(fullOnlyYAML)); err != nil {
		t.Errorf("unexpected error for starter-only harness with [full] only: %v", err)
	}
}
