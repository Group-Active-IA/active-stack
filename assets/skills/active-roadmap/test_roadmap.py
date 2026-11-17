import json
from pathlib import Path
import tempfile
import unittest

from roadmap import Roadmap, RoadmapError


class RoadmapTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        self.config = self.root / ".active-stack" / "roadmap.json"
        self.config.parent.mkdir()
        self.config.write_text(json.dumps({
            "mode": "clickup", "list_id": "42", "project_id": "demo",
            "assignees": {"Ana": "9"}
        }))
        self.roadmap = Roadmap(self.root)
        self.tasks = [
            {"id": "F1", "title": "Primera", "responsible": "Ana"},
            {"id": "F2", "title": "Segunda", "responsible": "Beto", "dependencies": ["F1"]},
            {"id": "F3", "title": "Tercera"},
        ]

    def begin(self, task_id):
        plan = self.roadmap.preview(self.tasks)
        return self.roadmap.begin(self.tasks, plan["approval"], [task_id])

    def test_preview_has_no_writes_and_exposes_pending_assignment(self):
        plan = self.roadmap.preview(self.tasks)
        self.assertEqual(plan["list_id"], "42")
        first, second, third = plan["rows"]
        self.assertEqual(first["assignee"], "9")
        self.assertIsNone(second["assignee"])
        self.assertTrue(second["assignment_pending"])
        self.assertIn("Beto", second["description"])
        self.assertIn("F1", second["description"])
        self.assertFalse(third["assignment_pending"])
        self.assertFalse(self.roadmap.state_path.exists())

    def test_selective_approval_rejects_stale_plan_and_duplicate_ids(self):
        preview = self.roadmap.preview(self.tasks)
        self.tasks[0]["title"] = "Editada"
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, preview["approval"], ["F1"])
        approval = self.roadmap.preview(self.tasks)["approval"]
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, approval, ["F1", "F1"])
        self.roadmap.begin(self.tasks, approval, ["F2"])
        self.assertEqual(set(self.roadmap.read()["tasks"]), {"F2"})

    def test_uncertain_creation_requires_recovery_and_keeps_remote_identity(self):
        plan = self.begin("F2")
        state = self.roadmap.preview(self.tasks)
        self.assertEqual(state["rows"][1]["action"], "recover")
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, state["approval"], ["F2"])
        self.roadmap.result("F2", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.assertEqual(self.roadmap.preview(self.tasks)["rows"][1]["action"], "unchanged")
        self.tasks[1]["title"] = "Editada"
        self.tasks[1]["current_description"] = plan["rows"][0]["description"]
        row = self.roadmap.preview(self.tasks)["rows"][1]
        self.assertEqual((row["action"], row["remote_id"]), ("update", "a1"))

    def test_failure_and_reconciliation_do_not_lose_last_snapshot(self):
        self.begin("F1")
        self.roadmap.result("F1", "failed", reason="Rejected")
        self.assertEqual(self.roadmap.preview(self.tasks)["rows"][0]["action"], "create")
        self.begin("F1")
        self.roadmap.result("F1", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.roadmap.reconcile("F1", status="done", assignees=["9"])
        self.roadmap.reconcile("F1", reason="Timeout")
        entry = self.roadmap.read()["tasks"]["F1"]
        self.assertEqual(entry["observed"]["status"], "done")
        self.assertEqual(entry["observed"]["assignees"], ["9"])
        self.assertTrue(entry["verification_pending"])

    def test_update_approval_includes_external_description_and_preserves_assignment(self):
        self.begin("F2")
        description = self.roadmap.read()["tasks"]["F2"]["pending"]["description"]
        self.roadmap.result("F2", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        task = self.tasks[1]
        task["title"] = "Editada"
        task["current_description"] = f"Nota manual\n{description}"
        plan = self.roadmap.preview(self.tasks)
        self.assertIn("Nota manual", plan["rows"][1]["description"])
        self.assertIsNone(plan["rows"][1]["assignee"])
        task["current_description"] += "\nOtra nota"
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, plan["approval"], ["F2"])
        self.begin("F2")
        with self.assertRaises(RoadmapError):
            self.roadmap.result("F2", "updated", remote_id="duplicate", url="https://app.clickup.com/t/duplicate")

    def test_update_without_live_managed_block_cannot_be_approved(self):
        self.begin("F1")
        self.roadmap.result("F1", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.tasks[0]["title"] = "Editada"
        with self.assertRaises(RoadmapError):
            self.begin("F1")
        self.tasks[0]["current_description"] = "Texto ajeno"
        with self.assertRaises(RoadmapError):
            self.begin("F1")

    def test_update_merges_clickup_escaped_markers_into_original_markdown(self):
        row = self.begin("F2")["rows"][0]
        self.roadmap.result("F2", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        marker = row["marker"]
        self.tasks[1].update(
            description="Nuevo alcance",
            current_description=(
                f"## Nota **manual**\n\\[{marker}\\]\n"
                f"Ficha: [CHANGES.md](https://changes.md)\n\\[/{marker}\\]\n"
                "\n> Conservar [enlace](https://example.com)."
            ),
        )
        update = self.begin("F2")["rows"][0]
        self.assertIn("Nuevo alcance", update["description"])
        self.assertIn(f"\\[{marker}\\]", update["description"])
        self.assertIn(f"\\[/{marker}\\]", update["description"])
        self.assertTrue(update["description"].startswith("## Nota **manual**\n"))
        self.assertTrue(update["description"].endswith("\n\n> Conservar [enlace](https://example.com)."))
        self.assertNotIn("https://changes.md", update["description"])

    def test_update_rejects_mixed_or_duplicate_markers(self):
        row = self.begin("F1")["rows"][0]
        self.roadmap.result("F1", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        self.tasks[0]["title"] = "Editada"
        marker = row["marker"]
        raw = f"[{marker}]\nNota\n[/{marker}]"
        escaped = f"\\[{marker}\\]\nNota\n\\[/{marker}\\]"
        for invalid in (f"[{marker}]\nNota\n\\[/{marker}\\]", raw + "\n" + escaped, f"\\[{marker}]\n" + raw):
            with self.subTest(description=invalid):
                self.tasks[0]["current_description"] = invalid
                self.assertEqual(self.roadmap.preview(self.tasks)["rows"][0]["action"], "inspect")

    def test_local_roadmap_cannot_publish(self):
        self.config.write_text('{"mode":"local"}')
        plan = self.roadmap.preview(self.tasks)
        self.assertEqual(plan["rows"][0]["action"], "local")
        with self.assertRaises(RoadmapError):
            self.roadmap.begin(self.tasks, plan["approval"], ["F1"])


if __name__ == "__main__":
    unittest.main()
