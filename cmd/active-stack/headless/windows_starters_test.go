// Package headless — tests for "windows starters list" and
// "windows starters install" (windows-contract-hub-operations, Tasks 3.1/3.3
// RED).
package headless_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	extinstaller "github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// fakeStarterCatalog implements headless.StarterCatalog for the starters
// list/install tests, decoupled from the real embedded catalog so tests can
// control the exact starter/harness/MCP shape.
type fakeStarterCatalog struct {
	starters      []model.Starter
	harnessesByID map[string][]model.Harness // resolved harness set per starter id
	mcpsByID      map[string][]model.MCP     // resolved MCP set per starter id
	allHarnesses  []model.Harness
}

func (f *fakeStarterCatalog) StarterByID(id string) (model.Starter, bool) {
	for _, s := range f.starters {
		if s.ID == id {
			return s, true
		}
	}
	return model.Starter{}, false
}

func (f *fakeStarterCatalog) AllStarters() []model.Starter { return f.starters }

func (f *fakeStarterCatalog) ResolveStarter(id string) ([]model.Harness, error) {
	h, ok := f.harnessesByID[id]
	if !ok {
		return nil, fmt.Errorf("fake catalog: starter %q not found", id)
	}
	return h, nil
}

func (f *fakeStarterCatalog) ResolveStarterMCPs(id string) ([]model.MCP, error) {
	m, ok := f.mcpsByID[id]
	if !ok {
		if _, starterOK := f.harnessesByID[id]; starterOK {
			return nil, nil
		}
		return nil, fmt.Errorf("fake catalog: starter %q not found", id)
	}
	return m, nil
}

// install.Catalog surface, needed because fakeStarterCatalog also acts as the
// catalog passed into RunWindowsStartersInstall -> BuildStarterInstallParams
// -> RunHeadless.
func (f *fakeStarterCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.allHarnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *fakeStarterCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.allHarnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeStarterCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.allHarnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeStarterCatalog) AllHarnesses() []model.Harness { return f.allHarnesses }

func twoStarterFakeCatalog() *fakeStarterCatalog {
	harnessA := model.Harness{ID: "harness-a", Name: "Harness A", Type: model.HarnessExternal, External: &model.External{Method: "npm"}, InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}}
	harnessB := model.Harness{ID: "harness-b", Name: "Harness B", Type: model.HarnessExternal, External: &model.External{Method: "npm"}, InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}}
	return &fakeStarterCatalog{
		starters: []model.Starter{
			{
				ID:              "alpha",
				Name:            "Alpha",
				Description:     model.LocalizedText{"es": "Primer starter", "en": "First starter"},
				LongDescription: model.LocalizedText{"es": "Descripción larga del primer starter.", "en": "Long description of the first starter."},
				Includes:        []string{"beta"},
			},
			{
				ID:              "beta",
				Name:            "Beta",
				Description:     model.LocalizedText{"es": "Segundo starter", "en": "Second starter"},
				LongDescription: model.LocalizedText{"es": "Descripción larga del segundo starter.", "en": "Long description of the second starter."},
			},
		},
		harnessesByID: map[string][]model.Harness{
			"alpha": {harnessA, harnessB},
			"beta":  {harnessB},
		},
		mcpsByID: map[string][]model.MCP{
			"alpha": {{Name: "mcp-1", Command: "cmd1"}, {Name: "mcp-2", Command: "cmd2"}},
			"beta":  {{Name: "mcp-1", Command: "cmd1"}},
		},
		allHarnesses: []model.Harness{harnessA, harnessB},
	}
}

