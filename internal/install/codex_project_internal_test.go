package install

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

func TestProjectTargetAdapterUsesCodexProjectLayout(t *testing.T) {
	root := t.TempDir()
	adapter := projectTargetAdapter{AgentAdapter: codex.NewAdapter(), base: root}

	if got, want := adapter.InstructionsPath(root), filepath.Join(root, "AGENTS.md"); got != want {
		t.Errorf("InstructionsPath = %q, want %q", got, want)
	}
	if got, want := adapter.SettingsPath(root), filepath.Join(root, ".codex", "config.toml"); got != want {
		t.Errorf("SettingsPath = %q, want %q", got, want)
	}
	if got := adapter.MCPStrategy(); got != external.StrategyMergeIntoTOML {
		t.Errorf("MCPStrategy = %v, want TOML", got)
	}
}

func TestCodexProjectDiagnosticDoesNotDegradeProgress(t *testing.T) {
	stderr, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	previous := os.Stderr
	os.Stderr = stderr
	defer func() { os.Stderr = previous }()
	var events []pipeline.ProgressEvent
	step, err := buildHarnessStep(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, []AgentAdapter{codex.NewAdapter()}, Options{Target: model.Project, OnProgress: func(e pipeline.ProgressEvent) { events = append(events, e) }}, t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if err = step.Run(); err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("supported Project delivery must not emit degraded progress: %+v", events)
	}
	if _, err = stderr.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	diagnostic, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(diagnostic), "Codex Project") || !strings.Contains(string(diagnostic), "legacy") {
		t.Fatalf("missing stderr diagnostic: %q", diagnostic)
	}
}
