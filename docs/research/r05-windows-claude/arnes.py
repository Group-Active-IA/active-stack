"""R-05: arnés para correr Claude Code real (claude -p) contra Engram aislado.

Todo en un directorio temporal con espacios; base de Engram descartable; sin plugins de usuario.
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

PROTO = Path(__file__).resolve().parents[1] / "r01-prototipo"
PY = shutil.which("python")


def git(cwd, *a):
    return subprocess.run(["git", "-C", str(cwd), *a], check=True, capture_output=True, text=True).stdout.strip()


def crear_repo(base, marcador=True):
    origin, repo = base / "origin.git", base / "repo con espacios"
    subprocess.run(["git", "init", "-q", "--bare", str(origin)], check=True)
    repo.mkdir()
    git(repo, "init", "-q", "-b", "main")
    git(repo, "config", "user.name", "Juan Cruz Robledo")
    git(repo, "config", "user.email", "a@b.c")
    (repo / ".engram").mkdir()
    (repo / ".engram" / "config.json").write_text('{"project_name": "r05-synthetic"}\n')
    if marcador:
        (repo / ".engram" / "active-stack.json").write_text('{"memory_policy": "cloud-branch"}\n')
    git(repo, "add", ".")
    git(repo, "commit", "-qm", "init")
    git(repo, "remote", "add", "origin", str(origin))
    git(repo, "push", "-q", "origin", "main")
    git(repo, "remote", "set-head", "origin", "main")
    return repo


def entorno(base):
    env = dict(os.environ)
    env["ENGRAM_DATA_DIR"] = str(base / "db")
    env["ENGRAM_CLOUD_AUTOSYNC"] = "0"
    return env


def config(base, modo):
    """modo: 'hook' (PreToolUse sobre mem_save), 'proxy' (proxy B), 'nada'."""
    mcp = {"command": "engram", "args": ["mcp", "--tools=agent"]}
    settings = {}
    hooks = []
    if modo.startswith("hook") or modo.startswith("ajusteA"):
        script = "hook.py" if modo.startswith("hook") else "hook_ajuste_a.py"
        hooks.append({"type": "command", "command": PY, "args": ["-B", str(PROTO / script)]})
    if modo.endswith("+sim") or modo == "proxy+sim":
        hooks.append({"type": "command", "command": PY, "args": ["-B", str(Path(__file__).with_name("sim_plugin_hook.py"))]})
    if modo.startswith("proxy"):
        mcp = {"command": PY, "args": ["-B", str(PROTO / "proxy_b.py"), "engram", "mcp", "--tools=agent"]}
    if "sim" in modo:
        settings.setdefault("hooks", {})["SessionStart"] = [{"hooks": [{"type": "command", "command": PY,
            "args": ["-B", str(Path(__file__).with_name("sim_session_start.py"))]}]}]
    if hooks:
        settings = {**settings, "hooks": {**settings.get("hooks", {}), "PreToolUse": [{"matcher": "mcp__engram__(mem_save|mem_session_summary|mem_session_finalize|mem_capture_passive|mem_save_prompt)", "hooks": hooks}]}}
    (base / "mcp.json").write_text(json.dumps({"mcpServers": {"engram": mcp}}))
    (base / "settings.json").write_text(json.dumps(settings))
    return base / "mcp.json", base / "settings.json"


def claude(base, repo, prompt, modo="hook", extra=(), model="haiku", timeout=300):
    mcp, settings = config(base, modo)
    cmd = ["claude", "-p", prompt, "--model", model, "--output-format", "stream-json", "--verbose",
           "--mcp-config", str(mcp), "--strict-mcp-config", "--settings", str(settings),
           "--setting-sources", "project,local", "--permission-mode", "bypassPermissions",
           "--no-session-persistence", *extra]
    env = entorno(base)
    server = None
    if "sim" in modo or "--plugin-dir" in extra:
        import socket
        with socket.socket() as so:
            so.bind(("127.0.0.1", 0)); port = so.getsockname()[1]
        env["ENGRAM_PORT"] = str(port)
        server = subprocess.Popen(["engram", "serve", str(port)], env=env, cwd=str(repo),
                                  stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        import time, urllib.request
        for _ in range(40):
            try:
                urllib.request.urlopen(f"http://127.0.0.1:{port}/health", timeout=1); break
            except Exception:
                time.sleep(0.25)
    try:
        run = subprocess.run(cmd, cwd=str(repo), env=env, capture_output=True, text=True,
                             encoding="utf-8", timeout=timeout)
    finally:
        if server:
            server.terminate()
    return run


def buscar(base, repo, query="r05"):
    run = subprocess.run(["engram", "search", query, "--project", "r05-synthetic", "--limit", "20"],
                         cwd=str(repo), env=entorno(base), capture_output=True, text=True, encoding="utf-8")
    t = subprocess.run(["engram", "timeline", "1", "--project", "r05-synthetic"], cwd=str(repo), env=entorno(base),
                       capture_output=True, text=True, encoding="utf-8")
    return run.stdout + run.stderr + "\nTIMELINE\n" + t.stdout


def resumen(run):
    """Extrae llamadas a herramientas, errores y uso de tokens del stream-json."""
    out = {"rc": run.returncode, "tools": [], "turnos": 0, "usage": None, "stderr": run.stderr[-400:]}
    for line in run.stdout.splitlines():
        try:
            e = json.loads(line)
        except ValueError:
            continue
        if e.get("type") == "assistant":
            out["turnos"] += 1
            for c in e["message"].get("content", []):
                if c.get("type") == "tool_use":
                    out["tools"].append({"name": c["name"], "input": c["input"]})
        if e.get("session_id"):
            out["claude_session"] = e["session_id"]
        if e.get("type") == "result":
            out["usage"] = e.get("usage")
            out["result"] = str(e.get("result"))[:300]
            out["is_error"] = e.get("is_error")
    return out


if __name__ == "__main__":
    modo = sys.argv[1] if len(sys.argv) > 1 else "hook"
    base = Path(tempfile.mkdtemp(prefix="r05 q3 ")).resolve()
    repo = crear_repo(base)
    git(repo, "switch", "-qc", "feat/R-05-x")
    p = ("Llamá UNA vez a la herramienta mcp__engram__mem_save con title='r05 canary', "
         "content='canary r05 sintetico', topic_key='k', type='discovery'. Luego respondé solo OK.")
    r = claude(base, repo, p, modo=modo)
    print(json.dumps(resumen(r), indent=1, ensure_ascii=False))
    print(buscar(base, repo))
    print("BASE", base)
