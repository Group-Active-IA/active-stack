import json, os, sqlite3, glob, subprocess, sys, threading, shutil
from pathlib import Path
W = Path(r"C:/Users/Juan Cruz Robledo/AppData/Local/Temp/r23e2e")
HOMES = {"A": W/"homeA", "B": W/"homeB"}
DB = W/"db"
SAVE = "mcp__engram__mem_save"


def env(p, nopath=False):
    h = HOMES[p]
    e = dict(os.environ)
    e.update(HOME=str(h), USERPROFILE=str(h), LOCALAPPDATA=str(h/"AppData/Local"), APPDATA=str(h/"AppData/Roaming"),
             ENGRAM_DATA_DIR=str(DB), ENGRAM_CLOUD_AUTOSYNC="0")
    if nopath:
        e["PATH"] = os.pathsep.join(x for x in e["PATH"].split(os.pathsep) if "AppData\\Local\\active-stack" not in x and "AppData/Local/active-stack" not in x)
    return e


def git(cwd, *a):
    return subprocess.run(["git", "-C", str(cwd), *a], check=True, capture_output=True, text=True).stdout.strip()


def repo(name, marker=True):
    base = W/name
    shutil.rmtree(base, ignore_errors=True)
    base.mkdir()
    origin, r = base/"origin.git", base/"repo con espacios"
    subprocess.run(["git", "init", "-q", "--bare", str(origin)], check=True)
    r.mkdir()
    git(r, "init", "-q", "-b", "main")
    git(r, "config", "user.name", "Persona A")
    git(r, "config", "user.email", "a@b.c")
    if marker:
        o = subprocess.run(["active-stack", "memory", "init", "--project", str(r), "--name", name, "--yes"], env=env("A"), capture_output=True, text=True)
        print("memory init:", o.returncode, o.stdout[-200:], o.stderr[-200:])
    else:
        (r/".engram").mkdir()
        (r/".engram"/"config.json").write_text(json.dumps({"project_name": name})+"\n")
    git(r, "add", ".")
    git(r, "commit", "-qm", "init")
    git(r, "remote", "add", "origin", str(origin))
    git(r, "push", "-q", "origin", "main")
    git(r, "remote", "set-head", "origin", "main")
    return r


def rows():
    f = glob.glob(str(DB/"*.db"))
    if not f:
        return []
    c = sqlite3.connect(f[0])
    q = "select id,session_id,type,title,topic_key,project,content from observations where deleted_at is null order by id"
    return [dict(zip(("id", "sesion", "tipo", "titulo", "clave", "proyecto", "contenido"), r)) for r in c.execute(q)]


def claude(p, cwd, prompt, model="haiku", nopath=False, timeout=360):
    cmd = ["claude", "-p", prompt, "--model", model, "--output-format", "stream-json", "--verbose", "--setting-sources", "user,project",
           "--permission-mode", "bypassPermissions", "--no-session-persistence"]
    r = subprocess.run(cmd, cwd=str(cwd), env=env(p, nopath), capture_output=True, text=True, encoding="utf-8", timeout=timeout)
    out = {"rc": r.returncode, "tools": [], "stderr": r.stderr[-300:]}
    for line in r.stdout.splitlines():
        try:
            e = json.loads(line)
        except ValueError:
            continue
        if e.get("type") == "assistant":
            for c in e["message"].get("content", []):
                if c.get("type") == "tool_use":
                    out["tools"].append(c["name"])
        if e.get("type") == "result":
            out["result"] = str(e.get("result"))[:300]
            out["cost"] = e.get("total_cost_usd")
    return out


def g(t, k="k", content=None):
    return f"llamá a {SAVE} con title='{t}', content='{content or 'r23 '+t}', topic_key='{k}', type='discovery'"


def show(n, res, n0):
    print("=====", n)
    print(json.dumps(res, ensure_ascii=False))
    for r in rows()[n0:]:
        print("  ROW", r["id"], r["proyecto"], "|", r["titulo"], "|", r["clave"], "|", r["contenido"].splitlines()[0][:90])


esc = sys.argv[1]
n0 = len(rows())
if esc == "V1":
    r = repo("r23v1")
    git(r, "switch", "-qc", "feature/x")
    show(esc, claude("A", r, "Sin explicar: "+g("v1 decision")+". Respondé OK."), n0)
