from concurrent.futures import ThreadPoolExecutor
import contextlib
import io
import json
import pathlib
import subprocess
import tempfile
from threading import Event
import unittest
from unittest import mock
from types import SimpleNamespace

import delivery
import team


class DeliveryTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        directory = pathlib.Path(self.temp.name)
        self.root = directory / "project"
        self.root.mkdir()
        self.worktree = directory / "R-11-delivery"
        self.git("init", "-q")
        self.git("config", "user.email", "fixture@example.test")
        self.git("config", "user.name", "Fixture")
        self.git("checkout", "-q", "-b", "main")
        self.git("commit", "--allow-empty", "-qm", "base")

    def git(self, *args):
        return subprocess.check_output(["git", "-C", str(self.root), *args], text=True).strip()

    def configure_team(self, names_versioned=True):
        active = self.root / ".active-stack"
        active.mkdir()
        (self.root / ".gitignore").write_text(
            ".active-stack/me.local.json\n.active-stack/team.local.json\n", encoding="utf-8")
        members = [{"id": "m1", "name": "Ana", "github": "ana-dev", "clickup_id": 123}]
        (active / "team.json").write_text(json.dumps({
            "size": 2, "names_versioned": names_versioned, "tracking": "clickup",
            "members": members if names_versioned else [{"id": "m1"}],
        }), encoding="utf-8")
        if not names_versioned:
            (active / "team.local.json").write_text(json.dumps({"members": members}), encoding="utf-8")
        self.git("add", ".")
        self.git("commit", "-qm", "team")

    def run_team(self, *args):
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            try:
                status = team.main(["--root", str(self.root), *args])
            except SystemExit as exc:
                status = exc.code
        return status, json.loads(output.getvalue())

    def test_individual_project_does_not_run_team_checks(self):
        status, result = self.run_team("identity")
        self.assertEqual((status, result), (0, {"ok": True, "team_mode": False}))
        self.assertFalse((self.root / ".active-stack").exists())

    def test_identity_matches_private_github_and_clickup_but_requires_selection(self):
        self.configure_team(names_versioned=False)
        status, result = self.run_team("identity", "--github", "Ana-Dev", "--clickup-id", "123")
        self.assertEqual(status, 0)
        self.assertEqual(result["suggested"], "m1")
        self.assertEqual(result["members"], [{"id": "m1", "name": "Ana"}])
        self.assertIsNone(result["me"])
        self.assertFalse((self.root / team.ME_LOCAL_PATH).exists())

    def test_unmatched_identity_offers_selection_or_read_only(self):
        self.configure_team()
        status, result = self.run_team("identity", "--github", "other")
        self.assertEqual(status, 0)
        self.assertEqual(result["mode"], "consultation")
        self.assertIsNone(result["suggested"])
        self.assertEqual(result["members"][0]["id"], "m1")

    def test_select_writes_only_local_ignored_identity(self):
        self.configure_team()
        status, result = self.run_team("identity", "--select", "m1")
        self.assertEqual((status, result["me"]), (0, "m1"))
        self.assertEqual(json.loads((self.root / team.ME_LOCAL_PATH).read_text()), {"id": "m1"})
        self.assertEqual(self.git("status", "--porcelain"), "")
        self.assertEqual(self.run_team("identity")[1]["mode"], "identified")

    def test_select_rejects_unignored_file(self):
        self.configure_team()
        (self.root / ".gitignore").unlink()
        status, result = self.run_team("identity", "--select", "m1")
        self.assertEqual(status, 1)
        self.assertIn("ignored", result["errors"][0])
        self.assertFalse((self.root / team.ME_LOCAL_PATH).exists())

    def select(self):
        self.configure_team()
        self.run_team("identity", "--select", "m1")

    def workspace(self, *args):
        return self.run_team("workspace", "--branch", "R-11-delivery", "--id", "R-11",
                             "--base", "main", *args)

    def test_workspace_uses_preview_and_creates_sibling_after_confirmation(self):
        self.select()
        head = self.git("rev-parse", "HEAD")
        status, preview = self.workspace()
        self.assertEqual((status, preview["action"]), (0, "create"))
        self.assertEqual(preview["base_sha"], head)
        self.assertFalse(self.worktree.exists())
        status, result = self.workspace("--apply", "--expected-base", head)
        self.assertEqual((status, result["action"]), (0, "ready"))
        self.assertTrue(self.worktree.is_dir())
        self.assertEqual(self.git("branch", "--show-current"), "main")
        self.assertEqual(self.git("rev-parse", "R-11-delivery"), head)

    def test_existing_branch_retry_and_existing_worktree_report_ownership(self):
        self.select()
        head = self.git("rev-parse", "HEAD")
        self.git("branch", "R-11-delivery")
        self.assertEqual(self.workspace()[1]["action"], "attach")
        self.assertEqual(self.workspace("--apply", "--expected-base", head, "--expected-tip", head)[1]["action"], "ready")
        self.assertEqual(self.workspace("--apply", "--expected-base", head)[1]["action"], "occupied")
        listing = self.git("worktree", "list", "--porcelain", "-z")
        paths = [
            pathlib.Path(field[len("worktree "):]).resolve()
            for field in listing.split("\0")
            if field.startswith("worktree ")
        ]
        self.assertEqual(paths.count(self.worktree.resolve()), 1)

    def test_dirty_checkout_does_not_create_or_change_worktree(self):
        self.select()
        head = self.git("rev-parse", "HEAD")
        (self.root / "untracked").write_text("local work", encoding="utf-8")
        self.assertEqual(self.workspace("--apply", "--expected-base", head)[1]["action"], "dirty")
        self.assertFalse(self.worktree.exists())
        self.assertEqual((self.root / "untracked").read_text(), "local work")
        self.assertEqual(self.git("branch", "--list", "R-11-delivery"), "")

    def test_workspace_checks_base_and_requires_identity(self):
        self.configure_team()
        self.assertEqual(self.workspace()[1]["action"], "consultation")
        self.run_team("identity", "--select", "m1")
        old = self.git("rev-parse", "HEAD")
        self.git("commit", "--allow-empty", "-qm", "updated")
        status, result = self.workspace("--apply", "--expected-base", old)
        self.assertEqual((status, result["action"]), (1, "base-changed"))
        self.assertFalse(self.worktree.exists())

    def test_workspace_does_not_attach_or_replace_an_existing_folder(self):
        self.select()
        self.worktree.mkdir()
        head = self.git("rev-parse", "HEAD")
        self.assertEqual(self.workspace("--apply", "--expected-base", head)[1]["action"], "collision")
        self.assertEqual(self.git("branch", "--list", "R-11-delivery"), "")


    def test_concurrent_slug_choices_share_one_worktree_per_ficha(self):
        self.select()
        head = self.git("rev-parse", "HEAD")
        state = team.validate_project(self.root)
        first_waiting = Event()
        release_first = Event()
        run = subprocess.run

        def delay_first(command, *args, **kwargs):
            if command[3:6] == ["worktree", "add", "-b"] and command[6] == "R-11-delivery":
                first_waiting.set()
                release_first.wait(10)
            return run(command, *args, **kwargs)

        def apply(branch):
            args = SimpleNamespace(id="R-11", branch=branch, base="main",
                                   apply=True, expected_base=head, expected_tip=None)
            return delivery.workspace(self.root.resolve(), args, state)

        with mock.patch.object(delivery.subprocess, "run", side_effect=delay_first):
            with ThreadPoolExecutor(max_workers=1) as executor:
                first = executor.submit(apply, "R-11-delivery")
                try:
                    self.assertTrue(first_waiting.wait(10))
                    self.assertEqual(apply("R-11-other")["action"], "busy")
                finally:
                    release_first.set()
                self.assertEqual(first.result()["action"], "ready")
        self.assertEqual(apply("R-11-other")["action"], "branch-conflict")
        self.assertFalse((self.root.parent / "R-11-other").exists())

    def inspect(self, command, evidence, *arguments):
        path = pathlib.Path(self.temp.name) / "evidence.json"
        path.write_text(json.dumps(evidence), encoding="utf-8")
        before = self.git("status", "--porcelain")
        result = self.run_team(command, "--file", str(path), *arguments)
        self.assertEqual(self.git("status", "--porcelain"), before)
        return result

    def test_overview_keeps_roadmap_and_provider_evidence_during_partial_outage(self):
        self.select()
        date = "2026-11-20T10:30:00Z"
        status, result = self.inspect("overview", {
            "sources": {
                "roadmap": {"source": "CHANGES.md", "checked_at": date, "available": True},
                "github": {"source": "GitHub PRs", "checked_at": date, "available": False,
                           "reason": "gh unavailable"},
                "clickup": {"source": "ClickUp", "checked_at": date, "available": True},
            },
            "fichas": [
                {"id": "R-11", "Responsable": "m1", "Depende de": "R-06, #32@abcdef1"},
                {"id": "R-12", "Depende de": "R-11"},
                {"id": "R-13", "Responsable": "m1", "Dependencias": "ninguna"},
            ],
            "observations": {
                "R-06": {"status": "merged", "assignee": None, "source": "clickup"},
                "R-11": {"status": "open", "assignee": "m1", "source": "clickup"},
                "#32": {"status": "closed", "source": "clickup"},
                "R-13": {"status": "open", "assignee": None, "source": "clickup"},
            },
        })
        self.assertEqual(status, 0)
        self.assertEqual([f["id"] for f in result["mine"]], ["R-11"])
        self.assertEqual(result["mine"][0]["status"],
                         {"value": "open", "source": "ClickUp", "checked_at": date})
        self.assertEqual(result["waiting"][0]["dependency"]["value"], "#32@abcdef1")
        self.assertEqual(result["blocks"][0]["id"], "R-12")
        self.assertEqual(result["sources"]["github"]["reason"], "gh unavailable")
        self.assertIsNone(result["reviews"])

    def test_overview_uses_local_ficha_and_fresh_review_evidence(self):
        self.select()
        date = "2026-11-20T10:30:00Z"
        status, result = self.inspect("overview", {
            "sources": {
                "roadmap": {"source": "CHANGES.md", "checked_at": date, "available": True},
                "github": {"source": "GitHub PRs", "checked_at": date, "available": True},
            },
            "fichas": [{"id": "R-11", "Responsable": "m1", "Depende de": "ninguna"}],
            "reviews": [42],
        })
        self.assertEqual(status, 0)
        self.assertEqual(result["mine"][0]["assignee"]["source"], "CHANGES.md")
        self.assertIsNone(result["mine"][0]["status"])
        self.assertEqual(result["reviews"], [{"value": 42, "source": "GitHub PRs", "checked_at": date}])
        self.assertEqual(result["waiting"], [])

    def test_overview_identifies_open_dependency_owner_and_ficha_sources(self):
        self.select()
        date = "2026-11-20T10:30:00Z"
        status, result = self.inspect("overview", {
            "sources": {
                "roadmap": {"source": "CHANGES.md", "checked_at": date, "available": True},
                "github": {"source": "GitHub", "checked_at": date, "available": True},
            },
            "fichas": [
                {"id": "R-11", "Responsable": "m1", "Depende de": "R-06, #7"},
                {"id": "R-06", "Responsable": "m1"},
            ],
            "observations": {"#7": {"status": "open", "assignee": "m1", "source": "github"}},
        })
        self.assertEqual(status, 0)
        waiting = result["waiting"]
        self.assertEqual(waiting[0]["assignee"],
                         {"value": "m1", "source": "CHANGES.md", "checked_at": date})
        self.assertEqual(waiting[1]["assignee"],
                         {"value": "m1", "source": "GitHub", "checked_at": date})
        self.assertEqual(waiting[0]["reference"],
                         {"value": "R-11", "source": "CHANGES.md", "checked_at": date})

    def reviewer_evidence(self, tracking="local"):
        self.configure_team()
        path = self.root / team.TEAM_PATH
        path.write_text(json.dumps({
            "size": 3,
            "tracking": tracking,
            "members": [
                {"id": "m1", "github": "ana-dev"},
                {"id": "m2", "github": "bea-dev"},
                {"id": "m3", "github": "caro-dev"},
            ],
        }), encoding="utf-8")
        date = "2026-11-25T10:30:00Z"
        return {
            "sources": {
                "roadmap": {"source": "CHANGES.md", "available": True, "checked_at": date},
                "github": {"source": "GitHub", "available": True, "checked_at": date},
                "clickup": {"source": "ClickUp", "available": True, "checked_at": date},
            },
            "fichas": [
                {"id": "R-11", "Responsable": "m1"},
                {"id": "R-21", "Responsable": "m3"},
                {"id": "R-19", "Responsable": "m2", "Depende de": "R-11"},
                {"id": "R-12", "Responsable": "m3"},
            ],
            "author_github": "Ana-Dev",
            "observations": {"R-21": {"status": "omitted", "source": "roadmap"}},
        }

    def test_reviewer_uses_roadmap_order_and_skips_confirmed_omissions(self):
        evidence = self.reviewer_evidence()
        status, result = self.inspect("reviewer", evidence, "--id", "R-11")
        self.assertEqual((status, result["action"], result["next_ficha"], result["github"]),
                         (0, "next", "R-19", "bea-dev"))

    def test_reviewer_waits_for_approved_independent_fallback(self):
        evidence = self.reviewer_evidence()
        evidence["unavailable_members"] = ["m2"]
        for fallback in (None, {"member": "m3", "approved": False},
                         {"member": "m1", "approved": True}):
            with self.subTest(fallback=fallback):
                if fallback is not None:
                    evidence["fallback"] = fallback
                result = self.inspect("reviewer", evidence, "--id", "R-11")
                self.assertEqual((result[0], result[1]["action"]), (1, "fallback-required"))
        evidence["fallback"] = {"member": "m3", "approved": True}
        status, result = self.inspect("reviewer", evidence, "--id", "R-11")
        self.assertEqual((status, result["action"], result["github"]), (0, "fallback", "caro-dev"))

    def test_reviewer_requires_a_fallback_for_the_last_ficha(self):
        evidence = self.reviewer_evidence()
        evidence["author_github"] = "caro-dev"
        evidence["fallback"] = {"member": "m2", "approved": True}
        status, result = self.inspect("reviewer", evidence, "--id", "R-12")
        self.assertEqual((status, result["action"], result["next_ficha"], result["member"]),
                         (0, "fallback", None, "m2"))

    def test_reviewer_uses_only_confirmed_clickup_assignment(self):
        evidence = self.reviewer_evidence(tracking="clickup")
        evidence["observations"]["R-19"] = {"assignee": "m3", "source": "clickup"}
        status, result = self.inspect("reviewer", evidence, "--id", "R-11")
        self.assertEqual((status, result["github"]), (0, "caro-dev"))
        evidence["sources"]["clickup"]["available"] = False
        status, result = self.inspect("reviewer", evidence, "--id", "R-11")
        self.assertEqual((status, result["action"], result["candidate"]),
                         (1, "fallback-required", None))

    def test_notices_report_all_five_and_do_not_write(self):
        self.select()
        status, result = self.inspect("notices", {
            "assignment": {"assignee": "m2", "branch": "R-11-team", "pr": 40},
            "dependency": {"expected": "1234567", "actual": "123456789"},
            "mode": "odd",
            "zone": ["internal/catalog"],
            "open_prs": [
                {"number": 42, "paths": ["internal/catalog/catalog.go"]},
                {"number": 43, "paths": ["openspec/changes/catalog/specs/catalog/spec.md"],
                 "sdd_zone": ["internal/catalog"]},
            ],
            "requirements_met": False,
        })
        self.assertEqual(status, 0)
        self.assertEqual([notice["id"] for notice in result["notices"]], [1, 3, 4, 5])
        status, result = self.inspect("notices", {
            "assignment": {"branch": "R-11-team"},
            "dependency": {"expected": "1234567", "actual": "abcdef123"},
            "mode": "odd",
            "zone": ["internal/catalog"],
            "open_prs": [
                {"number": 42, "paths": ["internal/catalog/catalog.go"]},
                {"number": 43, "paths": ["openspec/changes/catalog/specs/catalog/spec.md"],
                 "sdd_zone": ["internal/catalog"]},
            ],
            "requirements_met": False,
        })
        self.assertEqual(status, 0)
        self.assertEqual([notice["id"] for notice in result["notices"]], [1, 2, 3, 4, 5])

    def test_notices_ignore_current_pr_unrelated_scope_and_sdd_for_notice_four(self):
        self.select()
        status, result = self.inspect("notices", {
            "assignment": {"assignee": "m1"},
            "dependency": {"expected": "1234567", "actual": "123456789"},
            "mode": "sdd",
            "zone": ["internal/catalog"],
            "current_pr": 42,
            "open_prs": [
                {"number": 42, "paths": ["internal/catalog/catalog.go"]},
                {"number": 43, "paths": ["openspec/changes/catalog/tasks.md", "internal/catalog-old/a"],
                 "sdd_zone": ["internal/catalog"]},
            ],
            "requirements_met": True,
        })
        self.assertEqual(status, 0)
        self.assertEqual(result["notices"], [])
        self.assertEqual(result["unconfirmed"], [])

    def test_notices_keep_failed_checks_unconfirmed(self):
        self.select()
        status, result = self.inspect("notices", {
            "assignment": {}, "dependency": {"expected": "1234567"},
            "mode": "odd", "zone": ["internal/catalog"], "open_prs": None,
        })
        self.assertEqual(status, 0)
        self.assertEqual(result["notices"], [])
        self.assertEqual(result["unconfirmed"], [2, 3, 4, 5])

    def test_cross_client_worktree_uses_shared_member_ids_and_local_selection(self):
        self.configure_team(names_versioned=False)
        self.assertEqual(self.run_team("identity", "--github", "ana-dev")[1]["suggested"], "m1")
        self.run_team("identity", "--select", "m1")
        head = self.git("rev-parse", "HEAD")
        self.assertEqual(self.workspace("--apply", "--expected-base", head)[1]["action"], "ready")
        self.assertFalse((self.worktree / team.TEAM_LOCAL_PATH).exists())
        self.assertFalse((self.worktree / team.ME_LOCAL_PATH).exists())
        self.root = self.worktree
        self.assertEqual(self.run_team("identity")[1]["members"], [{"id": "m1", "name": None}])
        self.assertEqual(self.run_team("identity", "--select", "m1")[1]["me"], "m1")
        self.assertEqual(self.git("status", "--porcelain"), "")


if __name__ == "__main__":
    unittest.main()
