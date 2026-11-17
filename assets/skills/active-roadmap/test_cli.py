import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


class RoadmapCliTest(unittest.TestCase):
    def test_selective_publish_retry_and_reconcile(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            config = root / ".active-stack"
            config.mkdir()
            (config / "roadmap.json").write_text(json.dumps({
                "mode": "clickup", "project_id": "pilot", "list_id": "987", "assignees": {}
            }), encoding="utf-8")
            helper = Path(__file__).with_name("roadmap.py")

            def run(*args):
                result = subprocess.run(
                    [sys.executable, "-B", str(helper), "--root", str(root), *args],
                    check=True, capture_output=True, encoding="utf-8",
                )
                return json.loads(result.stdout)

            def payload(name, value):
                path = root / name
                path.write_text(json.dumps(value), encoding="utf-8")
                return str(path)

            diagram = "```mermaid\nflowchart LR\n    F1 --> F2\n```"
            tasks = payload("tasks.json", [
                {"id": "F1", "title": "First", "responsible": "Ana"},
                {"id": "F2", "title": "Second", "dependencies": ["F1"], "description": "Dependencies\n\n" + diagram},
                {"id": "F3", "title": "Third"},
            ])
            plan = run("preview", tasks)
            self.assertIn(diagram, plan["rows"][1]["description"])
            self.assertFalse((config / "roadmap-state.json").exists())
            batch = run("begin", tasks, plan["approval"], "F1", "F2")
            self.assertEqual([r["id"] for r in batch["rows"]], ["F1", "F2"])
            self.assertIn(diagram, batch["rows"][1]["description"])
            self.assertIsNone(batch["rows"][0]["assignee"])
            self.assertIn("Ana", batch["rows"][0]["description"])
            for task_id, result in (
                ("F1", {"outcome": "created", "remote_id": "one", "url": "https://app.clickup.com/t/one"}),
                ("F2", {"outcome": "failed", "reason": "simulated definitive rejection"}),
            ):
                run("result", task_id, payload("result.json", result))
            retry = run("preview", tasks)
            self.assertEqual([r["action"] for r in retry["rows"]], ["unchanged", "create", "create"])
            second = run("begin", tasks, retry["approval"], "F2")
            self.assertEqual([r["id"] for r in second["rows"]], ["F2"])
            run("result", "F2", payload("result.json", {
                "outcome": "created", "remote_id": "two", "url": "https://app.clickup.com/t/two"
            }))
            state = run("reconcile", "F1", payload("snapshot.json", {"status": "done", "assignees": ["ana"]}))
            self.assertEqual(state["observed"]["status"], "done")
            self.assertFalse(state["assignment_pending"])
            self.assertEqual(set(run("read")["tasks"]), {"F1", "F2"})
