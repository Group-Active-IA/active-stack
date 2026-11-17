import unittest

from roadmap import diagnose_clickup


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


if __name__ == "__main__":
    unittest.main()
