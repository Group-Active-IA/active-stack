import sys, json, glob, sqlite3, subprocess, os
src = open(r"C:/Users/Juan Cruz Robledo/AppData/Local/Temp/r23_mcp.py", encoding="utf-8").read().split("# V1\n")[0]
exec(compile(src, "lib", "exec"))
sys.stdout.reconfigure(encoding="utf-8")
_env0 = env
def env(p, nopath=False):
    e = _env0(p)
    if nopath:
        e["PATH"] = os.pathsep.join(x for x in e["PATH"].split(os.pathsep) if "active-stack" not in x.lower())
    return e
r = repo("p6"); git(r, "switch", "-qc", "feat/narr")
res, err = mcp(r, [("mem_session_summary", dict(content="## Goal\nr23 resumen\n## Accomplished\nnada", project="p6")),
                   ("mem_capture_passive", dict(content="## Key Learnings:\n1. r23 aprendizaje pasivo uno\n2. r23 aprendizaje pasivo dos", project="p6")),
                   ("mem_save_prompt", dict(content="r23 prompt guardado", project="p6"))])
for x in res: print("V6 resp:", json.dumps(x, ensure_ascii=False)[:250])
c = sqlite3.connect(glob.glob(str(DB/"*.db"))[0])
print("V6 obs:", c.execute("select title,topic_key,substr(content,1,80) from observations where project='p6'").fetchall())
print("V6 sessions:", c.execute("select id,substr(coalesce(summary,''),1,100) from sessions where project='p6'").fetchall())
print("V6 prompts:", c.execute("select substr(content,1,80) from user_prompts where project='p6'").fetchall())
r = repo("p8"); git(r, "switch", "-qc", "feat/nopath"); n = len(rows())
try:
    out = mcp(r, [("mem_save", dict(title="v8", content="r23 v8", topic_key="k", type="discovery", project="p8"))], nopath=True)
    print("V8 salida:", json.dumps(out)[:300])
except Exception as ex:
    print("V8 el proxy no arranca:", type(ex).__name__, str(ex)[:150])
print("V8 filas nuevas:", len(rows()) - n)
