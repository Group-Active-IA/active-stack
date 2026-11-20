"""Observador del piloto. No emite decisiones ni argumentos al runtime."""
import json
from pathlib import Path
import sys


def id_observacion(value):
    if isinstance(value, str):
        try:
            value = json.loads(value)
        except ValueError:
            return None
    if isinstance(value, list):
        for item in value:
            found = id_observacion(item)
            if found is not None:
                return found
    if isinstance(value, dict):
        if type(value.get("id")) is int and isinstance(value.get("result"), str):
            return value["id"]
        for field in ("content", "text", "result", "output"):
            if field in value:
                found = id_observacion(value[field])
                if found is not None:
                    return found
    return None


def extract(event, keys):
    phase = event.get("hook_event_name")
    if phase not in ("PreToolUse", "PostToolUse") or event.get("tool_name") != "mcp__engram__mem_save":
        return None
    if event.get("tool_input", {}).get("topic_key") not in keys:
        return None
    record = {k:event[k] for k in ("tool_use_id","session_id","agent_id") if k in event}
    record["phase"] = event["hook_event_name"]
    if phase == "PostToolUse":
        oid = id_observacion(event.get("tool_response"))
        if oid is not None:
            record["observation_id"] = oid
    return record


if __name__ == "__main__":
    try:
        folder = Path(__file__).resolve().parent
        event = json.load(sys.stdin)
        result = extract(event, json.loads((folder / "keys.json").read_text()))
        if result:
            with (folder / "metadata.jsonl").open("a") as stream:
                stream.write(json.dumps(result) + "\n")
    except (OSError, AttributeError, ValueError, TypeError, KeyError):
        pass
