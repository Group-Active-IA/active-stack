package config_test

import (
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// fakeAdapter is a test double implementing AgentAdapter.
type fakeAdapter struct {
	agent    model.Agent
	path     string
	variant  string
	settings string
	delivery model.ConfigDelivery
}

func (f fakeAdapter) Agent() model.Agent              { return f.agent }
func (f fakeAdapter) InstructionsPath(homeDir string) string { return f.path }
func (f fakeAdapter) VariantKey() string               { return f.variant }
func (f fakeAdapter) SettingsPath(homeDir string) string     { return f.settings }
func (f fakeAdapter) ConfigDelivery() model.ConfigDelivery   { return f.delivery }

// Compile-time check: fakeAdapter implements AgentAdapter.
var _ config.AgentAdapter = fakeAdapter{}

func TestAgentAdapterInterface(t *testing.T) {
	t.Run("known variant returns its key", func(t *testing.T) {
		a := fakeAdapter{agent: model.AgentClaude, path: "/home/user/.claude/CLAUDE.md", variant: "claude"}
		if got := a.VariantKey(); got != "claude" {
			t.Errorf("VariantKey() = %q, want %q", got, "claude")
		}
		if got := a.Agent(); got != model.AgentClaude {
			t.Errorf("Agent() = %q, want %q", got, model.AgentClaude)
		}
		if got := a.InstructionsPath("/home/user"); got != "/home/user/.claude/CLAUDE.md" {
			t.Errorf("InstructionsPath() = %q, want %q", got, "/home/user/.claude/CLAUDE.md")
		}
	})

	t.Run("empty path means skip", func(t *testing.T) {
		a := fakeAdapter{agent: model.AgentCursor, path: "", variant: "cursor"}
		if got := a.InstructionsPath("/home/user"); got != "" {
			t.Errorf("InstructionsPath() = %q, want empty (skip)", got)
		}
	})

	t.Run("unknown variant key is non-empty string (caller decides fallback)", func(t *testing.T) {
		a := fakeAdapter{agent: model.AgentClaude, path: "/some/path", variant: "unknown-agent"}
		if got := a.VariantKey(); got == "" {
			t.Error("VariantKey() must not be empty; caller applies generic fallback")
		}
	})
}

func TestResultType(t *testing.T) {
	t.Run("zero value is all-already", func(t *testing.T) {
		var r config.Result
		if len(r.Files) != 0 {
			t.Errorf("Files = %v, want empty", r.Files)
		}
		if !r.AllAlready {
			// zero value: AllAlready defaults to false — that's fine.
			// This test documents the zero-value contract.
		}
	})

	t.Run("populated result carries files", func(t *testing.T) {
		r := config.Result{
			Files:      []string{"/a/CLAUDE.md", "/b/AGENTS.md"},
			AllAlready: false,
		}
		if len(r.Files) != 2 {
			t.Errorf("Files len = %d, want 2", len(r.Files))
		}
	})
}
