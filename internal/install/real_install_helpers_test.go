package install_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	extinstaller "github.com/Group-Active-IA/active-stack/internal/harness/external"
	skillinstaller "github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// realRegistry wraps the default agent registry the same way main.go does.
type realRegistry struct{ r *agents.Registry }

func (a realRegistry) Get(agent model.Agent) (install.AgentAdapter, bool) {
	ad, ok := a.r.Get(agent)
	if !ok {
		return nil, false
	}
	return ad, true
}

func newRealRegistry(t *testing.T) realRegistry {
	t.Helper()
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	return realRegistry{r: reg}
}

// seedHome writes the files a person already has for each agent: user settings,
// an MCP file that already registers the engram server, and an instructions
// file with their own text. It returns nothing; the files are the baseline.
func seedHome(t *testing.T, home string, reg realRegistry) {
	t.Helper()
	put := func(path, body string) {
		t.Helper()
		if path == "" {
			return
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err == nil {
			return
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex, model.AgentOpenCode, model.AgentAntigravity} {
		ad, _ := reg.Get(agent)
		put(ad.InstructionsPath(home), "# Mis reglas\n\nTexto propio de la persona.\n")
		switch agent {
		case model.AgentClaude:
			put(ad.SettingsPath(home), "{\n  \"theme\": \"dark\"\n}\n")
			put(ad.MCPConfigPath(home, "engram"), "{\n  \"numStartups\": 7,\n  \"mcpServers\": {\n    \"engram\": {\n      \"type\": \"stdio\",\n      \"command\": \"engram\",\n      \"args\": [\n        \"mcp\"\n      ]\n    }\n  }\n}\n")
		case model.AgentCodex:
			put(ad.SettingsPath(home), "model = \"gpt-5\"\n\n[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\"]\n")
		case model.AgentOpenCode:
			put(ad.SettingsPath(home), "{\"theme\":\"opencode\",\"mcp\":{\"engram\":{\"type\":\"local\",\"command\":[\"engram\",\"mcp\"],\"enabled\":true}}}")
		case model.AgentAntigravity:
			put(ad.MCPConfigPath(home, "engram"), "{\"mcpServers\":{\"engram\":{\"command\":\"engram\",\"args\":[\"mcp\"]}}}")
		}
	}
}

// stubNetwork replaces the installers that download or clone, so a real install
// of the configuration harnesses can run in a test.
func stubNetwork(t *testing.T) {
	t.Helper()
	install.WithEmbeddedCommandsFS(assets.CommandsFS)
	t.Cleanup(install.ResetEmbeddedCommandsFS)
	t.Cleanup(install.SetExternalInstallFn(func(context.Context, model.Harness, system.PlatformProfile, []extinstaller.AgentAdapter, string) (extinstaller.Result, error) {
		return extinstaller.Result{AlreadyInstalled: true}, nil
	}))
	t.Cleanup(install.SetSkillInstallFn(func(interface{}, fs.FS, context.Context, model.Harness, []skillinstaller.AgentAdapter, string, string) ([]skillinstaller.Result, error) {
		return nil, nil
	}))
}

// runRealInstall builds and executes the real plan for intent over home.
func runRealInstall(t *testing.T, home string, intent install.Intent) pipeline.ExecutionResult {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := install.BuildPlan(cat, intent, install.Options{
		HomeDir:       home,
		Registry:      newRealRegistry(t),
		NoSelfInstall: true,
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	return pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan)
}

// treeHashes returns relative path -> sha256 of every file under home, with the
// home path normalized out of the contents. The backup store is not part of the
// installed result (it holds timestamps and copies).
func treeHashes(t *testing.T, home string) map[string]string {
	t.Helper()
	out := map[string]string{}
	skip := filepath.Join(home, ".active-stack", "backups")
	err := filepath.WalkDir(home, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p == skip {
				return filepath.SkipDir
			}
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		body := string(raw)
		for _, h := range []string{home, filepath.ToSlash(home), strings.ReplaceAll(home, `\`, `\\`)} {
			body = strings.ReplaceAll(body, h, "<HOME>")
		}
		sum := sha256.Sum256([]byte(body))
		rel, _ := filepath.Rel(home, p)
		out[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
