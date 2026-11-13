package catalog

import (
	"strings"
	"testing"
)

// Tarea 6.8 (opsx-task-context, D19): paridad por cliente y superficie. Lo que
// un cliente no puede recibir queda registrado como límite, no como entrega.

func TestParity_DeclaredSurfacesAreThoseOf04(t *testing.T) {
	var ids []string
	for _, s := range loadTC(t).Parity.Surfaces {
		ids = append(ids, s.ID)
	}
	if strings.Join(ids, ",") != "cli,ide,desktop-cloud" {
		t.Errorf("surfaces = %v, want cli, ide, desktop-cloud (codex-native-opsx/design.md:59-63)", ids)
	}
}

func TestParity_UndeliverableModulesAreLimitsNotDeliveries(t *testing.T) {
	tc := loadTC(t)
	cases := []struct {
		name, client, scope, module string
		delivered                   bool
	}{
		{"delegation without subagents (claude)", "claude", ScopeNoSubagents, "delegation", false},
		{"delegation without subagents (codex)", "codex", ScopeNoSubagents, "delegation", false},
		{"model-assignments in codex cli", "codex", "cli", "model-assignments", false},
		{"model-assignments in codex ide", "codex", "ide", "model-assignments", false},
		{"directed-code-reading in claude legacy", "claude", ScopeLegacyMode, "directed-code-reading", false},
		{"bounded-command-output in claude legacy", "claude", ScopeLegacyMode, "bounded-command-output", false},
		{"task-tool-selection in claude legacy", "claude", ScopeLegacyMode, "task-tool-selection", false},
		{"concise-agent-reporting in claude legacy", "claude", ScopeLegacyMode, "concise-agent-reporting", false},
		// triangulación: lo que sí se entrega
		{"delegation in claude cli", "claude", "cli", "delegation", true},
		{"model-assignments in claude cli", "claude", "cli", "model-assignments", true},
		{"directed-code-reading in claude cli", "claude", "cli", "directed-code-reading", true},
		{"directed-code-reading in codex cli", "codex", "cli", "directed-code-reading", true},
	}
	for _, c := range cases {
		d := tc.Delivery(c.client, c.scope, c.module)
		if d.Delivered != c.delivered {
			t.Errorf("%s: delivered = %v, want %v", c.name, d.Delivered, c.delivered)
		}
		if !c.delivered && d.Limit == "" {
			t.Errorf("%s: a limit must carry its reason", c.name)
		}
		if c.delivered && d.Limit != "" {
			t.Errorf("%s: a delivery carries no limit, got %q", c.name, d.Limit)
		}
	}
}

func TestParity_OmittedModulesHaveALimit(t *testing.T) {
	tc := loadTC(t)
	for _, m := range tc.Modules {
		for _, client := range m.OmittedIn {
			if tc.Delivery(client, "cli", m.ID).Delivered {
				t.Errorf("%s omitted in %s by the table but not declared as a limit", m.ID, client)
			}
		}
	}
}

func TestParity_CodexVerificationIsPending(t *testing.T) {
	var pending bool
	for _, p := range loadTC(t).Parity.Pending {
		if p.Client == "codex" && strings.Contains(p.Task, "2.5") {
			pending = true
		}
	}
	if !pending {
		t.Error("Codex parity must be declared pending on 04 task 2.5")
	}
}

func TestParity_RejectsInvalidDeclarations(t *testing.T) {
	for name, mutate := range map[string]func(p *Parity){
		"unknown module":    func(p *Parity) { p.Limits[0].Modules = []string{"nope"} },
		"unknown client":    func(p *Parity) { p.Limits[0].Client = "gemini" },
		"unknown scope":     func(p *Parity) { p.Limits[0].Scope = "mars" },
		"no reason":         func(p *Parity) { p.Limits[0].Reason = "" },
		"no limits":         func(p *Parity) { p.Limits = nil },
		"duplicate surface": func(p *Parity) { p.Surfaces = append(p.Surfaces, p.Surfaces[0]) },
	} {
		t.Run(name, func(t *testing.T) {
			tc := *loadTC(t)
			p := tc.Parity
			p.Limits = append([]ParityLimit(nil), p.Limits...)
			p.Surfaces = append([]ParitySurface(nil), p.Surfaces...)
			mutate(&p)
			tc.Parity = p
			if err := tc.validateParity(); err == nil {
				t.Error("want a validation error")
			}
		})
	}
}

func TestRenderLoadMarkdown_ListsParityLimits(t *testing.T) {
	md := loadTC(t).RenderLoadMarkdown()
	for _, want := range []string{"## Paridad por cliente y superficie", "Escritorio / cloud", "tarea 2.5 de la 04"} {
		if !strings.Contains(md, want) {
			t.Errorf("docs missing %q", want)
		}
	}
}
