"""Prototipo offline: eventos PreToolUse de mem_save y, desde R-05, de las vías narrativas (resumen, finalize, captura pasiva, prompt)."""
import getpass
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys


PREFIXES = ("mcp__engram__", "mcp__plugin_engram_engram__")
# R-05: vías narrativas; el texto lleva la cabecera, sin tocar clave ni título.
NARRATIVE_FIELDS = {"mem_session_summary": "content", "mem_session_finalize": "summary",
                    "mem_capture_passive": "content", "mem_save_prompt": "content"}


def tool_base(name):
    for prefix in PREFIXES:
        if name.startswith(prefix):
            return name.removeprefix(prefix)
    return None


def make_slug(branch):
    slug = re.sub(r"[^a-z0-9]+", "-", branch.lower()).strip("-")
    if len(slug) > 40:
        slug = slug[:33].rstrip("-") + "-" + hashlib.sha256(branch.encode()).hexdigest()[:6]
    if not slug:
        raise ValueError("rama sin slug")
    return slug


def rewrite(event):
    base = tool_base(event["tool_name"])
    if event["hook_event_name"] != "PreToolUse" or (base != "mem_save" and base not in NARRATIVE_FIELDS):
        return None
    if base in NARRATIVE_FIELDS and not isinstance(event["tool_input"].get(NARRATIVE_FIELDS[base]), str):
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
    on_branch = branch not in ([primary.removeprefix("origin/")] if primary else ["main", "master"])
    slug = make_slug(branch) if on_branch else None
    if base in NARRATIVE_FIELDS:
        field = NARRATIVE_FIELDS[base]
        if base == "mem_capture_passive":
            # Engram conserva solo los ítems de la lista: se etiqueta cada ítem, y en la rama principal no hay nada que añadir.
            if not slug:
                return None
            fields[field] = re.sub(r"(?m)^(\s*(?:[-*]|\d+[.)])\s+)", lambda m: f"{m.group(1)}[{slug}] ", fields[field])
        else:
            if on_branch:
                header = f"Branch: {branch} · {header}"
            fields[field] = f"{header}\n{fields[field]}"
        return {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow",
                                       "updatedInput": fields}}
    if on_branch:
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
