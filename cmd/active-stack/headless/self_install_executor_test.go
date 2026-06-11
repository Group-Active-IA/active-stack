package headless_test

import (
	"bytes"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestRunHeadless_PassesNoSelfInstallToOptions verifies that when
// params.NoSelfInstall is true, the BuildPlanFn receives NoSelfInstall=true.
func TestRunHeadless_PassesNoSelfInstallToOptions(t *testing.T) {
	var capturedNoSelfInstall bool

	cat := &fakeExecCatalog{harnesses: []model.Harness{}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	params := headless.ParsedFlags{
		TUI:           false,
		HomeDir:       t.TempDir(),
		NoSelfInstall: true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
		BuildPlanFn: func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			capturedNoSelfInstall = opts.NoSelfInstall
			return install.Plan{}, nil
		},
	}

	var out bytes.Buffer
	_ = headless.RunHeadless(params, cat, reg, &out)

	if !capturedNoSelfInstall {
		t.Error("expected NoSelfInstall=true forwarded to install.Options")
	}
}

// TestRunHeadless_DefaultSelfInstallEnabled verifies that without the flag,
// opts.NoSelfInstall is false (self-install ON).
func TestRunHeadless_DefaultSelfInstallEnabled(t *testing.T) {
	var capturedNoSelfInstall bool
	capturedNoSelfInstall = true // initialize to true; expect it to be set false

	cat := &fakeExecCatalog{harnesses: []model.Harness{}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	params := headless.ParsedFlags{
		TUI:     false,
		HomeDir: t.TempDir(),
		// NoSelfInstall not set → false (default)
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
		BuildPlanFn: func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			capturedNoSelfInstall = opts.NoSelfInstall
			return install.Plan{}, nil
		},
	}

	var out bytes.Buffer
	_ = headless.RunHeadless(params, cat, reg, &out)

	if capturedNoSelfInstall {
		t.Error("expected NoSelfInstall=false (self-install ON) when flag not set")
	}
}

// TestRunHeadless_ForwardsBinaryPath verifies that SelfInstallBinaryPath is
// forwarded to install.Options.
func TestRunHeadless_ForwardsBinaryPath(t *testing.T) {
	var capturedBinaryPath string

	cat := &fakeExecCatalog{harnesses: []model.Harness{}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	params := headless.ParsedFlags{
		TUI:           false,
		HomeDir:       t.TempDir(),
		BinaryPath:    "/tmp/my-binary",
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
		BuildPlanFn: func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			capturedBinaryPath = opts.SelfInstallBinaryPath
			return install.Plan{}, nil
		},
	}

	var out bytes.Buffer
	_ = headless.RunHeadless(params, cat, reg, &out)

	if capturedBinaryPath != "/tmp/my-binary" {
		t.Errorf("SelfInstallBinaryPath = %q, want %q", capturedBinaryPath, "/tmp/my-binary")
	}
}
