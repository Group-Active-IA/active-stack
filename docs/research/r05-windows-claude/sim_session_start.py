"""Sustituto del SessionStart del plugin 0.1.0: crea la sesión en Engram (POST /sessions)."""
import json, os, sys, urllib.request
e = json.load(sys.stdin)
for sid in ("sim-" + e["session_id"], e["session_id"]):  # la real se crea última
    body = json.dumps({"id": sid, "project": "r05-synthetic", "directory": e["cwd"]}).encode()
    req = urllib.request.Request(f"http://127.0.0.1:{os.environ['ENGRAM_PORT']}/sessions", body, {"Content-Type": "application/json"})
    urllib.request.urlopen(req, timeout=5).read()
