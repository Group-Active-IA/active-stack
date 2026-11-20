"""R-05 V12: hook del prototipo en Windows, ruta con espacios, python vs py y CRLF.

Uso: python -B v12_windows.py   (repo sintético en un directorio temporal con espacios)
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

HOOK = Path(__file__).resolve().parents[1] / "r01-prototipo" / "hook.py"


def git(cwd, *args):
    subprocess.run(["git", "-C", str(cwd), *args], check=True, capture_output=True, text=True)


def main():
    base = Path(tempfile.mkdtemp(prefix="r05 espacio ")).resolve()
    try:
        origin, repo = base / "origin.git", base / "repo con espacios"
        subprocess.run(["git", "init", "-q", "--bare", str(origin)], check=True)
        repo.mkdir()
        git(repo, "init", "-q", "-b", "main")
        git(repo, "config", "user.name", "Juan Cruz Robledo")
        git(repo, "config", "user.email", "a@b.c")
        (repo / ".engram").mkdir()
        # Marcador con CRLF a propósito (autocrlf de Git para Windows).
        (repo / ".engram" / "active-stack.json").write_bytes(b'{"memory_policy": "cloud-branch"}\r\n')
        git(repo, "add", ".")
        git(repo, "commit", "-qm", "init")
        git(repo, "remote", "add", "origin", str(origin))
        git(repo, "push", "-q", "origin", "main")
        git(repo, "remote", "set-head", "origin", "main")
        git(repo, "switch", "-qc", "feat/R-05-x")

        event = {"hook_event_name": "PreToolUse", "tool_name": "mcp__plugin_engram_engram__mem_save",
                 "session_id": "s1", "cwd": str(repo),
                 "tool_input": {"title": "T", "content": "C", "topic_key": "k", "session_id": "s1"}}
        out = {}
        for launcher in (["python"], ["py"]):
            # cwd del proceso distinto del repo: el hook debe resolver solo por event["cwd"].
            run = subprocess.run([*launcher, "-B", str(HOOK)], input=json.dumps(event), text=True,
                                 capture_output=True, cwd=str(base))
            out[launcher[0]] = {"rc": run.returncode, "stdout": run.stdout.strip(), "stderr": run.stderr.strip()}
        # Vía cmd.exe con redirección y rutas entre comillas (como lo ejecuta Claude en Windows).
        evt = base / "evt.json"
        evt.write_text(json.dumps(event), encoding="utf-8")
        cmd = f'python -B "{HOOK}" < "{evt}"'
        run = subprocess.run(f'cmd.exe /s /c "{cmd}"', capture_output=True, text=True, cwd=str(base))
        out["cmd.exe"] = {"rc": run.returncode, "stdout": run.stdout.strip(), "stderr": run.stderr.strip()}
        bash = shutil.which("bash")
        if bash:
            posix = lambda q: str(q).replace("\\", "/")
            run = subprocess.run([bash, "-c", f'python -B "{posix(HOOK)}" < "{posix(evt)}"'],
                                 capture_output=True, text=True, cwd=str(base))
            out["git-bash"] = {"rc": run.returncode, "stdout": run.stdout.strip(), "stderr": run.stderr.strip()}
        out["repo"] = str(repo)
        out["marker_bytes"] = (repo / ".engram" / "active-stack.json").read_bytes().decode()
        print(json.dumps(out, indent=2, ensure_ascii=False))
        ok = all(v["rc"] == 0 and "updatedInput" in v["stdout"] and "branch/feat-r-05-x/k" in v["stdout"]
                 for k, v in out.items() if k in ("python", "py", "cmd.exe", "git-bash"))
        print("RESULT", "PASS" if ok else "FAIL")
        return 0 if ok else 1
    finally:
        shutil.rmtree(base, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
