"""R-05: escenarios V2-V13, Q6, Q9, Q11 con Claude Code real (claude -p). Resultado en JSON."""
import json
import sqlite3
import glob
import subprocess
import sys
import tempfile
import threading
from pathlib import Path

import arnes
from arnes import claude, crear_repo, git, resumen

SAVE = "mcp__engram__mem_save"


def filas(base):
    f = glob.glob(str(base / "db" / "*.db"))
    if not f:
        return []
    c = sqlite3.connect(f[0])
    return [dict(zip(("id", "sesion", "tipo", "titulo", "clave", "rev", "contenido"), r)) for r in c.execute(
        "select id,session_id,type,title,topic_key,revision_count,content from observations where deleted_at is null")]


def nuevo(marcador=True, rama=None):
    base = Path(tempfile.mkdtemp(prefix="r05 esc ")).resolve()
    repo = crear_repo(base, marcador)
    if rama:
        git(repo, "switch", "-qc", rama)
    return base, repo


def corre(base, repo, prompt, modo, **kw):
    r = claude(base, repo, prompt, modo=modo, **kw)
    s = resumen(r)
    s["tools"] = [t["name"] for t in s["tools"]]
    s["usage"] = {k: (s["usage"] or {}).get(k) for k in ("input_tokens", "cache_creation_input_tokens", "cache_read_input_tokens", "output_tokens")}
    return s


def guardar(titulo, clave="k"):
    return f"llamá a {SAVE} con title='{titulo}', content='r05 {titulo}', topic_key='{clave}', type='discovery'"


def v2_v3(modo):
    base, repo = nuevo(rama="feat/R-05-a")
    p = ("Hacé EN ORDEN y sin explicar: 1) " + guardar("paso1") + ". 2) Con Bash: git switch -c feat/otra. 3) " + guardar("paso2") +
         ". 4) Con Bash: git switch main. 5) " + guardar("paso3") + ". Respondé OK.")
    return {"resumen": corre(base, repo, p, modo), "filas": filas(base)}


def v4(modo):
    base, repo = nuevo()
    wa, wb = base / "wt-a", base / "wt-b"
    git(repo, "worktree", "add", "-qb", "feat/wa", str(wa))
    git(repo, "worktree", "add", "-qb", "feat/wb", str(wb))
    res = {}

    def run(nombre, wt):
        res[nombre] = corre(base, wt, "Sin explicar: " + guardar("simultaneo") + ". Respondé OK.", modo)
    ts = [threading.Thread(target=run, args=(n, w)) for n, w in (("a", wa), ("b", wb))]
    [t.start() for t in ts]
    [t.join() for t in ts]
    return {"resumen": res, "filas": filas(base)}


def v5_q6(modo):
    base, repo = nuevo(rama="feat/R-05-sub")
    p = ("Usá la herramienta Agent (subagente general-purpose) y pedile que llame a " + SAVE +
         " con title='desde subagente', content='r05 sub', topic_key='k', type='discovery'. Respondé OK al terminar.")
    return {"resumen": corre(base, repo, p, modo), "filas": filas(base)}


def v7(modo):
    base, repo = nuevo(marcador=False, rama="feat/sin-marcador")
    return {"resumen": corre(base, repo, "Sin explicar: " + guardar("sin marcador") + ". OK.", modo), "filas": filas(base)}


def v9_timeout():
    """Timeout EXTERNO de PreToolUse: hook que duerme 10 s con timeout=2 s."""
    base, repo = nuevo(rama="feat/R-05-t")
    hook = {"type": "command", "command": arnes.PY, "timeout": 2,
            "args": ["-c", "import time;time.sleep(60)"]}
    st = {"hooks": {"PreToolUse": [{"matcher": SAVE, "hooks": [hook]}]}}
    (base / "mcp.json").write_text(json.dumps({"mcpServers": {"engram": {"command": "engram", "args": ["mcp", "--tools=agent"]}}}))
    (base / "settings.json").write_text(json.dumps(st))
    import time
    t0 = time.time()
    r = subprocess.run(["claude", "-p", "Sin explicar: " + guardar("timeout") + ". OK.", "--model", "haiku", "--output-format", "stream-json",
                        "--verbose", "--mcp-config", str(base / "mcp.json"), "--strict-mcp-config", "--settings", str(base / "settings.json"),
                        "--setting-sources", "project,local", "--permission-mode", "bypassPermissions", "--no-session-persistence"],
                       cwd=str(repo), env=arnes.entorno(base), capture_output=True, text=True, encoding="utf-8", timeout=300)
    s = resumen(r)
    s["tools"] = [t["name"] for t in s["tools"]]
    s.pop("usage", None)
    return {"resumen": s, "segundos": round(time.time() - t0, 1), "filas": filas(base)}


