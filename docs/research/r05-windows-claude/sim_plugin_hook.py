"""Sustituto del PreToolUse del plugin de Engram 0.1.5 (no instalado aquí): agrega session_id explícito."""
import json, sys
e = json.load(sys.stdin)
f = dict(e["tool_input"]); f["session_id"] = "sim-" + e["session_id"]
print(json.dumps({"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow", "updatedInput": f}}))
