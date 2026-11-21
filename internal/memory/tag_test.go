package memory

import (
	"os/user"
	"reflect"
	"strings"
	"testing"
)

var (
	onBranch  = SaveContext{Result: Result{Mode: ModeCloudBranch, Tagged: true, Branch: "feat/login", Slug: "feat-login"}, Author: "Ana Pérez"}
	onPrimary = SaveContext{Result: Result{Mode: ModeCloudBranch, Branch: "main"}, Author: "Ana Pérez"}
	modeOff   = SaveContext{Result: Result{Mode: ModeOff, Branch: "feat/login"}, Author: "Ana Pérez"}
)

func saveInput() map[string]any {
	return map[string]any{
		"title":          "Arreglé el token",
		"topic_key":      "auth/token",
		"content":        "El token vencía antes de tiempo.",
		"type":           "bugfix",
		"scope":          "project",
		"project":        "demo",
		"session_id":     "s-123",
		"capture_prompt": false,
	}
}

func tag(t *testing.T, tool string, in map[string]any, ctx SaveContext) (map[string]any, bool) {
	t.Helper()
	before := copyMap(in)
	out, changed := TagInput(tool, in, ctx)
	if !reflect.DeepEqual(in, before) {
		t.Fatalf("TagInput modified its input: %v", in)
	}
	return out, changed
}

func copyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// T-F5-1: without a marker (mode off) nothing changes.
func TestTagOffChangesNothing(t *testing.T) {
	in := saveInput()
	out, changed := tag(t, "mcp__engram__mem_save", in, modeOff)
	if changed || !reflect.DeepEqual(out, in) {
		t.Fatalf("mode off: changed=%v out=%v", changed, out)
	}
}

// T-F5-1b: on the primary branch only the Author line is added.
func TestTagPrimaryAddsOnlyAuthor(t *testing.T) {
	out, changed := tag(t, "mcp__engram__mem_save", saveInput(), onPrimary)
	want := saveInput()
	want["content"] = "Author: Ana Pérez\nEl token vencía antes de tiempo."
	if !changed || !reflect.DeepEqual(out, want) {
		t.Fatalf("got changed=%v %v\nwant %v", changed, out, want)
	}
}

// T-F5-2 and T-F5-3: on a branch topic_key, title and the first content line
// are rewritten and every other field is kept.
func TestTagBranchRewritesKeyTitleAndContent(t *testing.T) {
	out, changed := tag(t, "mcp__engram__mem_save", saveInput(), onBranch)
	want := saveInput()
	want["topic_key"] = "branch/feat-login/auth/token"
	want["title"] = "[feat-login] Arreglé el token"
	want["content"] = "Branch: feat/login · Author: Ana Pérez\nEl token vencía antes de tiempo."
	if !changed || !reflect.DeepEqual(out, want) {
		t.Fatalf("got changed=%v %v\nwant %v", changed, out, want)
	}
}

// T-F5-3: unknown fields travel unchanged too.
func TestTagKeepsUnknownFields(t *testing.T) {
	in := saveInput()
	in["future_field"] = map[string]any{"a": 1.0}
	out, _ := tag(t, "mem_save", in, onBranch)
	if !reflect.DeepEqual(out["future_field"], in["future_field"]) {
		t.Fatalf("future_field = %v", out["future_field"])
	}
}

// T-F5-4: tagging an already tagged memory changes nothing.
func TestTagIsIdempotent(t *testing.T) {
	for name, ctx := range map[string]SaveContext{"branch": onBranch, "primary": onPrimary} {
		t.Run(name, func(t *testing.T) {
			once, _ := tag(t, "mem_save", saveInput(), ctx)
			twice, changed := tag(t, "mem_save", once, ctx)
			if changed || !reflect.DeepEqual(twice, once) {
				t.Fatalf("second pass changed=%v %v", changed, twice)
			}
		})
	}
}

