"""Q3 con el plugin REAL de Engram (upstream v3.7.2, plugin 0.1.22) cargado con --plugin-dir."""
import json, os, sys, tempfile
from pathlib import Path
import arnes
from arnes import claude, crear_repo, git, resumen
from escenarios import filas

PLUGIN = Path(os.environ["TEMP"]) / "engram-upstream-3.7.2" / "plugin" / "claude-code"

def corre(modo, rama="feat/R-05-real"):
    base = Path(tempfile.mkdtemp(prefix="r05 real ")).resolve()
    repo = crear_repo(base)
    git(repo, "switch", "-qc", rama)
    p = "Llamá UNA vez a mcp__engram__mem_save con title='r05 canary', content='canary r05 real', topic_key='k', type='discovery'. Respondé OK."
    r = claude(base, repo, p, modo=modo, extra=["--plugin-dir", str(PLUGIN)])
    s = resumen(r); s.pop("usage", None); s["tools"] = [t["name"] for t in s["tools"]]
    f = filas(base)
    ok = bool(f) and f[0]["sesion"] == s.get("claude_session") and f[0]["titulo"].startswith("[")
    log = base / "hook.log"
    hl = [json.loads(x) for x in log.read_text(encoding="utf-8").splitlines()] if log.exists() else []
    return {"modo": modo, "etiquetado_y_sesion_correcta": ok, "hook_log": hl, "resumen": s, "filas": f}

if __name__ == "__main__":
    print(json.dumps(corre(sys.argv[1]), indent=1, ensure_ascii=False, default=str))
