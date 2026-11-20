import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


HOOK = Path(__file__).with_name("hook.py")


class SaveFixture(unittest.TestCase):
    def git(self, *args):
        return subprocess.check_output(["git", "-C", self.repo, *args], text=True).strip()

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.repo = Path(self.tmp.name) / "repo"
        self.repo.mkdir()
        self.git("init", "-q", "-b", "main")
        self.git("config", "user.name", "Persona Piloto")
        self.git("config", "user.email", "fixture@example.invalid")
        marker = self.repo / ".engram" / "active-stack.json"
        marker.parent.mkdir()
        marker.write_text('{"memory_policy":"cloud-branch"}')
        self.git("add", ".engram")
        self.git("-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
        self.git("remote", "add", "origin", "../offline.git")

    def event(self, cwd=None, tool="mcp__engram__mem_save"):
        return {"hook_event_name": "PreToolUse", "cwd": str(cwd or self.repo), "tool_name": tool,
                "tool_input": {"project": "fixture", "session_id": "original", "scope": "project",
                               "capture_prompt": False, "topic_key": "K", "title": "T", "content": "C"}}

    def run_hook(self, event):
        p = subprocess.run([sys.executable, str(HOOK)], input=json.dumps(event), text=True,
                           capture_output=True, check=True)
        return json.loads(p.stdout)["hookSpecificOutput"] if p.stdout else None

    def test_linked_worktree(self):
        linked = self.repo.parent / "linked"
        self.git("worktree", "add", "-qb", "feat/R-01", str(linked))
        event = self.event(linked)
        result = self.run_hook(event)
        self.assertEqual(result["permissionDecision"], "allow")
        self.assertEqual(result["updatedInput"], {**event["tool_input"], "topic_key": "branch/feat-r-01/K",
                         "title": "[feat-r-01] T", "content": "Branch: feat/R-01 · Author: Persona Piloto\nC"})

    def test_switch_checkout_main_and_alias(self):
        event = self.event(tool="mcp__plugin_engram_engram__mem_save")
        event["agent_id"] = "fixture-child"
        del event["tool_input"]["topic_key"]
        for command, args, prefix in [("switch", ["-qc", "otra"], "[otra] "),
                                      ("checkout", ["-qb", "tercera"], "[tercera] "),
                                      ("switch", ["-q", "main"], "")]:
            self.git(command, *args)
            data = self.run_hook(event)["updatedInput"]
            self.assertEqual(data["title"], prefix + "T")
            self.assertNotIn("topic_key", data)
        self.assertEqual(data["content"], "Author: Persona Piloto\nC")

    def test_outside_matcher_and_unmarked(self):
        for tool in ["mcp__otro__mem_save", "mcp__engram__mem_search"]:
            self.assertIsNone(self.run_hook(self.event(tool=tool)))
        (self.repo / ".engram" / "active-stack.json").unlink()
        self.assertIsNone(self.run_hook(self.event()))

    def test_long_branches_and_detached(self):
        slugs = []
        for suffix in ("a", "b"):
            self.git("checkout", "-qb", "X"*41 + suffix)
            title = self.run_hook(self.event())["updatedInput"]["title"]
            self.assertRegex(title, r"^\[[a-z0-9-]{1,33}-[0-9a-f]{6}\] T$")
            slugs.append(title)
        self.assertNotEqual(*slugs)
        self.git("checkout", "--detach", "-q")
        sha = self.git("rev-parse", "HEAD")[:7]
        self.assertEqual(self.run_hook(self.event())["updatedInput"]["title"], f"[detached-{sha}] T")

    def test_malformed_and_no_remote(self):
        p = subprocess.run([sys.executable, str(HOOK)], input="[]", capture_output=True, text=True)
        self.assertEqual((p.returncode, p.stdout), (0, ""))
        self.assertTrue(p.stderr)
        self.git("remote", "remove", "origin")
        self.assertIsNone(self.run_hook(self.event()))


if __name__ == "__main__":
    unittest.main()
