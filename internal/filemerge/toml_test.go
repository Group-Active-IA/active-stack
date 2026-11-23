package filemerge

import (
	"strings"
	"testing"
)

// ─── UpsertCodexEngramBlock ───────────────────────────────────────────────────

func TestUpsertCodexEngramBlock_Empty(t *testing.T) {
	result := UpsertCodexEngramBlock("", "")

	if !strings.Contains(result, "[mcp_servers.engram]") {
		t.Fatalf("result missing [mcp_servers.engram]; got:\n%s", result)
	}
	if !strings.Contains(result, `command = "engram"`) {
		t.Fatalf("result missing command = \"engram\"; got:\n%s", result)
	}
	if !strings.Contains(result, `"--tools=agent"`) {
		t.Fatalf("result missing --tools=agent; got:\n%s", result)
	}
	if !strings.Contains(result, `args = ["mcp", "--tools=agent"]`) {
		t.Fatalf("result has wrong args format; got:\n%s", result)
	}
}

func TestUpsertCodexEngramBlock_ExistingBlock(t *testing.T) {
	input := `[other_section]
key = "value"

[mcp_servers.engram]
command = "engram"
args = ["mcp"]

[another_section]
foo = "bar"
`
	result := UpsertCodexEngramBlock(input, "")

	// Must have exactly one [mcp_servers.engram] block.
	count := strings.Count(result, "[mcp_servers.engram]")
	if count != 1 {
		t.Fatalf("expected 1 [mcp_servers.engram] block, got %d; result:\n%s", count, result)
	}

	// Must preserve unrelated sections.
	if !strings.Contains(result, "[other_section]") {
		t.Fatalf("result missing [other_section]; got:\n%s", result)
	}
	if !strings.Contains(result, "[another_section]") {
		t.Fatalf("result missing [another_section]; got:\n%s", result)
	}

	// Must use the updated args with --tools=agent.
	if !strings.Contains(result, `"--tools=agent"`) {
		t.Fatalf("result missing --tools=agent; got:\n%s", result)
	}
}

func TestUpsertCodexEngramBlock_PreservesOtherSections(t *testing.T) {
	input := `model = "gpt-4o"

[settings]
timeout = 30
`
	result := UpsertCodexEngramBlock(input, "")

	if !strings.Contains(result, `model = "gpt-4o"`) {
		t.Fatalf("result missing top-level model key; got:\n%s", result)
	}
	if !strings.Contains(result, "[settings]") {
		t.Fatalf("result missing [settings] section; got:\n%s", result)
	}
	if !strings.Contains(result, "[mcp_servers.engram]") {
		t.Fatalf("result missing [mcp_servers.engram]; got:\n%s", result)
	}
}

func TestUpsertCodexEngramBlock_AbsolutePath(t *testing.T) {
	result := UpsertCodexEngramBlock("", "/usr/local/bin/engram")

	if !strings.Contains(result, "[mcp_servers.engram]") {
		t.Fatalf("result missing [mcp_servers.engram]; got:\n%s", result)
	}
	if !strings.Contains(result, `command = "/usr/local/bin/engram"`) {
		t.Fatalf("result missing absolute command path; got:\n%s", result)
	}
	if strings.Contains(result, `command = "engram"`) {
		t.Fatalf("result should NOT have relative command when absolute path given; got:\n%s", result)
	}
}

