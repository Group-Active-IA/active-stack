import sqlite3,sys,glob
f=glob.glob(sys.argv[1]+"/db/*.db")[0]; c=sqlite3.connect(f)
print([r for r in c.execute("select id,session_id,type,title from observations")])
