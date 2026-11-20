"""Envoltorio de diagnóstico: ejecuta hook.py y registra tiempo, código y si hubo salida (R05_LOG)."""
import json, os, subprocess, sys, time
from pathlib import Path
entrada = sys.stdin.read()
t0 = time.time()
r = subprocess.run([sys.executable, "-B", str(Path(__file__).resolve().parents[1] / "r01-prototipo" / "hook.py")],
                   input=entrada, capture_output=True, text=True)
with open(os.environ["R05_LOG"], "a", encoding="utf-8") as f:
    f.write(json.dumps({"inicio": t0, "seg": round(time.time() - t0, 2), "rc": r.returncode,
                        "salida": bool(r.stdout.strip()), "stderr": r.stderr.strip()[:120]}) + "\n")
sys.stdout.write(r.stdout)
sys.stderr.write(r.stderr)