if esc == "V2V3":
    r = repo("r23v2")
    git(r, "switch", "-qc", "feat/a")
    show(esc, claude("A", r, "Hacé EN ORDEN sin explicar: 1) "+g("v2 paso1")+". 2) Con Bash: git switch -c feat/otra. 3) "+g("v2 paso2")+". 4) Con Bash: git checkout -b feat/tercera. 5) "+g("v2 paso3")+". 6) Con Bash: git switch main. 7) "+g("v3 paso4")+". Respondé OK."), n0)
if esc == "V4":
    r = repo("r23v4")
    b = r.parent
    wa, wb = b/"wt-a", b/"wt-b"
    git(r, "worktree", "add", "-qb", "feat/wa", str(wa))
    git(r, "worktree", "add", "-qb", "feat/wb", str(wb))
    out = {}

    def run(n, w):
        out[n] = claude("A", w, "Sin explicar: "+g("v4 "+n)+". Respondé OK.")
    ts = [threading.Thread(target=run, args=(n, w)) for n, w in (("a", wa), ("b", wb))]
    [t.start() for t in ts]
    [t.join() for t in ts]
    show(esc, out, n0)
if esc == "V5":
    r = repo("r23v5")
    git(r, "switch", "-qc", "feat/sub")
    show(esc, claude("A", r, "Usá la herramienta Agent (subagente general-purpose) y pedile que llame a "+SAVE+" con title='v5 desde subagente', content='r23 sub', topic_key='k', type='discovery'. Respondé OK al terminar."), n0)
if esc == "V6":
    r = repo("r23v6")
    git(r, "switch", "-qc", "feat/narr")
    show(esc, claude("A", r, "Sin explicar, EN ORDEN: 1) llamá a mcp__engram__mem_session_summary con content='## Goal\\nr23 resumen\\n## Accomplished\\nnada' y project='r23v6'. 2) llamá a mcp__engram__mem_capture_passive con content='## Key Learnings:\\n1. r23 aprendizaje pasivo uno\\n2. r23 aprendizaje pasivo dos' y project='r23v6'. 3) llamá a mcp__engram__mem_save_prompt con content='r23 prompt guardado' y project='r23v6'. Respondé OK."), n0)
    c = sqlite3.connect(glob.glob(str(DB/"*.db"))[0])
    print("  PROMPTS", c.execute("select content from user_prompts").fetchall())
if esc == "V7":
    r = repo("r23v7", marker=False)
    git(r, "switch", "-qc", "feat/sin-marcador")
    show(esc, claude("A", r, "Sin explicar: "+g("v7 sin marcador")+". OK."), n0)
if esc == "V8":
    r = repo("r23v8")
    git(r, "switch", "-qc", "feat/nopath")
    show(esc, claude("A", r, "Sin explicar: "+g("v8 sin active-stack en PATH")+". OK.", nopath=True), n0)
if esc == "V9":
    r = repo("r23v9")
    git(r, "switch", "-qc", "feat/m")
    (r/".engram"/"active-stack.json").write_text("{no es json")
    show(esc, claude("A", r, "Sin explicar: "+g("v9 marcador malo")+". OK."), n0)
if esc == "V11":
    r = repo("r23v11")
    git(r, "switch", "-qc", "feat/una")
    steps = [("una", lambda: None, "k"),
             ("guion", lambda: (git(r, "switch", "-q", "main"), git(r, "switch", "-qc", "feat-una")), "k"),
             ("detached", lambda: (git(r, "switch", "-q", "main"), git(r, "checkout", "-q", "--detach")), "kd"),
             ("largo", lambda: (git(r, "switch", "-q", "main"), git(r, "switch", "-qc", "feature/"+"x"*60+"-final")), "kl")]
    for lab, setup, k in steps:
        setup()
        show(esc+lab, claude("A", r, "Sin explicar: "+g("v11 "+lab, k)+". OK."), n0)
if esc == "V13":
    r = repo("r23v13")
    git(r, "switch", "-qc", "feature/x")
    claude("A", r, "Sin explicar: "+g("decision puerto", "decision", "r23 usamos el puerto 4381")+". OK.")
    git(r, "switch", "-q", "main")
    res = claude("B", r, "Usá mcp__engram__mem_search con query='puerto' y project='r23v13'. Después decime en una frase: ¿qué puerto está decidido como verdad del equipo y qué rama lo originó?")
    show(esc+"-main", res, n0)
    git(r, "switch", "-qc", "feature/y")
    res = claude("B", r, "Usá mcp__engram__mem_search con query='puerto' y project='r23v13'. Después decime en una frase si lo encontrado es verdad del equipo o no verificado, y por qué.")
    show(esc+"-feat-y", res, n0)
