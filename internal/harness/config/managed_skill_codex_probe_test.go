package config_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"gopkg.in/yaml.v3"
)

// This opt-in probe uses the installed client for discovery only. It does not
// start an inference turn or establish activation, permissions or IDE behavior.
func TestManagedSkillCodexDiscovery(t *testing.T) {
	if os.Getenv("ACTIVE_STACK_CODEX_PROBE") != "1" {
		t.Skip("set ACTIVE_STACK_CODEX_PROBE=1 to probe an installed Codex client")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	version, err := exec.CommandContext(ctx, binary, "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("discovery client: %s", strings.TrimSpace(string(version)))

	fixture := t.TempDir()
	adapter := codex.NewAdapter()
	harness := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"delegation", "engram", "governance", "tdd", "model-routing"}}
	if _, err := config.InstallManagedSkill(harness, adapter, fixture); err != nil {
		t.Fatal(err)
	}
	skillRoot := filepath.Join(adapter.SkillsDir(fixture), harness.ID)
	resolvedRoot, err := filepath.EvalSymlinks(skillRoot)
	if err != nil {
		t.Fatal(err)
	}
	policyBytes, err := os.ReadFile(filepath.Join(skillRoot, "agents", "openai.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var metadata struct {
		Policy struct {
			AllowImplicitInvocation *bool `yaml:"allow_implicit_invocation"`
		} `yaml:"policy"`
	}
	if err := yaml.Unmarshal(policyBytes, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.Policy.AllowImplicitInvocation == nil || *metadata.Policy.AllowImplicitInvocation {
		t.Fatal("installed policy must explicitly disable implicit invocation")
	}

	command := exec.CommandContext(ctx, binary, "app-server", "--stdio")
	command.Dir = fixture
	command.Stderr = io.Discard
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stdin.Close()
		cancel()
		_ = command.Wait()
	}()
	responses := make(chan map[string]json.RawMessage, 4)
	scanErrors := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 4096), 8*1024*1024)
		for scanner.Scan() {
			var message map[string]json.RawMessage
			if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
				scanErrors <- err
				return
			}
			select {
			case responses <- message:
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			scanErrors <- err
		} else {
			scanErrors <- io.EOF
		}
	}()
	encoder := json.NewEncoder(stdin)
	send := func(message map[string]any) {
		t.Helper()
		if err := encoder.Encode(message); err != nil {
			t.Fatal(err)
		}
	}
	readResult := func(id int) json.RawMessage {
		t.Helper()
		for {
			select {
			case <-ctx.Done():
				t.Fatalf("Codex response %d timed out", id)
			case err := <-scanErrors:
				t.Fatalf("Codex protocol ended: %v", err)
			case message := <-responses:
				if string(message["id"]) != fmt.Sprint(id) {
					continue
				}
				if errorData, ok := message["error"]; ok && string(errorData) != "null" {
					t.Fatalf("Codex rejected request %d", id)
				}
				return message["result"]
			}
		}
	}
	send(map[string]any{"id": 1, "method": "initialize", "params": map[string]any{"clientInfo": map[string]string{"name": "active_stack_delivery_probe", "version": "0.1"}}})
	readResult(1)
	send(map[string]any{"method": "initialized", "params": map[string]any{}})
	send(map[string]any{"id": 2, "method": "skills/list", "params": map[string]any{"cwds": []string{fixture}, "forceReload": true}})
	var result any
	if err := json.Unmarshal(readResult(2), &result); err != nil {
		t.Fatal(err)
	}
	found := false
	var inspect func(any)
	inspect = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			path, _ := value["path"].(string)
			name, _ := value["name"].(string)
			if name != "" && strings.HasPrefix(path, resolvedRoot+string(filepath.Separator)) {
				if name != harness.ID || filepath.Clean(path) != filepath.Join(resolvedRoot, "SKILL.md") {
					t.Errorf("support file was discovered as an additional skill")
				} else {
					found = true
				}
			}
			for _, item := range value {
				inspect(item)
			}
		case []any:
			for _, item := range value {
				inspect(item)
			}
		}
	}
	inspect(result)
	if !found {
		t.Fatal("installed sdd-orchestrator skill was absent from fixture discovery")
	}
}
