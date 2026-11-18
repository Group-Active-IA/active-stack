import unittest
from unittest import mock

from roadmap import client_command, diagnose_clickup, parse_claude_mcp_list


CLAUDE_LIST = """Checking MCP server health…

claude.ai ClickUp: https://mcp.clickup.com/mcp - ✔ Connected
engram: engram mcp - ✔ Connected
excalidraw: https://mcp.excalidraw.com (HTTP) - ✔ Connected
"""


def registration(name="clickup", *, enabled=True, auth_status="oauth", url="https://mcp.clickup.com/mcp"):
    return {
        "name": name,
        "enabled": enabled,
        "auth_status": auth_status,
        "transport": {"type": "streamable_http", "url": url},
    }


class SetupTests(unittest.TestCase):
    def test_missing_registration_and_name_conflict(self):
        self.assertEqual(diagnose_clickup([])["status"], "missing")
        self.assertEqual(diagnose_clickup([registration(url="https://example.com/mcp")])["status"], "conflict")

    def test_reuses_official_registration_under_custom_name(self):
        result = diagnose_clickup([registration("workspace", auth_status="not_logged_in")])
        self.assertEqual(result, {"status": "login_required", "name": "workspace"})

    def test_disabled_registration_is_not_replaced(self):
        self.assertEqual(diagnose_clickup([registration(enabled=False)])["status"], "disabled")

    def test_registration_needs_a_read_probe_to_prove_access(self):
        registrations = [registration()]
        self.assertEqual(diagnose_clickup(registrations)["status"], "check_list")
        self.assertEqual(diagnose_clickup(registrations, "forbidden")["status"], "list_unavailable")
        self.assertEqual(diagnose_clickup(registrations, "ok")["status"], "ready")

    def test_provider_auth_error_requires_login(self):
        self.assertEqual(diagnose_clickup([registration()], "unauthorized")["status"], "login_required")

    def test_transient_read_error_does_not_infer_a_permissions_issue(self):
        self.assertEqual(diagnose_clickup([registration()], "error")["status"], "verification_pending")


class ClientCommandTests(unittest.TestCase):
    def test_resolves_windows_npm_shims_by_full_path(self):
        with mock.patch("roadmap.shutil.which", return_value=r"C:\npm\claude.CMD"):
            self.assertEqual(client_command("claude", "mcp", "list"), [r"C:\npm\claude.CMD", "mcp", "list"])

    def test_keeps_the_name_when_the_client_is_not_on_path(self):
        with mock.patch("roadmap.shutil.which", return_value=None):
            self.assertEqual(client_command("codex", "mcp"), ["codex", "mcp"])


class ClaudeSetupTests(unittest.TestCase):
    def test_claude_connector_is_recognized_and_still_needs_a_read_probe(self):
        registrations = parse_claude_mcp_list(CLAUDE_LIST)
        self.assertEqual([item["name"] for item in registrations], ["claude.ai ClickUp", "engram", "excalidraw"])
        self.assertEqual(registrations[1]["transport"], {"type": "stdio", "command": "engram mcp"})
        self.assertEqual(diagnose_clickup(registrations), {"status": "check_list", "name": "claude.ai ClickUp"})
        self.assertEqual(diagnose_clickup(registrations, "ok")["status"], "ready")

    def test_claude_missing_and_pending_authentication(self):
        self.assertEqual(diagnose_clickup(parse_claude_mcp_list("engram: engram mcp - ✔ Connected"))["status"], "missing")
        pending = parse_claude_mcp_list("clickup: https://mcp.clickup.com/mcp (HTTP) - ! Needs authentication")
        self.assertEqual(diagnose_clickup(pending), {"status": "login_required", "name": "clickup"})

    def test_claude_stdio_server_named_clickup_is_a_conflict(self):
        stdio = parse_claude_mcp_list("clickup: npx -y mcp-remote https://mcp.clickup.com/mcp - ✔ Connected")
        self.assertEqual(diagnose_clickup(stdio)["status"], "conflict")

    def test_claude_sse_endpoint_is_not_the_official_transport(self):
        sse = parse_claude_mcp_list("clickup: https://mcp.clickup.com/mcp (SSE) - ✔ Connected")
        self.assertEqual(diagnose_clickup(sse)["status"], "conflict")

    def test_claude_prefers_the_connected_registration(self):
        both = parse_claude_mcp_list(
            "clickup: https://mcp.clickup.com/mcp (HTTP) - ✗ Failed to connect\n"
            "claude.ai ClickUp: https://mcp.clickup.com/mcp - ✔ Connected"
        )
        self.assertEqual(diagnose_clickup(both), {"status": "check_list", "name": "claude.ai ClickUp"})

    def test_claude_ambiguous_registrations_can_be_selected_by_name(self):
        two = parse_claude_mcp_list(
            "plugin:clickup:clickup: https://mcp.clickup.com/mcp (HTTP) - ✔ Connected\n"
            "claude.ai ClickUp: https://mcp.clickup.com/mcp - ✔ Connected"
        )
        self.assertEqual(diagnose_clickup(two)["status"], "ambiguous")
        self.assertEqual(diagnose_clickup(two, name="claude.ai ClickUp"), {"status": "check_list", "name": "claude.ai ClickUp"})
        self.assertEqual(diagnose_clickup(two, name="otro")["status"], "missing")

    def test_claude_failed_connection_keeps_configuration_pending(self):
        failed = parse_claude_mcp_list("clickup: https://mcp.clickup.com/mcp (HTTP) - ✗ Failed to connect")
        self.assertEqual(diagnose_clickup(failed)["status"], "verification_pending")
        self.assertEqual(diagnose_clickup(failed, "ok")["status"], "ready")


if __name__ == "__main__":
    unittest.main()
