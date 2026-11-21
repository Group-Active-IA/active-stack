"""R-15: repite la carga de R-05 (Q3) con el plugin real de Engram 0.1.22.

Uso: python -B carga_r15.py <hooklog|proxy> <n> <salida_dir>
Corre n sesiones `claude -p` en paralelo con el arnés de R-05. Requiere el plugin en
<esta carpeta>/upstream/engram-3.7.2 (git clone --depth 1 --branch v3.7.2
https://github.com/Gentleman-Programming/engram.git, commit c6ba9e3).
"""
import json
import shutil
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI))
import q3_real  # noqa: E402

q3_real.PLUGIN = AQUI / "upstream" / "engram-3.7.2" / "plugin" / "claude-code"

# Una instalación npm deja `claude.cmd`, que subprocess no lanza sin shell: se usa el ejecutable que envuelve.
_EXE = shutil.which("claude") or "claude"
if _EXE.lower().endswith(".cmd"):
    _EXE = str(Path(_EXE).parent / "node_modules" / "@anthropic-ai" / "claude-code" / "bin" / "claude.exe")
_run = subprocess.run


def _run_claude(cmd, *a, **k):
    if isinstance(cmd, list) and cmd and cmd[0] == "claude":
        cmd = [_EXE, *cmd[1:]]
    return _run(cmd, *a, **k)


subprocess.run = _run_claude


def una(modo, out, i):
    try:
        r = q3_real.corre(modo)
    except Exception as e:  # una corrida rota no frena las demás
        r = {"modo": modo, "error": repr(e)}
    (out / f"{modo}-{i}.json").write_text(json.dumps(r, indent=1, ensure_ascii=False, default=str), encoding="utf-8")
    f = r.get("filas") or []
    hl = r.get("hook_log") or []
    return {"i": i, "ok": r.get("etiquetado_y_sesion_correcta"), "fila": bool(f),
            "titulo": f[0]["titulo"] if f else None, "hook_seg": [h["seg"] for h in hl],
            "hook_salida": [h["salida"] for h in hl], "error": r.get("error")}


if __name__ == "__main__":
    modo, n, out = sys.argv[1], int(sys.argv[2]), Path(sys.argv[3])
    out.mkdir(parents=True, exist_ok=True)
    with ThreadPoolExecutor(max_workers=n) as ex:
        res = list(ex.map(lambda i: una(modo, out, i), range(1, n + 1)))
    for x in res:
        print(json.dumps(x, ensure_ascii=False))
    print("RESUMEN", modo, sum(1 for x in res if x["ok"]), "/", n)