// T-F5-5: a missing or empty topic_key stays that way; title and content are tagged.
func TestTagMissingTopicKeyStaysEmpty(t *testing.T) {
	in := saveInput()
	delete(in, "topic_key")
	out, _ := tag(t, "mem_save", in, onBranch)
	if _, ok := out["topic_key"]; ok {
		t.Fatalf("topic_key was invented: %v", out["topic_key"])
	}
	if out["title"] != "[feat-login] Arreglé el token" || !strings.HasPrefix(out["content"].(string), "Branch: feat/login") {
		t.Fatalf("title/content not tagged: %v", out)
	}
	in["topic_key"] = ""
	out, _ = tag(t, "mem_save", in, onBranch)
	if out["topic_key"] != "" {
		t.Fatalf("empty topic_key became %v", out["topic_key"])
	}
}

// T-F5-6: the observation alias of content is tagged.
func TestTagObservationAlias(t *testing.T) {
	in := saveInput()
	delete(in, "content")
	in["observation"] = "Texto"
	out, _ := tag(t, "mem_save", in, onBranch)
	if out["observation"] != "Branch: feat/login · Author: Ana Pérez\nTexto" {
		t.Fatalf("observation = %q", out["observation"])
	}
	if _, ok := out["content"]; ok {
		t.Fatalf("content was invented")
	}
}

// Both Engram MCP prefixes and the bare name (proxy) are recognised; other tools are not touched.
func TestTagToolNames(t *testing.T) {
	for _, tool := range []string{"mem_save", "mcp__engram__mem_save", "mcp__plugin_engram_engram__mem_save"} {
		if _, changed := tag(t, tool, saveInput(), onBranch); !changed {
			t.Errorf("%s was not tagged", tool)
		}
	}
	for _, tool := range []string{"mem_search", "mcp__other__mem_save", "Bash", "mem_update"} {
		if _, changed := tag(t, tool, saveInput(), onBranch); changed {
			t.Errorf("%s was tagged", tool)
		}
	}
}

// R15-2: narrative routes get the header in their text field, without key or title.
func TestTagNarrativeRoutes(t *testing.T) {
	cases := map[string]string{"mem_session_summary": "content", "mem_session_finalize": "summary", "mem_save_prompt": "content"}
	for tool, field := range cases {
		t.Run(tool, func(t *testing.T) {
			in := map[string]any{field: "Terminé el login", "session_id": "s-1", "title": "x"}
			out, changed := tag(t, "mcp__engram__"+tool, in, onBranch)
			if !changed || out[field] != "Branch: feat/login · Author: Ana Pérez\nTerminé el login" || out["title"] != "x" || out["session_id"] != "s-1" {
				t.Fatalf("branch: changed=%v %v", changed, out)
			}
			out, _ = tag(t, tool, in, onPrimary)
			if out[field] != "Author: Ana Pérez\nTerminé el login" {
				t.Fatalf("primary: %v", out[field])
			}
			again, changed := tag(t, tool, out, onPrimary)
			if changed || !reflect.DeepEqual(again, out) {
				t.Fatalf("not idempotent: %v", again)
			}
			if _, changed := tag(t, tool, map[string]any{field: 3.0}, onBranch); changed {
				t.Fatalf("non-string %s was tagged", field)
			}
		})
	}
}

// R15-2: passive capture tags each list item on a branch and nothing on the primary branch.
func TestTagPassiveCapture(t *testing.T) {
	in := map[string]any{"content": "## Aprendizajes\n- uno\n* dos\n1. tres\n2) cuatro\ntexto suelto", "session_id": "s-1"}
	out, changed := tag(t, "mem_capture_passive", in, onBranch)
	want := "## Aprendizajes\n- [feat-login] uno\n* [feat-login] dos\n1. [feat-login] tres\n2) [feat-login] cuatro\ntexto suelto"
	if !changed || out["content"] != want {
		t.Fatalf("got changed=%v %q", changed, out["content"])
	}
	if again, changed := tag(t, "mem_capture_passive", out, onBranch); changed || again["content"] != want {
		t.Fatalf("not idempotent: %q", again["content"])
	}
	if _, changed := tag(t, "mem_capture_passive", in, onPrimary); changed {
		t.Fatalf("primary branch tagged passive capture")
	}
}