func TestUpsertCodexEngramBlock_Idempotent(t *testing.T) {
	input := `[other]
key = "val"
`
	first := UpsertCodexEngramBlock(input, "")
	second := UpsertCodexEngramBlock(first, "")

	if first != second {
		t.Fatalf("UpsertCodexEngramBlock is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	count := strings.Count(second, "[mcp_servers.engram]")
	if count != 1 {
		t.Fatalf("after two runs: expected 1 [mcp_servers.engram] block, got %d; result:\n%s", count, second)
	}
}

func TestUpsertCodexEngramBlockWindowsPath(t *testing.T) {
	// Windows paths contain backslashes which must be escaped in TOML double-quoted strings.
	// \U would be interpreted as a Unicode escape sequence → parse error.
	windowsCmd := `C:\Users\PERC\AppData\Local\engram\bin\engram.exe`
	result := UpsertCodexEngramBlock("", windowsCmd)

	// TOML double-quoted string must have double backslashes.
	want := `command = "C:\\Users\\PERC\\AppData\\Local\\engram\\bin\\engram.exe"`
	if !strings.Contains(result, want) {
		t.Fatalf("result missing properly escaped Windows path;\nwant substring: %s\ngot:\n%s", want, result)
	}
}

// ─── UpsertTopLevelTOMLString ─────────────────────────────────────────────────

func TestUpsertTopLevelTOMLString_NewKey(t *testing.T) {
	input := `[mcp_servers.engram]
command = "engram"
`
	result := UpsertTopLevelTOMLString(input, "model_instructions_file", "/home/user/.codex/instructions.md")

	if !strings.Contains(result, `model_instructions_file = "/home/user/.codex/instructions.md"`) {
		t.Fatalf("result missing model_instructions_file key; got:\n%s", result)
	}
	// Must appear before the first [section].
	idx := strings.Index(result, "model_instructions_file")
	sectionIdx := strings.Index(result, "[mcp_servers.engram]")
	if idx > sectionIdx {
		t.Fatalf("model_instructions_file should appear before [mcp_servers.engram]; got:\n%s", result)
	}
}

func TestUpsertTopLevelTOMLString_ReplaceKey(t *testing.T) {
	input := `model_instructions_file = "/old/path.md"

[mcp_servers.engram]
command = "engram"
`
	result := UpsertTopLevelTOMLString(input, "model_instructions_file", "/new/path.md")

	if !strings.Contains(result, `model_instructions_file = "/new/path.md"`) {
		t.Fatalf("result missing updated value; got:\n%s", result)
	}
	if strings.Contains(result, "/old/path.md") {
		t.Fatalf("result still has old value; got:\n%s", result)
	}
	count := strings.Count(result, "model_instructions_file")
	if count != 1 {
		t.Fatalf("expected 1 model_instructions_file, got %d; result:\n%s", count, result)
	}
}

func TestUpsertTopLevelTOMLString_Idempotent(t *testing.T) {
	input := `[mcp_servers.engram]
command = "engram"
`
	first := UpsertTopLevelTOMLString(input, "model_instructions_file", "/path/instructions.md")
	second := UpsertTopLevelTOMLString(first, "model_instructions_file", "/path/instructions.md")

	if first != second {
		t.Fatalf("UpsertTopLevelTOMLString is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// R-23/PP-03: the comments that sit right before the next header belong to the
// next block (for example a managed-block marker); removing a table must not
// take them with it.
func TestRemoveTOMLSection_KeepsCommentsOfTheNextBlock(t *testing.T) {
	in := "[mcp_servers.playwright]\ncommand = \"npx\"\nargs = [\"-y\"]\n\n# active-stack:memory-hooks begin\n# Managed by active-stack\n[[hooks.SessionStart]]\ntype = \"command\"\n# active-stack:memory-hooks end\n"
	got, err := RemoveTOMLSection(in, "mcp_servers.playwright")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# active-stack:memory-hooks begin", "# Managed by active-stack", "[[hooks.SessionStart]]", "# active-stack:memory-hooks end"} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "playwright") || strings.Contains(got, "npx") {
		t.Errorf("table not removed:\n%s", got)
	}
}

func TestRemoveTOMLSection_DropsCommentsInsideTheTable(t *testing.T) {
	in := "model = \"x\"\n\n[mcp_servers.a]\n# note about a\ncommand = \"c\"\n# trailing comment of a\n\n[other]\nk = 1\n"
	got, err := RemoveTOMLSection(in, "mcp_servers.a")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "note about a") || strings.Contains(got, "command") {
		t.Errorf("table body kept:\n%s", got)
	}
	if !strings.Contains(got, "[other]") || !strings.Contains(got, `model = "x"`) {
		t.Errorf("neighbours lost:\n%s", got)
	}
}

func TestRemoveTOMLSection_TableAtEndOfFile(t *testing.T) {
	in := "a = 1\n\n[mcp_servers.z]\ncommand = \"c\"\n# trailing\n"
	got, err := RemoveTOMLSection(in, "mcp_servers.z")
	if err != nil || strings.TrimSpace(got) != "a = 1" {
		t.Fatalf("got %q err %v", got, err)
	}
}
