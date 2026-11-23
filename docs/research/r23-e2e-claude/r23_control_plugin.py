import sys,os
sys.argv=["x","NONE"]
src=open(os.environ["RUN"],encoding="utf-8").read().split("esc = sys.argv[1]")[0]
exec(compile(src,"lib","exec"))
r=repo("ctl3"); git(r,"switch","-qc","feat/ctl")
print(claude("A",r,"Sin explicar: "+g("control sin active-stack")+". Respondé OK."))
for x in rows()[-1:]: print(x["titulo"],x["clave"])
