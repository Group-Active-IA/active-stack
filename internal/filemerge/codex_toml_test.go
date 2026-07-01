package filemerge

import (
	"strings"
	"testing"
)

func TestUpsertTOMLTopLevelPreservesNestedKey(t *testing.T) {
	input := "approval_policy = \"never\"\n\n[profile.ci]\napproval_policy = \"never\"\n"
	got, err := UpsertTOMLTopLevel(input, map[string]string{
		"approval_policy": `"on-request"`,
		"sandbox_mode":    `"workspace-write"`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `approval_policy = "on-request"`) {
		t.Fatalf("missing updated top-level key:\n%s", got)
	}
	if !strings.Contains(got, "[profile.ci]\napproval_policy = \"never\"") {
		t.Fatalf("nested key was changed:\n%s", got)
	}
}

func TestUpsertTOMLSectionReplacesNestedTables(t *testing.T) {
	input := "model = \"gpt\"\n\n[mcp_servers.engram]\nurl = \"old\"\n\n[mcp_servers.engram.env]\nTOKEN = \"old\"\n\n[mcp_servers.other]\ncommand = \"other\"\n"
	body := "command = \"engram\"\nargs = [\"mcp\", \"--tools=agent\"]"
	got, err := UpsertTOMLSection(input, "mcp_servers.engram", body)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "url = \"old\"") || strings.Contains(got, "TOKEN = \"old\"") {
		t.Fatalf("stale managed keys remain:\n%s", got)
	}
	if !strings.Contains(got, "[mcp_servers.other]") || !strings.Contains(got, `model = "gpt"`) {
		t.Fatalf("unrelated config was lost:\n%s", got)
	}
	second, err := UpsertTOMLSection(got, "mcp_servers.engram", body)
	if err != nil || second != got {
		t.Fatalf("upsert not idempotent: err=%v\n%s", err, second)
	}
}

func TestTOMLMutationsRejectInvalidInput(t *testing.T) {
	if _, err := UpsertTOMLSection("[broken", "mcp_servers.x", `url = "x"`); err == nil {
		t.Fatal("expected invalid TOML error")
	}
	if _, err := UpsertTOMLTopLevel("[broken", map[string]string{"sandbox_mode": `"read-only"`}); err == nil {
		t.Fatal("expected invalid TOML error")
	}
}

func TestUpsertTOMLTableValuesPreservesOtherValues(t *testing.T) {
	input := "[sandbox_workspace_write]\nwritable_roots = [\"/work\"]\nnetwork_access = true\n"
	got, err := UpsertTOMLTableValues(input, "sandbox_workspace_write", map[string]string{
		"network_access": "false",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `writable_roots = ["/work"]`) || !strings.Contains(got, "network_access = false") {
		t.Fatalf("table values not merged safely:\n%s", got)
	}
}