// T-F5-7: the author is cleaned: one line, no email, bounded length.
func TestSanitizeAuthor(t *testing.T) {
	cases := map[string]string{
		"  Ana   Pérez  ":        "Ana Pérez",
		"Ana\nPérez\r\tX":        "Ana Pérez X",
		"ana@example.com":        "",
		"Ana <ana@example.com>":  "",
		`CORP\ana`:               "ana",
		strings.Repeat("a", 100): strings.Repeat("a", 64),
		"Ana · Pérez":            "Ana Pérez",
		"":                       "",
	}
	for in, want := range cases {
		if got := SanitizeAuthor(in); got != want {
			t.Errorf("SanitizeAuthor(%q) = %q, want %q", in, got, want)
		}
	}
}

// T-F5-7: user.name of the repo; without it, the system user; never the email.
func TestResolveAuthor(t *testing.T) {
	dir := newRepo(t, true)
	git(t, dir, "config", "user.email", "ana@example.com")
	current, err := user.Current()
	if err != nil {
		t.Skip("no system user:", err)
	}
	if got, want := ResolveAuthor(dir), SanitizeAuthor(current.Username); got != want || got == "" {
		t.Fatalf("without user.name: %q, want system user %q", got, want)
	}
	git(t, dir, "config", "user.name", "Ana Pérez")
	if got := ResolveAuthor(dir); got != "Ana Pérez" {
		t.Fatalf("with user.name: %q", got)
	}
	git(t, dir, "config", "user.name", "ana@example.com")
	if got := ResolveAuthor(dir); strings.Contains(got, "@") || got == "" {
		t.Fatalf("email used as author: %q", got)
	}
}

// T-F5-1, T-F5-1b, T-F5-2 against real repositories.
func TestPrepareSave(t *testing.T) {
	t.Run("no marker", func(t *testing.T) {
		dir := newRepo(t, false)
		addOrigin(t, dir, "main")
		git(t, dir, "switch", "-q", "-c", "feat/login")
		out, changed, err := PrepareSave(dir, "mem_save", saveInput())
		if err != nil || changed || !reflect.DeepEqual(out, saveInput()) {
			t.Fatalf("err=%v changed=%v %v", err, changed, out)
		}
	})
	t.Run("primary", func(t *testing.T) {
		dir := newRepo(t, true)
		addOrigin(t, dir, "main")
		git(t, dir, "config", "user.name", "Ana Pérez")
		out, changed, err := PrepareSave(dir, "mem_save", saveInput())
		if err != nil || !changed || out["content"] != "Author: Ana Pérez\nEl token vencía antes de tiempo." || out["title"] != "Arreglé el token" {
			t.Fatalf("err=%v changed=%v %v", err, changed, out)
		}
	})
	t.Run("branch", func(t *testing.T) {
		dir := newRepo(t, true)
		addOrigin(t, dir, "main")
		git(t, dir, "config", "user.name", "Ana Pérez")
		git(t, dir, "switch", "-q", "-c", "feat/login")
		out, changed, err := PrepareSave(dir, "mcp__engram__mem_save", saveInput())
		if err != nil || !changed || out["topic_key"] != "branch/feat-login/auth/token" || out["title"] != "[feat-login] Arreglé el token" {
			t.Fatalf("err=%v changed=%v %v", err, changed, out)
		}
	})
}

// T-F5-10 and T-F5-11: context lines for the hooks.
func TestContextLines(t *testing.T) {
	if got := SessionContext(onBranch.Result); got != "Team memory: cloud-branch. Current branch: feat/login (slug: feat-login)." {
		t.Errorf("session on branch: %q", got)
	}
	if got := SessionContext(onPrimary.Result); got != "Team memory: cloud-branch. Current branch: main (untagged)." {
		t.Errorf("session on primary: %q", got)
	}
	if got := SessionContext(modeOff.Result); got != "" {
		t.Errorf("session off: %q", got)
	}
	if got := ReadContext(onBranch.Result); got != "Current branch: feat/login (slug: feat-login)." {
		t.Errorf("read on branch: %q", got)
	}
	if got := ReadContext(onPrimary.Result); got != "" {
		t.Errorf("read on primary: %q", got)
	}
	if got := ReadContext(modeOff.Result); got != "" {
		t.Errorf("read off: %q", got)
	}
}
