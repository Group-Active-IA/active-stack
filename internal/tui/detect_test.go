package tui

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestDetectInstalledAgentsIncludesCodex(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := DetectInstalledAgents(home)
	if !slices.Contains(got, model.AgentCodex) {
		t.Fatalf("DetectInstalledAgents() = %v, want codex", got)
	}
}
