package memoryproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// The child is this test binary run as a fake MCP server (see TestMain).
func TestMain(m *testing.M) {
	switch os.Getenv("MEMORYPROXY_CHILD") {
	case "cat":
		// echo stdin to stdout byte for byte
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			os.Stdout.Write(buf[:n])
			if err != nil {
				os.Exit(0)
			}
		}
	case "exit3":
		os.Stderr.WriteString("child boom\n")
		os.Exit(3)
	case "hello-exit":
		os.Stdout.WriteString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{}}\r\n")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func run(t *testing.T, mode, input string, prep PrepareFunc) (code int, out, errOut string) {
	t.Helper()
	t.Setenv("MEMORYPROXY_CHILD", mode)
	var o, e bytes.Buffer
	code = Run(context.Background(), Options{
		Child: []string{os.Args[0]}, Dir: "proj", Prepare: prep,
		Stdin: strings.NewReader(input), Stdout: &o, Stderr: &e, Budget: 2 * time.Second,
	})
	return code, o.String(), e.String()
}

// recorder tags like the real Prepare would: it adds a header and records calls.
type recorder struct {
	mu    sync.Mutex
	calls []string
	err   error
	sleep time.Duration
}

func (r *recorder) prepare(dir, tool string, in map[string]any) (map[string]any, bool, error) {
	r.mu.Lock()
	r.calls = append(r.calls, dir+"|"+tool)
	r.mu.Unlock()
	if r.sleep > 0 {
		time.Sleep(r.sleep)
	}
	if r.err != nil {
		return in, false, r.err
	}
	out := map[string]any{}
	for k, v := range in {
		out[k] = v
	}
	out["content"] = "Author: T\n" + in["content"].(string)
	return out, true, nil
}

func TestPassThroughUntouched(t *testing.T) {
	r := &recorder{}
	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\r\n",
		`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n",
		`{"jsonrpc":"2.0","id":"abc","method":"tools/call","params":{"name":"mem_search","arguments":{"query":"x"}}}` + "\r\n",
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"mem_update","arguments":{"id":1}}}` + "\n",
		`[{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"mem_save","arguments":{"content":"a"}}}]` + "\n",
		`not json at all` + "\n",
		`{"jsonrpc":"2.0","id":9007199254740993,"method":"tools/call","params":{"name":"other","arguments":null}}` + "\n",
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"mem_save"}}` + "\n",
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"mem_save","arguments":"oops"}}` + "\n",
		`{"jsonrpc":"2.0","id":7,"method":"tools/call"}` + "\n",
		"\n",
		`{"jsonrpc":"2.0","id":8,"method":"ping"}`, // no final newline
	}
	input := strings.Join(lines, "")
	code, out, _ := run(t, "cat", input, r.prepare)
	if code != 0 || out != input {
		t.Fatalf("code %d, output differs:\n got %q\nwant %q", code, out, input)
	}
	if len(r.calls) != 0 {
		t.Fatalf("Prepare called for untagged traffic: %v", r.calls)
	}
}

func TestTagsSaveTools(t *testing.T) {
	for _, tool := range []string{"mem_save", "mem_session_summary", "mem_session_finalize", "mem_capture_passive", "mem_save_prompt"} {
		for _, eol := range []string{"\n", "\r\n"} {
			name := tool + "/lf"
			if eol == "\r\n" {
				name = tool + "/crlf"
			}
			t.Run(name, func(t *testing.T) {
				r := &recorder{}
				in := `{"jsonrpc":"2.0","id":4294967296,"method":"tools/call","params":{"name":"` + tool +
					`","_meta":{"progressToken":7},"arguments":{"content":"a & <b>","big":12345678901234567890,"session_id":"s1"}}}` + eol
				code, out, _ := run(t, "cat", in, r.prepare)
				if code != 0 {
					t.Fatal(code)
				}
				if !strings.HasSuffix(out, eol) || (eol == "\n" && strings.HasSuffix(out, "\r\n")) {
					t.Fatalf("line ending not preserved: %q", out)
				}
				var msg struct {
					ID     json.Number
					Method string
					Params struct {
						Name string
						Meta map[string]any `json:"_meta"`
						Args map[string]any `json:"arguments"`
					}
				}
				dec := json.NewDecoder(strings.NewReader(out))
				dec.UseNumber()
				if err := dec.Decode(&msg); err != nil {
					t.Fatal(err)
				}
				if msg.ID.String() != "4294967296" || msg.Method != "tools/call" || msg.Params.Name != tool {
					t.Fatalf("envelope changed: %+v", msg)
				}
				if msg.Params.Args["content"] != "Author: T\na & <b>" {
					t.Fatalf("content = %q", msg.Params.Args["content"])
				}
				if msg.Params.Args["session_id"] != "s1" || msg.Params.Meta["progressToken"] == nil {
					t.Fatalf("other fields lost: %+v", msg.Params)
				}
				if !strings.Contains(out, "12345678901234567890") || !strings.Contains(out, "a & <b>") {
					t.Fatalf("numbers or characters re-encoded: %q", out)
				}
				if len(r.calls) != 1 || r.calls[0] != "proj|"+tool {
					t.Fatalf("calls = %v", r.calls)
				}
			})
		}
	}
}

