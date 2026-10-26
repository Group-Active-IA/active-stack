package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestDetectInstalledAgents_IncludesAntigravity(t *testing.T) {
	home := t.TempDir()
	agyDir := filepath.Join(home, ".gemini", "antigravity")
	if err := os.MkdirAll(agyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", agyDir, err)
	}

	got := DetectInstalledAgents(home)
	if len(got) != 1 || got[0] != model.AgentAntigravity {
		t.Fatalf("DetectInstalledAgents() = %v, want [%s]", got, model.AgentAntigravity)
	}
}
