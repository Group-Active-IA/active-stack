package uninstall_test

import (
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ─────────────────────────────────────────────────────────────────
// fakeCatalog implements uninstall.Catalog using an in-memory harness list.
// ─────────────────────────────────────────────────────────────────

type fakeCatalog struct {
	harnesses []model.Harness
}

func (f *fakeCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *fakeCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────────
// fakeAdapter satisfies uninstall.AgentAdapter.
// Paths are homeDir-relative for easy assertion.
// ─────────────────────────────────────────────────────────────────

type fakeAdapter struct {
	agent    model.Agent
	homeDir  string
	delivery model.ConfigDelivery
}

func (a fakeAdapter) Agent() model.Agent                     { return a.agent }
func (a fakeAdapter) InstructionsPath(homeDir string) string { return homeDir + "/instr.md" }
func (a fakeAdapter) SkillsDir(homeDir string) string        { return homeDir + "/skills" }
func (a fakeAdapter) SettingsPath(homeDir string) string     { return homeDir + "/settings.json" }
func (a fakeAdapter) ConfigDelivery() model.ConfigDelivery   { return a.delivery }

// fakeAdapterCustomPath allows overriding individual paths for path-resolution tests.
type fakeAdapterCustomPath struct {
	agent            model.Agent
	instructionsPath string
	skillsDir        string
	settingsPath     string
	delivery         model.ConfigDelivery
}

func (a fakeAdapterCustomPath) Agent() model.Agent                   { return a.agent }
func (a fakeAdapterCustomPath) ConfigDelivery() model.ConfigDelivery { return a.delivery }
func (a fakeAdapterCustomPath) InstructionsPath(homeDir string) string {
	if a.instructionsPath != "" {
		return a.instructionsPath
	}
	return homeDir + "/instr.md"
}
func (a fakeAdapterCustomPath) SkillsDir(homeDir string) string {
	if a.skillsDir != "" {
		return a.skillsDir
	}
	return homeDir + "/skills"
}
func (a fakeAdapterCustomPath) SettingsPath(homeDir string) string {
	if a.settingsPath != "" {
		return a.settingsPath
	}
	return homeDir + "/settings.json"
}

// ─────────────────────────────────────────────────────────────────
// fakeRegistry and fakeRegistryCustom
// ─────────────────────────────────────────────────────────────────

type fakeRegistry struct {
	adapters map[model.Agent]uninstall.AgentAdapter
}

func (r *fakeRegistry) Get(agent model.Agent) (uninstall.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// fakeRegistryCustom is the same but holds fakeAdapterCustomPath values.
type fakeRegistryCustom struct {
	adapters map[model.Agent]uninstall.AgentAdapter
}

func (r *fakeRegistryCustom) Get(agent model.Agent) (uninstall.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// ─────────────────────────────────────────────────────────────────
// buildUninstallOptions is a convenience constructor for tests.
// ─────────────────────────────────────────────────────────────────

func buildUninstallOptions(homeDir string, reg uninstall.Registry) uninstall.Options {
	return uninstall.Options{
		HomeDir:  homeDir,
		Registry: reg,
	}
}
