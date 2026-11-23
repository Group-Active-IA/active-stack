import json, glob, sys, statistics as st
tag = sys.argv[1]
rows = {}
for f in sorted(glob.glob(f"C:/r24b/out/{tag}-*.json")):
    arm = f.split("-")[1]; tools = []; txt = ""; u = None; cost = None; turns = None
    for line in open(f, encoding="utf-8"):
        try: e = json.loads(line)
        except ValueError: continue
        if e.get("type") == "assistant":
            for c in e["message"].get("content", []):
                if c.get("type") == "tool_use":
                    a = c["input"]; tools.append(c["name"] + ":" + (a.get("skill") or a.get("command") or a.get("file_path") or a.get("pattern") or "")[:90])
        if e.get("type") == "result":
            txt = str(e.get("result")); u = e.get("usage"); cost = e.get("total_cost_usd"); turns = e.get("num_turns")
    if u:
        tot = u["input_tokens"] + u["cache_creation_input_tokens"] + u["cache_read_input_tokens"] + u["output_tokens"]
    else: tot = None
    rows.setdefault(arm, []).append(dict(f=f.split("/")[-1], tools=tools, txt=txt, u=u, cost=cost, turns=turns, tot=tot))
for arm, rs in rows.items():
    print("=====", arm)
    for r in rs:
        print(r["f"], "turns", r["turns"], "cost", r["cost"], "tot", r["tot"], r["u"] and (r["u"]["input_tokens"], r["u"]["cache_creation_input_tokens"], r["u"]["cache_read_input_tokens"], r["u"]["output_tokens"]))
        print("  tools:", r["tools"])
        print("  txt:", r["txt"][:350].replace("\n", " "))
    for k in ("tot", "cost"):
        v = [r[k] for r in rs if r[k] is not None]
        if len(v) > 1: print(k, "media %.4f sd %.4f min %s max %s n=%d" % (st.mean(v), st.stdev(v), min(v), max(v), len(v)))
