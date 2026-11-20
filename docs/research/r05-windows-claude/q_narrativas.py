"""R-05: vías narrativas con Claude real + plugin 0.1.22: summary, capture_passive, save_prompt."""
import json, os, sqlite3, glob, sys, tempfile
from pathlib import Path
from arnes import claude, crear_repo, git, resumen
PLUGIN = Path(os.environ["TEMP"]) / "engram-upstream-3.7.2" / "plugin" / "claude-code"

def corre(modo):
    base = Path(tempfile.mkdtemp(prefix="r05 nar ")).resolve()
    repo = crear_repo(base); git(repo, "switch", "-qc", "feat/R-05-nar")
    p = ("Hacé estas 3 llamadas EN ORDEN, sin explicar, una por herramienta: "
         "1) mcp__engram__mem_session_summary con content='## Goal\nr05 resumen sintetico'. "
         "2) mcp__engram__mem_capture_passive con content='## Key Learnings:\n- r05 aprendizaje sintetico de prueba' y source='r05'. "
         "3) mcp__engram__mem_save_prompt con content='r05 prompt sintetico'. Respondé OK.")
    r = claude(base, repo, p, modo=modo, extra=["--plugin-dir", str(PLUGIN)], timeout=400)
    s = resumen(r); s.pop("usage", None); s["tools"] = [t["name"] for t in s["tools"]]
    db = sqlite3.connect(glob.glob(str(base / "db" / "*.db"))[0])
    obs = [(t, c[:60]) for t, c in db.execute("select type,content from observations")]
    prompts = [c[:60] for (c,) in db.execute("select content from user_prompts")] if db.execute("select name from sqlite_master where name='user_prompts'").fetchone() else []
    return {"modo": modo, "resumen": s, "observaciones": obs, "prompts": prompts}

if __name__ == "__main__":
    print(json.dumps(corre(sys.argv[1]), indent=1, ensure_ascii=False))
