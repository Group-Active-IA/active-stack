"""R17-D: how does Engram treat a project_name with capitals and Unicode?

Read-only probe (mem_current_project) against an Engram whose data directory,
HOME and USERPROFILE point to a temp folder, so the person's real database is
never opened. Same MCP handshake as docs/research/r01-prototipo/sonda_engram.py.
"""
import json, os, subprocess, sys
from pathlib import Path
from tempfile import TemporaryDirectory

NAMES = ["active-stack", "Mi-App", "Proyecto-Ñandú", "MAYUSCULAS_con_guion_bajo", "mi repo con espacios"]


def call_mcp(repo, env, method, params):
    with subprocess.Popen(["engram", "mcp", "--tools=all"], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                          stderr=subprocess.DEVNULL, text=True, encoding="utf-8", cwd=repo, env=env) as server:
        def send(m):
            server.stdin.write(json.dumps({"jsonrpc": "2.0", **m}) + "\n"); server.stdin.flush()

        def reply(i):
            for line in server.stdout:
                m = json.loads(line)
                if m.get("id") == i:
                    if "error" in m:
                        raise RuntimeError(m["error"])
                    return m["result"]
            raise RuntimeError("engram ended")
        try:
            send({"id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05",
                  "clientInfo": {"name": "r17d-probe", "version": "1"}, "capabilities": {}}})
            reply(1)
            send({"method": "notifications/initialized"})
            send({"id": 2, "method": method, "params": params})
            return reply(2)
        finally:
            server.stdin.close()
            try:
                server.wait(timeout=5)
            except subprocess.TimeoutExpired:
                server.kill()


with TemporaryDirectory() as td:
    base = Path(td)
    env = {"PATH": os.environ["PATH"], "HOME": str(base / "home"), "USERPROFILE": str(base / "home"),
           "ENGRAM_DATA_DIR": str(base / "db"), "ENGRAM_CLOUD_AUTOSYNC": "0",
           "GIT_CONFIG_GLOBAL": str(base / "gc"), "GIT_CONFIG_NOSYSTEM": "1"}
    if "SystemRoot" in os.environ:
        env["SystemRoot"] = os.environ["SystemRoot"]
    (base / "home").mkdir(); (base / "db").mkdir(); (base / "gc").write_text("")
    print(subprocess.run(["engram", "version"], env=env, capture_output=True, text=True).stdout.strip())
    for name in NAMES:
        repo = base / ("r-" + str(abs(hash(name)) % 10**6))
        (repo / ".engram").mkdir(parents=True)
        (repo / ".engram" / "config.json").write_text(json.dumps({"project_name": name}) + "\n", encoding="utf-8")
        subprocess.run(["git", "init", "-q", "-b", "main", str(repo)], env=env, check=True)
        res = call_mcp(repo, env, "tools/call", {"name": "mem_current_project", "arguments": {}})
        text = json.loads(res["content"][0]["text"])
        print(json.dumps({"written": name, "detected": text.get("project"), "source": text.get("project_source")}, ensure_ascii=False))
    # nothing may have been created outside the temp folder
    print("real data dir untouched check: probe used", env["ENGRAM_DATA_DIR"])
