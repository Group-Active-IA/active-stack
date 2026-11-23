import sys, json, glob, sqlite3, subprocess, os
src = open(r"C:/Users/Juan Cruz Robledo/AppData/Local/Temp/r23_mcp.py", encoding="utf-8").read().split("# V1\n")[0]
exec(compile(src, "lib", "exec"))
sys.stdout.reconfigure(encoding="utf-8")


def envp(p, nopath=False):
    e = env(p)
    if nopath:
        e["PATH"] = os.pathsep.join(x for x in e["PATH"].split(os.pathsep) if "active-stack" not in x.lower())
    return e


# V6 detalle
r = repo("n6"); git(r, "switch", "-qc", "feat/narr")
res, err = mcp(r, [("mem_session_summary", dict(content="## Goal\nr23 resumen\n## Accomplished\nnada", project="n6", session_id="s-n6")),
                   ("mem_capture_passive", dict(content="## Key Learnings:\n1. r23 aprendizaje pasivo uno\n2. r23 aprendizaje pasivo dos", project="n6", session_id="s-n6")),
                   ("mem_save_prompt", dict(content="r23 prompt guardado", project="n6", session_id="s-n6"))])
for x in res: print("V6 resp:", json.dumps(x, ensure_ascii=False)[:300])
c = sqlite3.connect(glob.glob(str(DB/"*.db"))[0])
print("V6 obs:", c.execute("select title,topic_key,substr(content,1,70) from observations where project='n6'").fetchall())
print("V6 sessions:", c.execute("select id,project,substr(coalesce(summary,''),1,90) from sessions where project='n6'").fetchall())
print("V6 prompts:", c.execute("select project,substr(content,1,80) from user_prompts where project='n6'").fetchall())

# V8 PATH sin active-stack: el proxy ni arranca
r = repo("n8"); git(r, "switch", "-qc", "feat/nopath")
e = envp("A", True)
print("V8 PATH tiene active-stack:", any("active-stack" in x.lower() for x in e["PATH"].split(os.pathsep)))
try:
    pr = subprocess.run(["active-stack", "--version"], env=e, capture_output=True, text=True)
    print("V8 resuelve active-stack:", pr.returncode)
except Exception as ex:
    print("V8 active-stack no resoluble:", type(ex).__name__)
# guardado por engram directo (lo que pasa si el proxy no arranca y el agente solo tiene engram)
n = len(rows())
env_bak = env
globals()["env"] = lambda p, nopath=False: envp(p, True)
try:
    out = mcp(r, [("mem_save", dict(title="v8", content="r23 v8", topic_key="k", type="discovery", project="n8"))])
    print("V8 proxy via PATH:", out)
except Exception as ex:
    print("V8 el proxy no arranca:", type(ex).__name__, str(ex)[:100])
globals()["env"] = env_bak
print("V8 filas nuevas:", len(rows()) - n)

# hooks de lectura con nombre de evento correcto
r = repo("n13"); git(r, "switch", "-qc", "feature/x")
mcp(r, [("mem_save", dict(title="decision puerto", content="r23 usamos el puerto 4381", topic_key="decision", type="decision", project="n13"))])


def hook2(cwd, ev, hn, p="B", sin_marcador=False):
    r_ = subprocess.run(["active-stack", "memory", "hook", ev], cwd=str(cwd), env=env(p), input=json.dumps({"cwd": str(cwd), "hook_event_name": hn, "session_id": "s1", "tool_name": "mcp__engram__mem_search", "tool_input": {"query": "puerto"}}), capture_output=True, text=True, encoding="utf-8")
    return r_.returncode, r_.stdout[:500], r_.stderr[:200]


git(r, "switch", "-q", "main")
print("V13 session-start main:", hook2(r, "session-start", "SessionStart"))
print("V13 pre-read main:", hook2(r, "pre-read", "PreToolUse"))
git(r, "switch", "-qc", "feature/y")
print("V13 session-start feature/y:", hook2(r, "session-start", "SessionStart"))
res, _ = mcp(r, [("mem_search", dict(query="puerto", project="n13"))], p="B")
print("V13 search B en feature/y:", json.dumps(res, ensure_ascii=False)[:500])
# V9 con stderr: marcador invalido, hook de lectura
(r/".engram"/"active-stack.json").write_text("{no es json")
print("V9 hook con marcador invalido:", hook2(r, "session-start", "SessionStart"))
print("V9 cwd inexistente en hook:", subprocess.run(["active-stack", "memory", "hook", "session-start"], cwd=str(r), env=env("B"), input="{basura", capture_output=True, text=True).returncode)
