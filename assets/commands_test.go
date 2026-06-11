package assets_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/assets"
)

// TestCommandsFS_ClaudeVariant_ResolvesAndContainsActiveStackInvocation asserts that:
// - the embedded CommandsFS contains the Claude variant at commands/claude/active/starter-add.md
// - the body contains "active-stack starter add" (thin wrapper requirement)
// - the body contains "$ARGUMENTS" (the confirmed argument token from TBD-1)
// RED: fails until CommandsFS and the Claude asset file are added (4.2).
func TestCommandsFS_ClaudeVariant_ResolvesAndContainsActiveStackInvocation(t *testing.T) {
	data, err := fs.ReadFile(assets.CommandsFS, "commands/claude/active/starter-add.md")
	if err != nil {
		t.Fatalf("CommandsFS: failed to read Claude variant: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "active-stack starter add") {
		t.Errorf("Claude variant body must contain 'active-stack starter add'; got:\n%s", body)
	}
	if !strings.Contains(body, "$ARGUMENTS") {
		t.Errorf("Claude variant body must contain '$ARGUMENTS' (confirmed from TBD-1); got:\n%s", body)
	}
}

// TestCommandsFS_OpenCodeVariant_ResolvesAndContainsActiveStackInvocation asserts that:
// - the embedded CommandsFS contains the OpenCode variant at commands/opencode/active-starter-add.md
// - the body contains "active-stack starter add" (thin wrapper requirement)
// - the body contains "$ARGUMENTS" (the confirmed argument token from TBD-1)
// RED: fails until CommandsFS and the OpenCode asset file are added (4.2).
func TestCommandsFS_OpenCodeVariant_ResolvesAndContainsActiveStackInvocation(t *testing.T) {
	data, err := fs.ReadFile(assets.CommandsFS, "commands/opencode/active-starter-add.md")
	if err != nil {
		t.Fatalf("CommandsFS: failed to read OpenCode variant: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "active-stack starter add") {
		t.Errorf("OpenCode variant body must contain 'active-stack starter add'; got:\n%s", body)
	}
	if !strings.Contains(body, "$ARGUMENTS") {
		t.Errorf("OpenCode variant body must contain '$ARGUMENTS' (confirmed from TBD-1); got:\n%s", body)
	}
}

// TestCommandsFS_ClaudeVariant_Frontmatter_RichFields asserts that the Claude
// variant frontmatter carries name, description, category, and tags (per D2 + spec).
// TRIANGULATE (4.3): the namespaced /active:starter-add semantics are expressed via
// the file path (active/starter-add.md) — the path drives the invocation name.
func TestCommandsFS_ClaudeVariant_Frontmatter_RichFields(t *testing.T) {
	data, err := fs.ReadFile(assets.CommandsFS, "commands/claude/active/starter-add.md")
	if err != nil {
		t.Fatalf("CommandsFS: failed to read Claude variant: %v", err)
	}
	body := string(data)
	for _, field := range []string{"name:", "description:", "category:", "tags:"} {
		if !strings.Contains(body, field) {
			t.Errorf("Claude variant frontmatter must contain %q; got:\n%s", field, body)
		}
	}
}

// TestCommandsFS_OpenCodeVariant_Frontmatter_DescriptionOnly asserts that the
// OpenCode variant frontmatter carries only description (flat, hyphenated, per D2 + spec).
// TRIANGULATE (4.3): no name/category/tags — those are Claude-only.
func TestCommandsFS_OpenCodeVariant_Frontmatter_DescriptionOnly(t *testing.T) {
	data, err := fs.ReadFile(assets.CommandsFS, "commands/opencode/active-starter-add.md")
	if err != nil {
		t.Fatalf("CommandsFS: failed to read OpenCode variant: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "description:") {
		t.Errorf("OpenCode variant frontmatter must contain 'description:'; got:\n%s", body)
	}
	for _, field := range []string{"name:", "category:", "tags:"} {
		if strings.Contains(body, field) {
			t.Errorf("OpenCode variant frontmatter must NOT contain %q (flat, description-only); got:\n%s", field, body)
		}
	}
}

// TestCommandsFS_NoStarterLogicInBody asserts that neither command body
// reimplements starter resolution logic — both are thin wrappers.
// TRIANGULATE (4.3): the thin-wrapper requirement from D3.
func TestCommandsFS_NoStarterLogicInBody(t *testing.T) {
	claudeData, err := fs.ReadFile(assets.CommandsFS, "commands/claude/active/starter-add.md")
	if err != nil {
		t.Fatalf("read Claude variant: %v", err)
	}
	opencodeData, err := fs.ReadFile(assets.CommandsFS, "commands/opencode/active-starter-add.md")
	if err != nil {
		t.Fatalf("read OpenCode variant: %v", err)
	}
	// These strings would indicate reimplementation of starter resolution logic.
	forbiddenTerms := []string{"func ", "package ", "import ", "starter.Resolve", "starter.Install"}
	for _, body := range []string{string(claudeData), string(opencodeData)} {
		for _, term := range forbiddenTerms {
			if strings.Contains(body, term) {
				t.Errorf("command body must not reimplement starter logic (found %q); body:\n%s", term, body)
			}
		}
	}
}
