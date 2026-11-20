"""Proxy stdio del piloto R-01; el proceso upstream responde directamente al cliente."""
import json
import os
import re
from pathlib import Path
import subprocess
import sys

from hook import rewrite


def trace_test_call(incoming, upstream):
    markers = os.getenv("R01_TRACE_MARKERS", "").split(",")
    target = os.getenv("R01_TRACE_FILE")
    if not target or len(markers) != 2 or not all(re.fullmatch(r"r01b-[a-f0-9]{32}", m) for m in markers):
        return
    if incoming.get("title") not in markers:
        return
    allowed = ("session_id", "title", "topic_key", "content", "project", "type", "scope")
    record = {"cwd": str(Path.cwd()), "incoming": {k: incoming[k] for k in allowed if k in incoming},
              "upstream": {k: upstream[k] for k in allowed if k in upstream}}
    with open(target, "a", encoding="utf-8") as output:
        output.write(json.dumps(record) + "\n")


def transform(line):
    msg = json.loads(line)
    if not isinstance(msg, dict) or msg.get("method") != "tools/call":
        return line
    params = msg.get("params")
    if not isinstance(params, dict) or params.get("name") != "mem_save":
        return line
    args = params.get("arguments")
    if not isinstance(args, dict):
        return line
    try:
        result = rewrite({"hook_event_name": "PreToolUse", "tool_name": "mcp__engram__mem_save",
                          "tool_input": args, "cwd": os.getcwd()})
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        print(f"Etiquetado omitido ({type(error).__name__})", file=sys.stderr)
        return line
    if not result:
        return line
    msg["params"]["arguments"] = result["hookSpecificOutput"]["updatedInput"]
    trace_test_call(args, msg["params"]["arguments"])
    return (json.dumps(msg) + "\n").encode()


if __name__ == "__main__":
    with subprocess.Popen(sys.argv[1:], stdin=subprocess.PIPE) as upstream:
        for line in sys.stdin.buffer:
            upstream.stdin.write(transform(line))
            upstream.stdin.flush()
        upstream.stdin.close()
        sys.exit(upstream.wait())
