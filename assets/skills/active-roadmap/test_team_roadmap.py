import json
from pathlib import Path
import subprocess
import tempfile
import unittest

from roadmap import Roadmap, RoadmapError


class TeamRoadmapTests(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.root = Path(directory.name)
        subprocess.run(["git", "init", "-q", str(self.root)], check=True)
        self.active = self.root / ".active-stack"
        self.active.mkdir()
        self.settings = {
            "mode": "clickup", "list_id": "42", "project_id": "demo",
            "assignees": {"m1": "wrong", "m2": "wrong"},
        }
        self.write("roadmap.json", self.settings)
        self.team = {
            "size": 2, "tracking": "clickup", "names_versioned": True,
            "members": [{"id": "m1", "name": "Ana", "clickup_id": 111}, {"id": "m2"}],
        }
        self.write("team.json", self.team)
        self.write("me.local.json", {"id": "m1"})
        self.roadmap = Roadmap(self.root)
        self.tasks = [
            {"id": "F1", "title": "Primera", "responsible": "m1"},
            {"id": "F2", "title": "Segunda", "responsible": "m2"},
        ]

    def write(self, name, data):
        (self.active / name).write_text(json.dumps(data), encoding="utf-8")

    def preview(self, ids, **options):
        return self.roadmap.preview(self.tasks, selected=ids, **options)

    def linked(self, task_id="F1"):
        plan = self.preview([task_id])
        self.roadmap.begin(self.tasks, plan["approval"], [task_id])
        self.roadmap.result(task_id, "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        task = next(task for task in self.tasks if task["id"] == task_id)
        task.update(current_remote_id="a1", current_text=plan["rows"][0]["description"],
                    current_title=plan["rows"][0]["title"])

    def test_publish_uses_team_ids_and_warns_unlinked_member(self):
        rows = self.preview(["F1", "F2"])["rows"]
        self.assertEqual(rows[0]["assignee"], 111)
        self.assertIsNone(rows[1]["assignee"])
        self.assertTrue(rows[1]["assignment_pending"])

    def test_publish_resolves_private_members_by_stable_id(self):
        self.team.update(names_versioned=False, members=[{"id": "m1"}, {"id": "m2"}])
        self.write("team.json", self.team)
        self.write("team.local.json", {"members": [{"id": "m1", "name": "Ana", "clickup_id": 555}]})
        row = self.preview(["F1"])["rows"][0]
        self.assertEqual(row["assignee"], 555)
        self.assertNotIn("Ana", row["description"])

    def test_selection_changes_require_fresh_preview_and_leave_other_row_untouched(self):
        plan = self.preview(["F1"])
        self.assertEqual([row["id"] for row in plan["rows"]], ["F1"])
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, plan["approval"], ["F2"])
        self.assertEqual(self.roadmap.read()["tasks"], {})
        result = self.roadmap.begin(self.tasks, plan["approval"], ["F1"])
        self.assertEqual([row["id"] for row in result["rows"]], ["F1"])
        self.assertEqual(set(self.roadmap.read()["tasks"]), {"F1"})

    def test_take_previews_current_and_new_assignees(self):
        self.linked()
        self.tasks[0].update(operation="take", current_assignees=[222])
        plan = self.preview(["F1"])
        row = plan["rows"][0]
        self.assertEqual(row["current_assignees"], [222])
        self.assertEqual(row["changes"], {"assignees": [111]})
        self.tasks[0]["current_assignees"] = [333]
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, plan["approval"], ["F1"])
        self.assertNotIn("pending", self.roadmap.read()["tasks"]["F1"])

    def test_link_preview_checks_branch_identity_and_preserves_external_markdown(self):
        old = self.preview(["F1"])["rows"][0]["description"]
        self.linked()
        self.tasks[0].update(operation="link_pr", current_description="Extra\n" + old,
                             branch="F2-wrong", pr_url="https://github.com/demo/repo/pull/2")
        with self.assertRaises(RoadmapError):
            self.preview(["F1"])
        self.tasks[0]["branch"] = "F1-correct"
        row = self.preview(["F1"])["rows"][0]
        self.assertTrue(row["changes"]["description"].startswith("Extra\n"))
        self.assertIn(self.tasks[0]["pr_url"], row["changes"]["description"])

    def test_close_includes_pr_sha_handoff_and_status(self):
        old = self.preview(["F1"])["rows"][0]["description"]
        self.linked()
        self.tasks[0].update(
            operation="close", current_description=old, branch="F1-delivery",
            pr_url="https://github.com/demo/repo/pull/2", merge_sha="a" * 40,
            handoff="knowledge-base/entrega/F1.md", current_status="review", closing_status="done",
        )
        row = self.preview(["F1"])["rows"][0]
        self.assertEqual(row["changes"]["status"], "done")
        for field in ("pr_url", "merge_sha", "handoff"):
            self.assertIn(self.tasks[0][field], row["changes"]["description"])

    def test_clear_denial_or_local_mode_cannot_publish(self):
        plan = self.preview(["F1"], access="forbidden")
        self.assertEqual(plan["mode"], "local")
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, plan["approval"], ["F1"])
        self.team["tracking"] = "local"
        for member in self.team["members"]:
            member.pop("clickup_id", None)
        self.write("team.json", self.team)
        self.assertEqual(self.preview(["F1"])["rows"][0]["action"], "local")
        self.assertEqual(self.roadmap.read()["tasks"], {})

    def test_shared_state_keeps_links_but_observations_are_local(self):
        self.linked()
        self.roadmap.reconcile("F1", status="done", assignees=[111])
        shared = json.loads(self.roadmap.state_path.read_text(encoding="utf-8"))
        self.assertEqual(shared["tasks"]["F1"]["remote_id"], "a1")
        self.assertNotIn("observed", shared["tasks"]["F1"])
        self.assertNotIn("result", shared["tasks"]["F1"])
        self.assertEqual(self.roadmap.read()["tasks"]["F1"]["observed"]["assignees"], [111])

    def test_assignment_reconciliation_changes_only_assignee_under_row_approval(self):
        self.linked()
        self.tasks[0].update(operation="assign", current_assignees=[222])
        plan = self.preview(["F1"])
        self.assertEqual(plan["rows"][0]["changes"], {"assignees": [111]})
        result = self.roadmap.begin(self.tasks, plan["approval"], ["F1"])
        self.assertEqual(result["rows"][0]["changes"], {"assignees": [111]})
        self.roadmap.result("F1", "updated", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.assertEqual(self.roadmap.read()["tasks"]["F1"]["remote_id"], "a1")
        self.tasks[0]["current_assignees"] = [111]
        self.assertEqual(self.preview(["F1"])["rows"][0]["action"], "unchanged")

    def test_existing_link_without_member_mapping_warns_in_fresh_worktree(self):
        self.linked("F2")
        self.roadmap.journal_path().unlink()
        self.tasks[1].update(operation="assign", current_assignees=[])
        row = self.preview(["F2"])["rows"][0]
        self.assertTrue(row["assignment_pending"])
        self.assertEqual(row["changes"], {})
        self.assertEqual(row["action"], "unchanged")

    def test_delivery_proof_survives_later_publication(self):
        old = self.preview(["F1"])["rows"][0]["description"]
        self.linked()
        url = "https://github.com/demo/repo/pull/2"
        self.tasks[0].update(operation="link_pr", current_description=old,
                             branch="F1-delivery", pr_url=url)
        plan = self.preview(["F1"])
        self.roadmap.begin(self.tasks, plan["approval"], ["F1"])
        self.roadmap.result("F1", "updated", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.tasks[0] = {"id": "F1", "title": "Nuevo", "responsible": "m1",
                         "current_description": plan["rows"][0]["changes"]["description"]}
        self.assertIn(url, self.preview(["F1"])["rows"][0]["description"])

    def test_missing_private_map_does_not_fall_back_to_legacy_config(self):
        self.team.update(names_versioned=False, members=[{"id": "m1"}, {"id": "m2"}])
        self.write("team.json", self.team)
        self.assertTrue(self.preview(["F1"])["rows"][0]["assignment_pending"])
        self.linked()
        self.tasks[0].update(operation="take", current_assignees=[])
        with self.assertRaisesRegex(RoadmapError, "ID confirmado"):
            self.preview(["F1"])

    def test_selected_publish_marks_existing_unlinked_member_pending_in_fresh_worktree(self):
        self.linked("F2")
        self.roadmap.journal_path().unlink()
        row = self.preview(["F2"])["rows"][0]
        self.assertTrue(row["assignment_pending"])


    def test_team_update_requires_current_remote_identity_marker(self):
        self.linked()
        self.tasks[0].update(operation="take", current_assignees=[])
        for field, invalid in (("current_remote_id", "a2"), ("current_text", "Other task")):
            with self.subTest(field=field):
                original = self.tasks[0][field]
                self.tasks[0][field] = invalid
                with self.assertRaisesRegex(RoadmapError, "identidad"):
                    self.preview(["F1"])
                self.tasks[0][field] = original

    def test_link_keeps_title_change_pending_for_later_publication(self):
        old = self.preview(["F1"])["rows"][0]["description"]
        self.linked()
        self.tasks[0].update(operation="link_pr", current_description=old,
                             branch="F1-delivery", pr_url="https://github.com/demo/repo/pull/2",
                             title="Local edited title", current_title="Primera")
        plan = self.preview(["F1"])
        self.roadmap.begin(self.tasks, plan["approval"], ["F1"])
        self.roadmap.result("F1", "updated", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.tasks[0].update(operation="publish", current_description=plan["rows"][0]["changes"]["description"])
        self.assertEqual(self.preview(["F1"])["rows"][0]["action"], "update")