def v9_marcador_invalido(modo):
    base, repo = nuevo(rama="feat/R-05-m")
    (repo / ".engram" / "active-stack.json").write_text("{no es json")
    return {"resumen": corre(base, repo, "Sin explicar: " + guardar("marcador malo") + ". OK.", modo), "filas": filas(base)}


def v11(modo):
    base, repo = nuevo(rama="feat/una")
    p1 = "Sin explicar: " + guardar("una") + ". OK."
    out = {"feat/una": corre(base, repo, p1, modo)}
    git(repo, "switch", "-q", "main")
    git(repo, "switch", "-qc", "feat-una")
    out["feat-una"] = corre(base, repo, "Sin explicar: " + guardar("una-guion") + ". OK.", modo)
    git(repo, "switch", "-q", "main")
    git(repo, "checkout", "-q", "--detach")
    out["detached"] = corre(base, repo, "Sin explicar: " + guardar("detached", "kd") + ". OK.", modo)
    git(repo, "switch", "-q", "main")
    largo = "feature/" + "x" * 60 + "-final"
    git(repo, "switch", "-qc", largo)
    out["largo"] = corre(base, repo, "Sin explicar: " + guardar("largo", "kl") + ". OK.", modo)
    return {"resumen": out, "filas": filas(base)}


def v13(modo):
    base, repo = nuevo(rama="feat/R-05-lect")
    corre(base, repo, "Sin explicar: " + guardar("decision puerto", "decision") + " con content='r05 usamos el puerto 4381'. OK.", modo)
    git(repo, "switch", "-q", "main")
    r = corre(base, repo, "Usá mem_search (mcp__engram__mem_search) con query='puerto' y decime solo qué puerto se decidió y en qué rama.", modo,
              extra=["--output-format", "json"] if False else [])
    return {"resumen": r, "filas": filas(base)}


def q9():
    """Permisos: ¿git branch --show-current pide confirmación en modo default (sin bypass)?"""
    base, repo = nuevo(rama="feat/R-05-p")
    res = {}
    for modo_perm in ("default", "plan", "acceptEdits"):
        cmd = ["claude", "-p", "Ejecutá con Bash exactamente: git branch --show-current . Respondé solo el nombre de la rama.", "--model", "haiku",
               "--output-format", "json", "--permission-mode", modo_perm, "--no-session-persistence", "--setting-sources", "project,local"]
        r = subprocess.run(cmd, cwd=str(repo), capture_output=True, text=True, encoding="utf-8", timeout=200)
        try:
            j = json.loads(r.stdout)
        except ValueError:
            j = {"raw": r.stdout[:200], "err": r.stderr[:200]}
        res[modo_perm] = {"result": str(j.get("result"))[:120], "permission_denials": j.get("permission_denials")}
    return res


def q11(par):
    out = {}
    for modo in ("hook", "proxy"):   # 'hook' = sin B (guardado directo reescrito por hook), 'proxy' = con B
        base, repo = nuevo(rama="feat/R-05-q11")
        p = ("Sin explicar y en este orden: " + guardar("q11a", "a") + "; " + guardar("q11b", "b") + "; " + guardar("q11c", "c") + ". Respondé OK.")
        m = "nada" if modo == "hook" else "proxy"
        out["sin_B" if modo == "hook" else "con_B"] = corre(base, repo, p, m)
    return out


if __name__ == "__main__":
    nombre = sys.argv[1]
    modo = sys.argv[2] if len(sys.argv) > 2 else "proxy"
    f = globals()[nombre]
    try:
        res = f(modo) if nombre not in ("v9_timeout", "q9", "q11") else (f() if nombre != "q11" else f(int(modo) if modo.isdigit() else 0))
    except Exception as e:  # registrar el fallo, no ocultarlo
        res = {"error": repr(e)}
    print(json.dumps(res, indent=1, ensure_ascii=False, default=str))
