package catalog

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// [EN] Fase 1 (R-07): entrada `engram-cloud` del catálogo.

func engramCloud(t *testing.T) (*Catalog, model.Harness) {
	t.Helper()
	cat, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	h, ok := cat.ByID("engram-cloud")
	if !ok {
		t.Fatal("missing harness engram-cloud")
	}
	return cat, h
}

// T-F1-1: existe, no está en Lite y sí en Full/Custom.
func TestEngramCloudModes(t *testing.T) {
	_, h := engramCloud(t)
	if h.Type != model.HarnessConfig {
		t.Errorf("type = %q, want config", h.Type)
	}
	if h.InMode(model.ModeLite) {
		t.Error("engram-cloud must not be in Lite")
	}
	if !h.InMode(model.ModeFull) || !h.InMode(model.ModeCustom) {
		t.Error("engram-cloud must be in Full and Custom")
	}
}

// T-F1-2: agents es exactamente [claude, codex].
func TestEngramCloudAgentsExactly(t *testing.T) {
	_, h := engramCloud(t)
	want := []model.Agent{model.AgentClaude, model.AgentCodex}
	if !reflect.DeepEqual(h.Agents, want) {
		t.Errorf("agents = %v, want exactly %v", h.Agents, want)
	}
	for _, a := range []model.Agent{model.AgentOpenCode, model.AgentAntigravity} {
		if h.SupportsAgent(a) {
			t.Errorf("engram-cloud must not support %s", a)
		}
	}
}

// T-F1-3: cada dependencia existe en el catálogo.
func TestEngramCloudDependenciesExist(t *testing.T) {
	cat, h := engramCloud(t)
	want := []string{"engram", "sdd-orchestrator"}
	if !reflect.DeepEqual(h.DependsOn, want) {
		t.Errorf("depends_on = %v, want %v", h.DependsOn, want)
	}
	for _, dep := range h.DependsOn {
		if _, ok := cat.ByID(dep); !ok {
			t.Errorf("dependency %q not in catalog", dep)
		}
	}
}

// T-F1-4: textos es/en no vacíos.
func TestEngramCloudLocalizedTexts(t *testing.T) {
	_, h := engramCloud(t)
	for _, lang := range []string{"es", "en"} {
		if strings.TrimSpace(h.Description[lang]) == "" {
			t.Errorf("description[%s] is empty", lang)
		}
		if strings.TrimSpace(h.LongDescription[lang]) == "" {
			t.Errorf("long_description[%s] is empty", lang)
		}
	}
}

// T-F1-5: sin tokens, URLs de servidor ni gentle-*.
func TestEngramCloudNoSecretsOrLeftovers(t *testing.T) {
	_, h := engramCloud(t)
	urlRe := regexp.MustCompile(`https?://[^\s)]+`)
	allowedHost := "https://github.com/Gentleman-Programming/engram/"
	tokenRe := regexp.MustCompile(`(?i)(token|bearer|secret|password|api[_-]?key)\s*[:=]`)
	for _, field := range []model.LocalizedText{h.Description, h.LongDescription} {
		for lang, text := range field {
			if strings.Contains(strings.ToLower(text), "gentle-") {
				t.Errorf("[%s] contains gentle-*: %q", lang, text)
			}
			if tokenRe.MatchString(text) {
				t.Errorf("[%s] looks like it embeds a credential", lang)
			}
			for _, u := range urlRe.FindAllString(text, -1) {
				if !strings.HasPrefix(u, allowedHost) {
					t.Errorf("[%s] URL %q is not Engram documentation", lang, u)
				}
			}
		}
	}
}

// T-F1-6: no es best_effort.
func TestEngramCloudNotBestEffort(t *testing.T) {
	_, h := engramCloud(t)
	if h.BestEffort {
		t.Error("engram-cloud must not be best_effort")
	}
}

// T-F1-7: long_description declara la limitación y enlaza la documentación.
func TestEngramCloudLongDescriptionStatesLimits(t *testing.T) {
	_, h := engramCloud(t)
	for lang, text := range h.LongDescription {
		if !strings.Contains(text, "OpenCode") || !strings.Contains(text, "Antigravity") {
			t.Errorf("[%s] must state that OpenCode and Antigravity are out", lang)
		}
		if !strings.Contains(text, "https://github.com/Gentleman-Programming/engram/blob/main/docs/engram-cloud/README.md") {
			t.Errorf("[%s] must link the Engram Cloud documentation", lang)
		}
		if strings.Contains(strings.ToLower(text), "hook") {
			t.Errorf("[%s] must not present a hook as the mechanism (DA-05)", lang)
		}
	}
}
