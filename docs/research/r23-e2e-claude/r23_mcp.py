"""R-23 sin LLM: maneja el proxy REAL (`active-stack memory proxy -- engram mcp`) por MCP stdio y los hooks reales."""
import json, subprocess, sys, glob, sqlite3, os, shutil, threading, time
sys.argv = [sys.argv[0], "NONE"]
import importlib.util
spec = importlib.util.spec_from_file_location("r", r"C:/Users/Juan Cruz Robledo/AppData/Local/Temp/r23_run.py")
src = open(spec.origin, encoding="utf-8").read().split("esc = sys.argv[1]")[0]
ns = {}
exec(compile(src, "r23_run_lib", "exec"), ns)
env, git, repo, rows, DB, W = ns["env"], ns["git"], ns["repo"], ns["rows"], ns["DB"], ns["W"]


def mcp(cwd, calls, p="A", nopath=False, cmd=None):
    """calls: lista de (tool, args). Devuelve respuestas."""
    cmd = cmd or ["active-stack", "memory", "proxy", "--", "engram", "mcp"]
    pr = subprocess.Popen(cmd, cwd=str(cwd), env=env(p, nopath), stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, encoding="utf-8")
    msgs = [{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "r23", "version": "1"}}},
            {"jsonrpc": "2.0", "method": "notifications/initialized"}]
    for i, (t, a) in enumerate(calls):
        msgs.append({"jsonrpc": "2.0", "id": 10+i, "method": "tools/call", "params": {"name": t, "arguments": a}})
    out = []
    try:
        for m in msgs:
            pr.stdin.write(json.dumps(m)+"\n"); pr.stdin.flush()
            if "id" in m:
                out.append(json.loads(pr.stdout.readline()))
    finally:
        pr.stdin.close()
        try:
            pr.wait(timeout=10)
        except Exception:
            pr.kill()
    return out[1:], pr.stderr.read()[-300:]


def save(cwd, t, k="k", content=None, **kw):
    return mcp(cwd, [("mem_save", dict(title=t, content=content or "r23 "+t, topic_key=k, type="discovery", project=cwd_proj(cwd), **kw))], **{x: y for x, y in kw.items() if x in ("p", "nopath")})


def cwd_proj(cwd):
    return json.load(open(next(p for p in [cwd/".engram"/"config.json"]), encoding="utf-8"))["project_name"].lower()


def dump(n0, label):
    print("==", label)
    for r in rows()[n0:]:
        print("  ROW", r["id"], r["sesion"], "|", r["titulo"], "|", r["clave"], "|", r["contenido"].splitlines()[0][:90])


def hook(cwd, ev, p="A"):
    r = subprocess.run(["active-stack", "memory", "hook", ev], cwd=str(cwd), env=env(p), input=json.dumps({"cwd": str(cwd), "hook_event_name": ev, "session_id": "s1", "tool_name": "mcp__engram__mem_search"}), capture_output=True, text=True, encoding="utf-8")
    return r.returncode, r.stdout[:400], r.stderr[:200]


def S(n):
    return len(rows())


# V1
r = repo("m1"); git(r, "switch", "-qc", "feature/x"); n = S(0)
print(save(r, "v1 decision")[1] or "ok"); dump(n, "V1 feature/x")
# V2/V3 (cambio de rama entre llamadas del mismo proceso proxy)
r = repo("m2"); git(r, "switch", "-qc", "feat/a"); n = S(0)
pr = subprocess.Popen(["active-stack", "memory", "proxy", "--", "engram", "mcp"], cwd=str(r), env=env("A"), stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True, encoding="utf-8")
def rpc(i, m, a=None):
    pr.stdin.write(json.dumps({"jsonrpc": "2.0", "id": i, "method": m, "params": a or {}})+"\n"); pr.stdin.flush(); return pr.stdout.readline()
