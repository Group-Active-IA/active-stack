import subprocess, json, sys, os
ARMS = sys.argv[5].split(",")
PROMPT = sys.argv[1]; tag = sys.argv[2]; N = int(sys.argv[3]); model = sys.argv[4]
repos = {"equipo": "C:/r24b/p", "sin": "C:/r24b/pn2"}
os.makedirs("C:/r24b/out", exist_ok=True)
for i in range(1, N+1):
    for arm in ARMS:
        out = f"C:/r24b/out/{tag}-{arm}-{i}.json"
        subprocess.run(["C:/Program Files/Git/bin/bash.exe", "C:/r24b/cl.sh", repos[arm], out, PROMPT, "--model", model, "--output-format", "stream-json", "--verbose",
            "--allowedTools", "Bash Read Skill Glob Grep", "--disallowedTools", "Bash(git push:*)", "--setting-sources", "user,project",
            "--max-budget-usd", "0.80", "--no-session-persistence"])
        print("hecho", arm, i, flush=True)
