package tui

import (
	"io"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// TestModelDeps_NewFields verifies that the new fields added to ModelDeps
// are accessible after construction and do not break existing construction.
func TestModelDeps_NewFields_Starters(t *testing.T) {
	starters := []model.Starter{
		{ID: "my-starter", Name: "My Starter"},
	}
	deps := ModelDeps{
		Starters: starters,
	}
	m := newModel(deps)
	if len(m.deps.Starters) != 1 {
		t.Errorf("Starters len = %d, want 1", len(m.deps.Starters))
	}
	if m.deps.Starters[0].ID != "my-starter" {
		t.Errorf("Starters[0].ID = %q, want %q", m.deps.Starters[0].ID, "my-starter")
	}
}

func TestModelDeps_NewFields_BackupDir(t *testing.T) {
	deps := ModelDeps{BackupDir: "/some/backup/dir"}
	m := newModel(deps)
	if m.deps.BackupDir != "/some/backup/dir" {
		t.Errorf("BackupDir = %q, want %q", m.deps.BackupDir, "/some/backup/dir")
	}
}

func TestModelDeps_NewFields_RunUninstall(t *testing.T) {
	called := false
	deps := ModelDeps{
		RunUninstall: func(flags headless.ParsedUninstallFlags, w io.Writer) int {
			called = true
			return 0
		},
	}
	m := newModel(deps)
	if m.deps.RunUninstall == nil {
		t.Fatal("RunUninstall should not be nil after construction")
	}
	code := m.deps.RunUninstall(headless.ParsedUninstallFlags{}, io.Discard)
	if !called {
		t.Error("RunUninstall callback was not invoked")
	}
	if code != 0 {
		t.Errorf("RunUninstall returned %d, want 0", code)
	}
}

func TestModelDeps_NewFields_RunStarter(t *testing.T) {
	var gotID, gotPath string
	deps := ModelDeps{
		RunStarter: func(starterID, projectPath string, agents []model.Agent, w io.Writer) int {
			gotID = starterID
			gotPath = projectPath
			return 0
		},
	}
	m := newModel(deps)
	if m.deps.RunStarter == nil {
		t.Fatal("RunStarter should not be nil after construction")
	}
	m.deps.RunStarter("my-starter", "/proj", nil, io.Discard)
	if gotID != "my-starter" {
		t.Errorf("RunStarter gotID = %q, want %q", gotID, "my-starter")
	}
	if gotPath != "/proj" {
		t.Errorf("RunStarter gotPath = %q, want %q", gotPath, "/proj")
	}
}
