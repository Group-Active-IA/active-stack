"""Prototipo offline, limitado a eventos PreToolUse de mem_save."""
import getpass
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys


def rewrite(event):
    if event["hook_event_name"] != "PreToolUse" or event["tool_name"] not in (
        "mcp__engram__mem_save", "mcp__plugin_engram_engram__mem_save"
    ):
        return None
    cwd = Path(event["cwd"])

    def git(*args, optional=False):
        result = subprocess.run(["git", "-C", cwd, *args], text=True, timeout=2,
                                capture_output=True, check=not optional)
        return result.stdout.strip()

    marker = Path(git("rev-parse", "--show-toplevel")) / ".engram" / "active-stack.json"
    if not marker.is_file() or json.loads(marker.read_text()).get("memory_policy") != "cloud-branch":
        return None
    if not git("remote"):
        return None
    primary = git("symbolic-ref", "--short", "refs/remotes/origin/HEAD", optional=True)
    branch = git("branch", "--show-current") or "detached-" + git("rev-parse", "HEAD")[:7]
    author = git("config", "user.name", optional=True) or getpass.getuser()
    fields = dict(event["tool_input"])
    header = f"Author: {author}"
    if branch not in ([primary.removeprefix("origin/")] if primary else ["main", "master"]):
        slug = re.sub(r"[^a-z0-9]+", "-", branch.lower()).strip("-")
        if len(slug) > 40:
            slug = slug[:33].rstrip("-") + "-" + hashlib.sha256(branch.encode()).hexdigest()[:6]
        if not slug:
            raise ValueError("rama sin slug")
        if fields.get("topic_key"):
            fields["topic_key"] = f'branch/{slug}/{fields["topic_key"]}'
        fields["title"] = f'[{slug}] {fields["title"]}'
        header = f"Branch: {branch} · {header}"
    fields["content"] = f'{header}\n{fields["content"]}'
    return {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow",
                                   "updatedInput": fields}}


if __name__ == "__main__":
    try:
        result = rewrite(json.load(sys.stdin))
        if result:
            print(json.dumps(result))
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        print(f"Prototipo sin reescritura ({type(error).__name__})", file=sys.stderr)