rpc(1, "initialize", {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "r23", "version": "1"}})
pr.stdin.write(json.dumps({"jsonrpc": "2.0", "method": "notifications/initialized"})+"\n"); pr.stdin.flush()
def sv(i, t): return rpc(i, "tools/call", {"name": "mem_save", "arguments": dict(title=t, content="r23 "+t, topic_key="k", type="discovery", project="m2")})
sv(2, "v2 paso1"); git(r, "switch", "-qc", "feat/otra"); sv(3, "v2 paso2 (git switch)")
git(r, "switch", "-q", "feat/a"); git(r, "checkout", "-q", "-b", "feat/tercera"); sv(4, "v2 paso3 (git checkout -b)")
git(r, "switch", "-q", "main"); sv(5, "v3 paso4 main")
pr.stdin.close(); pr.wait(timeout=10); dump(n, "V2/V3 (mismo proceso proxy)")
# V4 worktrees
r = repo("m4"); b = r.parent; wa, wb = b/"wt-a", b/"wt-b"
git(r, "worktree", "add", "-qb", "feat/wa", str(wa)); git(r, "worktree", "add", "-qb", "feat/wb", str(wb)); n = S(0)
def t_(w, t): mcp(w, [("mem_save", dict(title=t, content="r23 "+t, topic_key="k", type="discovery", project="m4"))])
th = [threading.Thread(target=t_, args=(wa, "v4 a")), threading.Thread(target=t_, args=(wb, "v4 b"))]; [x.start() for x in th]; [x.join() for x in th]
dump(n, "V4 worktrees (misma clave k)")
# V6 vias narrativas
r = repo("m6"); git(r, "switch", "-qc", "feat/narr"); n = S(0)
res, _ = mcp(r, [("mem_session_summary", dict(content="## Goal\nr23 resumen\n## Accomplished\nnada", project="m6", session_id="manual-save-m6")),
                 ("mem_capture_passive", dict(content="## Key Learnings:\n1. r23 aprendizaje pasivo uno\n2. r23 aprendizaje pasivo dos", project="m6", session_id="manual-save-m6")),
                 ("mem_save_prompt", dict(content="r23 prompt guardado", project="m6", session_id="manual-save-m6"))])
dump(n, "V6 narrativas")
c = sqlite3.connect(glob.glob(str(DB/"*.db"))[0]); print("  PROMPTS", c.execute("select content from user_prompts where project='m6'").fetchall())
# V7 sin marcador
r = repo("m7", marker=False); git(r, "switch", "-qc", "feat/sin-marcador"); n = S(0)
mcp(r, [("mem_save", dict(title="v7 sin marcador", content="r23 v7", topic_key="k", type="discovery", project="m7"))]); dump(n, "V7 sin marcador")
# V8 sin active-stack en PATH
r = repo("m8"); git(r, "switch", "-qc", "feat/nopath"); n = S(0)
try:
    out = mcp(r, [("mem_save", dict(title="v8", content="r23 v8", topic_key="k", type="discovery", project="m8"))], nopath=True)
    print("  V8 salida", out)
except Exception as e:
    print("  V8 el proxy NO arranca:", type(e).__name__, str(e)[:120])
dump(n, "V8 (PATH sin active-stack)")
# V9 marcador invalido
r = repo("m9"); git(r, "switch", "-qc", "feat/m"); (r/".engram"/"active-stack.json").write_text("{no es json"); n = S(0)
res, err = mcp(r, [("mem_save", dict(title="v9 marcador malo", content="r23 v9", topic_key="k", type="discovery", project="m9"))]); print("  stderr:", err); dump(n, "V9 marcador invalido")
# V11
r = repo("m11"); n = S(0)
for lab, setup, k in (("una", lambda: git(r, "switch", "-qc", "feat/una"), "k"), ("guion", lambda: (git(r, "switch", "-q", "main"), git(r, "switch", "-qc", "feat-una")), "k"),
                      ("detached", lambda: (git(r, "switch", "-q", "main"), git(r, "checkout", "-q", "--detach")), "kd"),
                      ("largo", lambda: (git(r, "switch", "-q", "main"), git(r, "switch", "-qc", "feature/"+"x"*60+"-final")), "kl")):
    setup(); mcp(r, [("mem_save", dict(title="v11 "+lab, content="r23 "+lab, topic_key=k, type="discovery", project="m11"))])
dump(n, "V11")
# V13 / lectura: hooks de lectura
r = repo("m13"); git(r, "switch", "-qc", "feature/x"); n = S(0)
mcp(r, [("mem_save", dict(title="decision puerto", content="r23 usamos el puerto 4381", topic_key="decision", type="decision", project="m13"))], p="A")
git(r, "switch", "-q", "main")
print("  HOOK session-start en main (B):", hook(r, "session-start", "B"))
print("  HOOK pre-read en main (B):", hook(r, "pre-read", "B"))
git(r, "switch", "-qc", "feature/y")
print("  HOOK session-start en feature/y (B):", hook(r, "session-start", "B"))
res, _ = mcp(r, [("mem_search", dict(query="puerto", project="m13"))], p="B"); print("  SEARCH B:", json.dumps(res)[:600])
dump(n, "V13 filas")
