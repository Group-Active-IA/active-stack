package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestMain lets this test binary act as a fake Engram MCP server that echoes
// its input, so the proxy command is tested over real pipes.
func TestMain(m *testing.M) {
	if os.Getenv("ACTIVE_STACK_FAKE_MCP") == "echo" {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			os.Stdout.Write(buf[:n])
			if err != nil {
				os.Exit(0)
			}
		}
	}
	os.Exit(m.Run())
}

func runProxy(t *testing.T, input string, args ...string) (int, string, string) {
	t.Helper()
	old := memoryStdin
	memoryStdin = strings.NewReader(input)
	defer func() { memoryStdin = old }()
	var stdout, stderr bytes.Buffer
	code := runMemoryDispatch(append([]string{"proxy"}, args...), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestProxyUsageErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"no separator":   {"engram", "mcp"},
		"empty command":  {"--project", ".", "--"},
		"no args at all": {},
		"bad flag":       {"--nope", "--", "engram"},
	} {
		t.Run(name, func(t *testing.T) {
			code, out, errOut := runProxy(t, "", args...)
			if code != memoryExitUsage || out != "" || errOut == "" {
				t.Fatalf("code %d out %q err %q", code, out, errOut)
			}
		})
	}
}

// With the marker on a branch, a save through the real command comes out
// tagged; every other message is echoed as is.
func TestProxyTagsSaveEndToEnd(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	t.Setenv("ACTIVE_STACK_FAKE_MCP", "echo")
	save := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"mem_save","arguments":{"title":"T","content":"c","topic_key":"k"}}}`
	ping := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	code, out, errOut := runProxy(t, ping+"\r\n"+save+"\r\n", "--project", dir, "--", os.Args[0])
	if code != 0 {
		t.Fatalf("exit %d stderr %q", code, errOut)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\r\n"), "\r\n")
	if len(lines) != 2 || lines[0] != ping {
		t.Fatalf("output = %q", out)
	}
	var msg struct {
		Params struct {
			Arguments map[string]string `json:"arguments"`
		} `json:"params"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &msg); err != nil {
		t.Fatal(err)
	}
	a := msg.Params.Arguments
	if a["topic_key"] != "branch/feature-x-217d2b/k" || a["title"] != "[feature-x-217d2b] T" ||
		!strings.HasPrefix(a["content"], "Branch: feature/x · Author: ") {
		t.Fatalf("not tagged: %v", a)
	}
}

// Without the marker the save goes through byte for byte.
func TestProxyNoMarkerUntouched(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ACTIVE_STACK_FAKE_MCP", "echo")
	save := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"mem_save","arguments":{"title":"T","content":"c"}}}` + "\n"
	code, out, _ := runProxy(t, save, "--project", dir, "--", os.Args[0])
	if code != 0 || out != save {
		t.Fatalf("code %d out %q", code, out)
	}
}