func TestFailuresForwardOriginal(t *testing.T) {
	in := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"mem_save","arguments":{"content":"a"}}}` + "\r\n"
	cases := map[string]*recorder{
		"error":   {err: errors.New("git missing")},
		"timeout": {sleep: 400 * time.Millisecond},
	}
	for name, r := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv("MEMORYPROXY_CHILD", "cat")
			var o, e bytes.Buffer
			code := Run(context.Background(), Options{
				Child: []string{os.Args[0]}, Dir: "p", Prepare: r.prepare,
				Stdin: strings.NewReader(in), Stdout: &o, Stderr: &e, Budget: 100 * time.Millisecond,
			})
			if code != 0 || o.String() != in {
				t.Fatalf("code %d, out %q", code, o.String())
			}
			if !strings.Contains(e.String(), "tagging skipped") {
				t.Fatalf("stderr = %q", e.String())
			}
		})
	}
}

func TestNotChangedForwardsOriginalBytes(t *testing.T) {
	in := `{ "jsonrpc":"2.0", "id":1, "method":"tools/call", "params":{"name":"mem_save","arguments":{"content":"a"}} }` + "\r\n"
	noop := func(_, _ string, m map[string]any) (map[string]any, bool, error) { return m, false, nil }
	code, out, _ := run(t, "cat", in, noop)
	if code != 0 || out != in {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestChildExitCodeAndStderr(t *testing.T) {
	code, _, errOut := run(t, "exit3", strings.Repeat(`{"jsonrpc":"2.0","id":1,"method":"ping"}`+"\n", 3), nil)
	if code != 3 {
		t.Fatalf("code = %d, want 3", code)
	}
	if !strings.Contains(errOut, "child boom") {
		t.Fatalf("child stderr not forwarded: %q", errOut)
	}
}

func TestChildExitsWhileClientStillOpen(t *testing.T) {
	t.Setenv("MEMORYPROXY_CHILD", "hello-exit")
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pw.Close()
	defer pr.Close()
	var o bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- Run(context.Background(), Options{Child: []string{os.Args[0]}, Dir: "p",
			Stdin: pr, Stdout: &o, Stderr: &bytes.Buffer{}})
	}()
	select {
	case code := <-done:
		if code != 0 || o.String() != "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{}}\r\n" {
			t.Fatalf("code %d out %q", code, o.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("proxy did not return after the child exited")
	}
}

func TestChildNotFound(t *testing.T) {
	var o, e bytes.Buffer
	code := Run(context.Background(), Options{Child: []string{"definitely-not-a-real-binary-xyz"}, Dir: "p",
		Stdin: strings.NewReader(""), Stdout: &o, Stderr: &e})
	if code != 1 || e.Len() == 0 || o.Len() != 0 {
		t.Fatalf("code %d out %q err %q", code, o.String(), e.String())
	}
}

func TestNoChildCommand(t *testing.T) {
	var e bytes.Buffer
	if code := Run(context.Background(), Options{Stdin: strings.NewReader(""), Stdout: &bytes.Buffer{}, Stderr: &e}); code != 2 {
		t.Fatalf("code %d", code)
	}
}

func TestLargeLine(t *testing.T) {
	big := strings.Repeat("x", 3<<20)
	in := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"mem_search","arguments":{"query":"` + big + `"}}}` + "\n"
	code, out, _ := run(t, "cat", in, nil)
	if code != 0 || out != in {
		t.Fatalf("code %d, len %d vs %d", code, len(out), len(in))
	}
}
