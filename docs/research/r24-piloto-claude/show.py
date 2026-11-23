import json,sys
for f in sys.argv[1:]:
    tools=[];txt="";cost=None;u=None
    for l in open(f,encoding="utf-8",errors="replace"):
        try:e=json.loads(l)
        except:continue
        if e.get("type")=="assistant":
            for c in e["message"].get("content",[]):
                if c.get("type")=="tool_use":
                    a=c["input"];tools.append(c["name"]+":"+str(a.get("skill") or a.get("command") or a.get("file_path") or a.get("pattern") or "")[:140].replace("\n"," "))
        if e.get("type")=="result":txt=str(e.get("result"));cost=e.get("total_cost_usd");u=e.get("usage")
    print("#####",f,"cost",cost)
    for t in tools:print("  T:",t)
    sys.stdout.buffer.write(("  R: "+txt+"\n").encode("utf-8"))
