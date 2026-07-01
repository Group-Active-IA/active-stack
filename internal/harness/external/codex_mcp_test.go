package external

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestWriteCodexMCPProjectEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("model = \"keep\"\n\n[mcp_servers.other]\ncommand = \"other\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mcp := model.MCP{Name: "engram", Command: `C:\bin\engram.exe`, Args: []string{"mcp", "--tools=agent"}}
	changed, err := WriteCodexMCPProjectEntry(mcp, path, filepath.Join(t.TempDir(), "backup"))
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed")
	}
	raw, _ := os.ReadFile(path)
	got := string(raw)
	for _, want := range []string{`model = "keep"`, "[mcp_servers.other]", "[mcp_servers.engram]", `"--tools=agent"`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
	changed, err = WriteCodexMCPProjectEntry(mcp, path, filepath.Join(t.TempDir(), "backup2"))
	if err != nil || changed {
		t.Fatalf("second write changed=%v err=%v", changed, err)
	}
}
