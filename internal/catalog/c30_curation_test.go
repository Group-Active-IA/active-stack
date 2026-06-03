// Package catalog — C-30 starter curation tests.
//
// All tests in this file follow the strict TDD cycle mandated by C-30:
//   - Written RED (failing) before any curated YAML is added.
//   - GREEN once the curated harnesses + starters section is in harnesses.yaml.
//   - After §4 GREEN, the file goes through §5 REFACTOR (YAML tidy only).
//
// Spec source: openspec/changes/c30-starter-curation/specs/starter-catalog-curation/spec.md
package catalog

import (
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// 3.2 RED — base starter resolves to exactly the 6 transversal harnesses
// ─────────────────────────────────────────────────────────────────────────────

// TestC30_BaseStarter_ResolvesTo6Transversals asserts that:
//  1. The embedded catalog exposes a starter with id "base".
//  2. ResolveStarter("base") returns exactly the 6 transversal harness ids.
//  3. No harness id appears more than once (dedup).
//
// RED: fails because "base" starter does not exist in the embedded catalog yet.
func TestC30_BaseStarter_ResolvesTo6Transversals(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 1. Starter must exist.
	_, ok := c.StarterByID("base")
	if !ok {
		t.Fatal("starter 'base' not found in embedded catalog — add it to harnesses.yaml")
	}

	// 2. Resolve and check the six transversal ids.
	wantTransversals := []string{
		"test-driven-development",
		"systematic-debugging",
		"requesting-code-review",
		"receiving-code-review",
		"code-review-excellence",
		"agile-product-owner",
	}

	harnesses, err := c.ResolveStarter("base")
	if err != nil {
		t.Fatalf("ResolveStarter('base') error = %v", err)
	}

	ids := make(map[string]int, len(harnesses))
	for _, h := range harnesses {
		ids[h.ID]++
	}

	// Must contain exactly the 6 transversals.
	for _, want := range wantTransversals {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('base'): missing transversal harness %q", want)
		}
	}
	// Must not contain other harnesses.
	wantSet := make(map[string]bool, len(wantTransversals))
	for _, id := range wantTransversals {
		wantSet[id] = true
	}
	for id := range ids {
		if !wantSet[id] {
			t.Errorf("ResolveStarter('base'): unexpected harness %q in base starter", id)
		}
	}

	// 3. No duplicates.
	for id, count := range ids {
		if count > 1 {
			t.Errorf("ResolveStarter('base'): harness %q appears %d times, want 1", id, count)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3.3 RED — backend resolves to base ∪ 12 backend-layer harnesses, deduped
// ─────────────────────────────────────────────────────────────────────────────

// TestC30_BackendStarter_ResolvesToBaseUnionBackendLayer asserts that
// ResolveStarter("backend") returns the union of the 6 base transversals and
// the 12 backend-layer harnesses, each appearing exactly once.
//
// RED: fails because "backend" is a placeholder with no harnesses, and "base"
// does not exist yet.
func TestC30_BackendStarter_ResolvesToBaseUnionBackendLayer(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	baseTransversals := []string{
		"test-driven-development",
		"systematic-debugging",
		"requesting-code-review",
		"receiving-code-review",
		"code-review-excellence",
		"agile-product-owner",
	}
	backendLayer := []string{
		"clean-architecture",
		"fastapi-domain-service",
		"fastapi-code-review",
		"redis-best-practices",
		"websocket-engineer",
		"alembic-migrations",
		"fastapi-templates",
		"python-testing-patterns",
		"postgresql-table-design",
		"postgresql-optimization",
		"api-security-best-practices",
		"multi-stage-dockerfile",
	}

	harnesses, err := c.ResolveStarter("backend")
	if err != nil {
		t.Fatalf("ResolveStarter('backend') error = %v", err)
	}

	ids := make(map[string]int, len(harnesses))
	for _, h := range harnesses {
		ids[h.ID]++
	}

	// All base transversals must be present (inherited via includes: [base]).
	for _, want := range baseTransversals {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('backend'): missing base transversal %q", want)
		}
	}
	// All backend-layer harnesses must be present.
	for _, want := range backendLayer {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('backend'): missing backend-layer harness %q", want)
		}
	}
	// Dedup: each id exactly once.
	for id, count := range ids {
		if count > 1 {
			t.Errorf("ResolveStarter('backend'): harness %q appears %d times, want 1", id, count)
		}
	}

	// Total count: 6 base + 12 backend = 18.
	wantTotal := len(baseTransversals) + len(backendLayer)
	if len(harnesses) != wantTotal {
		t.Errorf("ResolveStarter('backend'): %d harnesses, want %d (6 base + 12 backend)", len(harnesses), wantTotal)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3.4 RED — ux-ui resolves to base ∪ 11 frontend-layer harnesses, deduped
// ─────────────────────────────────────────────────────────────────────────────

// TestC30_UxUiStarter_ResolvesToBaseUnionFrontendLayer asserts that
// ResolveStarter("ux-ui") returns the union of the 6 base transversals and
// the 11 frontend-layer harnesses, each appearing exactly once.
//
// RED: fails because "ux-ui" is a placeholder with no harnesses.
func TestC30_UxUiStarter_ResolvesToBaseUnionFrontendLayer(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	baseTransversals := []string{
		"test-driven-development",
		"systematic-debugging",
		"requesting-code-review",
		"receiving-code-review",
		"code-review-excellence",
		"agile-product-owner",
	}
	frontendLayer := []string{
		"vercel-react-best-practices",
		"zustand-store-pattern",
		"react19-form-pattern",
		"dashboard-crud-page",
		"ws-frontend-subscription",
		"help-system-content",
		"interface-design",
		"pwa-development",
		"typescript-advanced-types",
		"tailwind-design-system",
		"playwright-best-practices",
	}

	harnesses, err := c.ResolveStarter("ux-ui")
	if err != nil {
		t.Fatalf("ResolveStarter('ux-ui') error = %v", err)
	}

	ids := make(map[string]int, len(harnesses))
	for _, h := range harnesses {
		ids[h.ID]++
	}

	// All base transversals must be present (inherited via includes: [base]).
	for _, want := range baseTransversals {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('ux-ui'): missing base transversal %q", want)
		}
	}
	// All frontend-layer harnesses must be present.
	for _, want := range frontendLayer {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('ux-ui'): missing frontend-layer harness %q", want)
		}
	}
	// Dedup: each id exactly once.
	for id, count := range ids {
		if count > 1 {
			t.Errorf("ResolveStarter('ux-ui'): harness %q appears %d times, want 1", id, count)
		}
	}

	// Total count: 6 base + 11 frontend = 17.
	wantTotal := len(baseTransversals) + len(frontendLayer)
	if len(harnesses) != wantTotal {
		t.Errorf("ResolveStarter('ux-ui'): %d harnesses, want %d (6 base + 11 frontend)", len(harnesses), wantTotal)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3.5 RED — active-ia deduplicates base (key edge case)
// ─────────────────────────────────────────────────────────────────────────────

// TestC30_ActiveIa_DeduplicatesBase asserts the key compositional edge case:
// active-ia includes both backend and ux-ui, which both include base.
// The 6 base transversals must appear exactly ONCE in the resolved set.
// active-ia also directly bundles monorepo-scaffold.
//
// RED: fails because the curated harnesses and the "base" starter do not exist.
func TestC30_ActiveIa_DeduplicatesBase(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	baseTransversals := []string{
		"test-driven-development",
		"systematic-debugging",
		"requesting-code-review",
		"receiving-code-review",
		"code-review-excellence",
		"agile-product-owner",
	}
	backendLayer := []string{
		"clean-architecture",
		"fastapi-domain-service",
		"fastapi-code-review",
		"redis-best-practices",
		"websocket-engineer",
		"alembic-migrations",
		"fastapi-templates",
		"python-testing-patterns",
		"postgresql-table-design",
		"postgresql-optimization",
		"api-security-best-practices",
		"multi-stage-dockerfile",
	}
	frontendLayer := []string{
		"vercel-react-best-practices",
		"zustand-store-pattern",
		"react19-form-pattern",
		"dashboard-crud-page",
		"ws-frontend-subscription",
		"help-system-content",
		"interface-design",
		"pwa-development",
		"typescript-advanced-types",
		"tailwind-design-system",
		"playwright-best-practices",
	}

	harnesses, err := c.ResolveStarter("active-ia")
	if err != nil {
		t.Fatalf("ResolveStarter('active-ia') error = %v", err)
	}

	ids := make(map[string]int, len(harnesses))
	for _, h := range harnesses {
		ids[h.ID]++
	}

	// KEY EDGE CASE: each base transversal must appear exactly ONCE even though
	// both backend and ux-ui (both included by active-ia) include base.
	for _, base := range baseTransversals {
		if ids[base] == 0 {
			t.Errorf("ResolveStarter('active-ia'): missing base transversal %q", base)
		}
		if ids[base] > 1 {
			t.Errorf("ResolveStarter('active-ia'): base transversal %q appears %d times, want 1 (dedup)", base, ids[base])
		}
	}

	// Backend-layer must be present, each once.
	for _, want := range backendLayer {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('active-ia'): missing backend-layer harness %q", want)
		}
		if ids[want] > 1 {
			t.Errorf("ResolveStarter('active-ia'): backend harness %q appears %d times, want 1", want, ids[want])
		}
	}

	// Frontend-layer must be present, each once.
	for _, want := range frontendLayer {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter('active-ia'): missing frontend-layer harness %q", want)
		}
		if ids[want] > 1 {
			t.Errorf("ResolveStarter('active-ia'): frontend harness %q appears %d times, want 1", want, ids[want])
		}
	}

	// monorepo-scaffold must be present.
	if ids["monorepo-scaffold"] == 0 {
		t.Error("ResolveStarter('active-ia'): missing monorepo-scaffold")
	}
	if ids["monorepo-scaffold"] > 1 {
		t.Errorf("ResolveStarter('active-ia'): monorepo-scaffold appears %d times, want 1", ids["monorepo-scaffold"])
	}

	// Global dedup: each id at most once.
	for id, count := range ids {
		if count > 1 {
			t.Errorf("ResolveStarter('active-ia'): harness %q appears %d times, want 1", id, count)
		}
	}

	// Total: 6 base + 12 backend + 11 frontend + 1 monorepo-scaffold = 30.
	wantTotal := len(baseTransversals) + len(backendLayer) + len(frontendLayer) + 1
	if len(harnesses) != wantTotal {
		t.Errorf("ResolveStarter('active-ia'): %d harnesses, want %d", len(harnesses), wantTotal)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3.6 RED — curated catalog loads clean; no duplicate harness ids
// ─────────────────────────────────────────────────────────────────────────────

// TestC30_CuratedCatalog_LoadsClean asserts that after adding the curated
// entries, catalog.Load() succeeds and the 4 starters (base, backend, ux-ui,
// active-ia) are all present.
//
// RED: fails because "base" starter and the 30 curated harnesses are absent.
func TestC30_CuratedCatalog_LoadsClean(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error on embedded catalog: %v", err)
	}

	for _, id := range []string{"base", "backend", "ux-ui", "active-ia"} {
		if _, ok := c.StarterByID(id); !ok {
			t.Errorf("starter %q not found in embedded catalog after curation", id)
		}
	}
}

// TestC30_CuratedCatalog_NoDuplicateHarnessIDs asserts that no harness id
// appears more than once in the embedded catalog — specifically that the
// substrate harnesses (openspec-*, find-skills, sdd-orchestrator, permissions)
// are not accidentally duplicated when the curated entries are added.
//
// RED: compile-passes but fails once curated entries introduce a duplicate (or
// passes trivially if no curation has been done yet — the real value is it
// guards against regressions after §4).
func TestC30_CuratedCatalog_NoDuplicateHarnessIDs(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v — duplicate ids would be caught here", err)
	}

	seen := make(map[string]int, len(c.Harnesses))
	for _, h := range c.Harnesses {
		seen[h.ID]++
	}
	for id, count := range seen {
		if count > 1 {
			t.Errorf("harness id %q appears %d times in the catalog, want 1", id, count)
		}
	}
}

// TestC30_CuratedSkillHarnesses_OwnPointAtMonorepo asserts that all 14
// own/fork curated skills (identified by absence of third_party=true) point at
// the published monorepo JuanCruzRobledo/jr-skills with method=clone and a
// non-empty path.
//
// RED: fails because the 14 curated own skill harnesses do not exist yet.
func TestC30_CuratedSkillHarnesses_OwnPointAtMonorepo(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	ownCuratedIDs := []string{
		// propias puras
		"fastapi-domain-service",
		"alembic-migrations",
		"zustand-store-pattern",
		"react19-form-pattern",
		"dashboard-crud-page",
		"ws-frontend-subscription",
		"help-system-content",
		"monorepo-scaffold",
		// forks expandidos
		"postgresql-table-design",
		"tailwind-design-system",
		"typescript-advanced-types",
		"interface-design",
		"clean-architecture",
		"fastapi-code-review",
	}

	const wantRepo = "JuanCruzRobledo/jr-skills"

	for _, id := range ownCuratedIDs {
		h, ok := c.ByID(id)
		if !ok {
			t.Errorf("own/fork skill harness %q not found in catalog", id)
			continue
		}
		if h.ThirdParty {
			t.Errorf("own/fork skill %q must NOT have third_party=true", id)
		}
		if h.Source == nil {
			t.Errorf("own/fork skill %q has nil source", id)
			continue
		}
		if h.Source.Repo != wantRepo {
			t.Errorf("own/fork skill %q: source.repo = %q, want %q", id, h.Source.Repo, wantRepo)
		}
		if h.Source.Method != "clone" {
			t.Errorf("own/fork skill %q: source.method = %q, want clone", id, h.Source.Method)
		}
		if h.Source.Path == "" {
			t.Errorf("own/fork skill %q: source.path is empty, want skills/%s", id, id)
		}
	}
}

// TestC30_CuratedSkillHarnesses_ThirdPartyAreBestEffort asserts that all 16
// third-party curated skills have third_party=true AND best_effort=true.
//
// RED: fails because the 16 third-party curated skill harnesses do not exist yet.
func TestC30_CuratedSkillHarnesses_ThirdPartyAreBestEffort(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	thirdPartyCuratedIDs := []string{
		"fastapi-templates",
		"python-testing-patterns",
		"postgresql-optimization",
		"api-security-best-practices",
		"multi-stage-dockerfile",
		"systematic-debugging",
		"test-driven-development",
		"playwright-best-practices",
		"receiving-code-review",
		"code-review-excellence",
		"vercel-react-best-practices",
		"requesting-code-review",
		"agile-product-owner",
		"redis-best-practices",
		"websocket-engineer",
		"pwa-development",
	}

	for _, id := range thirdPartyCuratedIDs {
		h, ok := c.ByID(id)
		if !ok {
			t.Errorf("third-party skill harness %q not found in catalog", id)
			continue
		}
		if !h.ThirdParty {
			t.Errorf("third-party skill %q: ThirdParty = false, want true", id)
		}
		if !h.BestEffort {
			t.Errorf("third-party skill %q: BestEffort = false, want true", id)
		}
	}
}
