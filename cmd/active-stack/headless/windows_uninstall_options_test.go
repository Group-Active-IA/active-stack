// Package headless — tests for "windows uninstall-options"
// (windows-contract-hub-operations, Task 6.1 RED).
package headless_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
)

// TestRunWindowsUninstallOptions_CarriesDetectedAgentsModesAndStrategies
// asserts that "windows uninstall-options" reports detected_agents equal to
// what "windows detect" reports for the same home, a non-empty modes list,
// and strategies including both targeted and restore.
func TestRunWindowsUninstallOptions_CarriesDetectedAgentsModesAndStrategies(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(home+"/.claude", 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	var detectOut bytes.Buffer
	if err := headless.RunWindowsDetect(home, nil, &detectOut); err != nil {
		t.Fatalf("RunWindowsDetect() error = %v", err)
	}
	var detectResp struct {
		DetectedAgents []string `json:"detected_agents"`
	}
	if err := json.Unmarshal(detectOut.Bytes(), &detectResp); err != nil {
		t.Fatalf("unmarshal detect json: %v", err)
	}

	var out bytes.Buffer
	if err := headless.RunWindowsUninstallOptions(home, i18n.LangEN, &out); err != nil {
		t.Fatalf("RunWindowsUninstallOptions() error = %v", err)
	}

	var resp struct {
		DetectedAgents []string `json:"detected_agents"`
		Modes          []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"modes"`
		Strategies []struct {
			ID               string `json:"id"`
			Label            string `json:"label"`
			Description      string `json:"description"`
			Default          bool   `json:"default"`
			RequiresManifest bool   `json:"requires_manifest"`
		} `json:"strategies"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal uninstall-options json: %v\nbody=%s", err, out.String())
	}

	if len(resp.DetectedAgents) != len(detectResp.DetectedAgents) {
		t.Fatalf("detected_agents = %v, want %v (from windows detect)", resp.DetectedAgents, detectResp.DetectedAgents)
	}
	for i := range resp.DetectedAgents {
		if resp.DetectedAgents[i] != detectResp.DetectedAgents[i] {
			t.Fatalf("detected_agents = %v, want %v (from windows detect)", resp.DetectedAgents, detectResp.DetectedAgents)
		}
	}

	if len(resp.Modes) == 0 {
		t.Fatal("modes must be non-empty")
	}

	var targeted, restore *struct {
		ID               string `json:"id"`
		Label            string `json:"label"`
		Description      string `json:"description"`
		Default          bool   `json:"default"`
		RequiresManifest bool   `json:"requires_manifest"`
	}
	for i := range resp.Strategies {
		s := resp.Strategies[i]
		switch s.ID {
		case "targeted":
			targeted = &s
		case "restore":
			restore = &s
		}
	}
	if targeted == nil {
		t.Fatal("strategies must include targeted")
	}
	if !targeted.Default || targeted.RequiresManifest {
		t.Errorf("targeted strategy = %+v, want default=true requires_manifest=false", targeted)
	}
	if restore == nil {
		t.Fatal("strategies must include restore")
	}
	if !restore.RequiresManifest {
		t.Errorf("restore strategy = %+v, want requires_manifest=true", restore)
	}
}

// TestRunWindowsUninstallOptions_LongDescriptions covers task 5.1: default EN
// keeps targeted default:true/requires_manifest:false and restore
// requires_manifest:true; each strategy and mode carries a populated
// long_description.
func TestRunWindowsUninstallOptions_LongDescriptions(t *testing.T) {
	home := t.TempDir()

	var out bytes.Buffer
	if err := headless.RunWindowsUninstallOptions(home, i18n.LangEN, &out); err != nil {
		t.Fatalf("RunWindowsUninstallOptions() error = %v", err)
	}

	var resp struct {
		Modes []struct {
			ID              string `json:"id"`
			LongDescription string `json:"long_description"`
		} `json:"modes"`
		Strategies []struct {
			ID               string `json:"id"`
			Default          bool   `json:"default"`
			RequiresManifest bool   `json:"requires_manifest"`
			LongDescription  string `json:"long_description"`
		} `json:"strategies"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal uninstall-options json: %v\nbody=%s", err, out.String())
	}

	for _, m := range resp.Modes {
		if m.LongDescription == "" {
			t.Errorf("mode %q long_description is empty, want populated", m.ID)
		}
	}

	var targeted, restore *struct {
		ID               string `json:"id"`
		Default          bool   `json:"default"`
		RequiresManifest bool   `json:"requires_manifest"`
		LongDescription  string `json:"long_description"`
	}
	for i := range resp.Strategies {
		s := resp.Strategies[i]
		if s.LongDescription == "" {
			t.Errorf("strategy %q long_description is empty, want populated", s.ID)
		}
		switch s.ID {
		case "targeted":
			targeted = &s
		case "restore":
			restore = &s
		}
	}
	if targeted == nil || !targeted.Default || targeted.RequiresManifest {
		t.Errorf("targeted strategy = %+v, want default=true requires_manifest=false", targeted)
	}
	if restore == nil || !restore.RequiresManifest {
		t.Errorf("restore strategy = %+v, want requires_manifest=true", restore)
	}
}

// TestRunWindowsUninstallOptions_SpanishLocalizesLabels covers task 5.1's ES
// case: --lang es localizes strategy/mode labels+descriptions while id and
// requires_manifest values are unchanged.
func TestRunWindowsUninstallOptions_SpanishLocalizesLabels(t *testing.T) {
	home := t.TempDir()

	var out bytes.Buffer
	if err := headless.RunWindowsUninstallOptions(home, i18n.LangES, &out); err != nil {
		t.Fatalf("RunWindowsUninstallOptions() error = %v", err)
	}

	var resp struct {
		Modes []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"modes"`
		Strategies []struct {
			ID               string `json:"id"`
			Label            string `json:"label"`
			RequiresManifest bool   `json:"requires_manifest"`
		} `json:"strategies"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal uninstall-options json: %v\nbody=%s", err, out.String())
	}

	wantModeLabels := map[string]string{"lite": "Rápido", "full": "Completo", "custom": "Personalizado"}
	for _, m := range resp.Modes {
		if want, ok := wantModeLabels[m.ID]; ok && m.Label != want {
			t.Errorf("mode %q label = %q, want %q", m.ID, m.Label, want)
		}
	}

	for _, s := range resp.Strategies {
		switch s.ID {
		case "targeted":
			if s.Label != "Dirigida" {
				t.Errorf("targeted label = %q, want Dirigida", s.Label)
			}
			if s.RequiresManifest {
				t.Error("targeted requires_manifest must stay false under es")
			}
		case "restore":
			if s.Label != "Restaurar desde backup" {
				t.Errorf("restore label = %q, want %q", s.Label, "Restaurar desde backup")
			}
			if !s.RequiresManifest {
				t.Error("restore requires_manifest must stay true under es")
			}
		}
	}
}