// TestRunWindowsStartersList_CatalogOrder asserts that "windows starters
// list" emits {"starters":[{id,name,description,includes,harnesses,mcp_count}]}
// in catalog order for a catalog with >=2 starters.
func TestRunWindowsStartersList_CatalogOrder(t *testing.T) {
	cat := twoStarterFakeCatalog()

	var out bytes.Buffer
	if err := headless.RunWindowsStartersList(cat, i18n.LangEN, &out); err != nil {
		t.Fatalf("RunWindowsStartersList() error = %v", err)
	}

	var resp struct {
		Starters []struct {
			ID              string   `json:"id"`
			Name            string   `json:"name"`
			Description     string   `json:"description"`
			Includes        []string `json:"includes"`
			Harnesses       []string `json:"harnesses"`
			MCPCount        int      `json:"mcp_count"`
			LongDescription string   `json:"long_description"`
		} `json:"starters"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal starters list json: %v\nbody=%s", err, out.String())
	}

	// L2 (catalog-localized-descriptions): long_description is now populated
	// from the starter's localized LongDescription (was empty pre-L2).
	for _, s := range resp.Starters {
		if s.LongDescription == "" {
			t.Errorf("starter %q long_description is empty, want populated from the catalog", s.ID)
		}
	}

	if len(resp.Starters) != 2 {
		t.Fatalf("starters len = %d, want 2", len(resp.Starters))
	}
	if resp.Starters[0].ID != "alpha" || resp.Starters[1].ID != "beta" {
		t.Fatalf("starters order = %+v, want [alpha, beta]", resp.Starters)
	}
	if resp.Starters[0].Name != "Alpha" || resp.Starters[0].Description != "First starter" {
		t.Errorf("starters[0] = %+v, want Name=Alpha Description='First starter'", resp.Starters[0])
	}
	if len(resp.Starters[0].Includes) != 1 || resp.Starters[0].Includes[0] != "beta" {
		t.Errorf("starters[0].Includes = %v, want [beta]", resp.Starters[0].Includes)
	}
	if len(resp.Starters[0].Harnesses) != 2 {
		t.Errorf("starters[0].Harnesses = %v, want 2 entries", resp.Starters[0].Harnesses)
	}
	if resp.Starters[0].MCPCount != 2 {
		t.Errorf("starters[0].MCPCount = %d, want 2", resp.Starters[0].MCPCount)
	}
	if resp.Starters[1].MCPCount != 1 {
		t.Errorf("starters[1].MCPCount = %d, want 1", resp.Starters[1].MCPCount)
	}
}

// TestRunWindowsStartersList_LocalizedDescriptionAndLongDescription covers
// L2 (catalog-localized-descriptions) D4: "windows starters list" under
// --lang es vs en emits the localized description and long_description for
// each starter, while name stays unchanged across languages.
func TestRunWindowsStartersList_LocalizedDescriptionAndLongDescription(t *testing.T) {
	cat := twoStarterFakeCatalog()

	type starterResp struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		LongDescription string `json:"long_description"`
	}

	for _, tc := range []struct {
		lang         i18n.Lang
		wantAlphaDes string
		wantAlphaLng string
	}{
		{lang: i18n.LangEN, wantAlphaDes: "First starter", wantAlphaLng: "Long description of the first starter."},
		{lang: i18n.LangES, wantAlphaDes: "Primer starter", wantAlphaLng: "Descripción larga del primer starter."},
	} {
		t.Run(string(tc.lang), func(t *testing.T) {
			var out bytes.Buffer
			if err := headless.RunWindowsStartersList(cat, tc.lang, &out); err != nil {
				t.Fatalf("RunWindowsStartersList() error = %v", err)
			}
			var resp struct {
				Starters []starterResp `json:"starters"`
			}
			if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal starters list json: %v\nbody=%s", err, out.String())
			}
			if len(resp.Starters) == 0 || resp.Starters[0].ID != "alpha" {
				t.Fatalf("starters = %+v, want first entry alpha", resp.Starters)
			}
			alpha := resp.Starters[0]
			if alpha.Name != "Alpha" {
				t.Errorf("alpha.Name = %q, want %q (name must not localize)", alpha.Name, "Alpha")
			}
			if alpha.Description != tc.wantAlphaDes {
				t.Errorf("alpha.Description = %q, want %q", alpha.Description, tc.wantAlphaDes)
			}
			if alpha.LongDescription != tc.wantAlphaLng {
				t.Errorf("alpha.LongDescription = %q, want %q", alpha.LongDescription, tc.wantAlphaLng)
			}
		})
	}
}

// TestRunWindowsStartersList_EmptyCatalog asserts an empty catalog yields
// {"starters":[]} (never null).
func TestRunWindowsStartersList_EmptyCatalog(t *testing.T) {
	cat := &fakeStarterCatalog{}

	var out bytes.Buffer
	if err := headless.RunWindowsStartersList(cat, i18n.LangEN, &out); err != nil {
		t.Fatalf("RunWindowsStartersList() error = %v", err)
	}

	got := out.String()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(got), &raw); err != nil {
		t.Fatalf("unmarshal: %v\nbody=%s", err, got)
	}
	if string(raw["starters"]) != "[]" {
		t.Fatalf("starters = %s, want []", raw["starters"])
	}
}

// ── windows starters install ────────────────────────────────────────────────

func stubExternalInstall() func() {
	return install.SetExternalInstallFnWithDownload(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
		downloadFn extinstaller.DownloadEventFunc,
	) (extinstaller.Result, error) {
		return extinstaller.Result{}, nil
	})
}

type startersInstallFakeAdapter struct{ agent model.Agent }

func (a startersInstallFakeAdapter) Agent() model.Agent               { return a.agent }
func (a startersInstallFakeAdapter) InstructionsPath(d string) string { return d + "/CLAUDE.md" }
func (a startersInstallFakeAdapter) SkillsDir(d string) string        { return d + "/skills" }
func (a startersInstallFakeAdapter) CommandsDir(d string) string      { return d + "/commands" }
func (a startersInstallFakeAdapter) SettingsPath(d string) string     { return d + "/settings.json" }
func (a startersInstallFakeAdapter) MCPConfigPath(d, s string) string {
	return d + "/mcp/" + s + ".json"
}
func (a startersInstallFakeAdapter) MCPStrategy() extinstaller.MCPStrategy {
	return extinstaller.StrategySeparateFile
}
func (a startersInstallFakeAdapter) VariantKey() string { return string(a.agent) }
func (a startersInstallFakeAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a startersInstallFakeAdapter) PathsFor(base string, _ model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{
		InstructionsPath: base + "/CLAUDE.md",
		SkillsDir:        base + "/skills",
		SettingsPath:     base + "/settings.json",
		CommandsDir:      base + "/commands",
	}.WithMCPConfigFn(func(s string) string { return base + "/mcp/" + s + ".json" }).
		WithMCPStrategy(model.MCPStrategySingleFileMerge)
}

type startersInstallFakeReg struct {
	adapters map[model.Agent]install.AgentAdapter
}

func (r *startersInstallFakeReg) Get(agent model.Agent) (install.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// TestRunWindowsStartersInstall_Success asserts that a successful starters
// install streams the install event contract ending in starter_finished.
func TestRunWindowsStartersInstall_Success(t *testing.T) {
	cat := twoStarterFakeCatalog()
	reg := &startersInstallFakeReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: startersInstallFakeAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	defer stubExternalInstall()()

	projectRoot := t.TempDir()
	flags := headless.ParsedStarterAddFlags{
		StarterID:   "alpha",
		ProjectPath: projectRoot,
		Yes:         true,
		Agents:      []model.Agent{model.AgentClaude},
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsStartersInstall(flags, cat, reg, nil, &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsStartersInstall() exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := splitTestJSONLines(out.String())
	if len(lines) < 2 {
		t.Fatalf("expected multiple json stream lines, got %d: %q", len(lines), out.String())
	}
	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "starter_finished" || !last.Success {
		t.Fatalf("last event = %+v, want starter_finished success=true", last)
	}
}

// TestRunWindowsStartersInstall_DryRunDoesNotMutate asserts that --dry-run
// produces the stream without creating/modifying any files in the project.
func TestRunWindowsStartersInstall_DryRunDoesNotMutate(t *testing.T) {
	cat := twoStarterFakeCatalog()
	reg := &startersInstallFakeReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: startersInstallFakeAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	defer stubExternalInstall()()

	projectRoot := t.TempDir()
	before := listDirEntries(t, projectRoot)

	flags := headless.ParsedStarterAddFlags{
		StarterID:   "alpha",
		ProjectPath: projectRoot,
		DryRun:      true,
		Yes:         true,
		Agents:      []model.Agent{model.AgentClaude},
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsStartersInstall(flags, cat, reg, nil, &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsStartersInstall() dry-run exit = %d; output:\n%s", exitCode, out.String())
	}

	after := listDirEntries(t, projectRoot)
	if len(before) != len(after) {
		t.Fatalf("dry-run mutated project dir: before=%v after=%v", before, after)
	}
}

// TestRunWindowsStartersInstall_UnknownStarterFails asserts an unknown
// starter id fails the stream with a non-zero exit code.
func TestRunWindowsStartersInstall_UnknownStarterFails(t *testing.T) {
	cat := twoStarterFakeCatalog()
	reg := &startersInstallFakeReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: startersInstallFakeAdapter{agent: model.AgentClaude},
	}}

	flags := headless.ParsedStarterAddFlags{
		StarterID:   "does-not-exist",
		ProjectPath: t.TempDir(),
		Yes:         true,
		Agents:      []model.Agent{model.AgentClaude},
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsStartersInstall(flags, cat, reg, nil, &out)
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit for unknown starter; output:\n%s", out.String())
	}
}

func listDirEntries(t *testing.T, dir string) []string {
	t.Helper()
	var names []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		names = append(names, rel)
		return nil
	})
	return names
}

func splitTestJSONLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 {
				out = append(out, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		line := s[start:]
		if len(line) > 0 {
			out = append(out, line)
		}
	}
	return out
}
