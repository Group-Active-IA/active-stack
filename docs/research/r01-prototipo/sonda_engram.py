"""R-01: mediciones locales con Git y Engram MCP. Ejecutar con Python 3."""

import json
import os
import shutil
import subprocess
from pathlib import Path
from tempfile import TemporaryDirectory


PROJECT = "r01-synthetic"
CONFIG = Path(".engram/config.json")
WRITE_CONFIG = json.dumps({"project_name": PROJECT}) + "\n"


def show(label, result):
    print(label, json.dumps(result, ensure_ascii=False), flush=True)


def call_mcp(repo, env, method, params):
    """Usa una sesión MCP stdio nueva. La base persiste entre llamadas."""
    with subprocess.Popen(
        ["engram", "mcp", "--tools=all"],
        stdin=subprocess.PIPE, stdout=subprocess.PIPE,
        text=True, cwd=repo, env=env,
    ) as server:
        def send(message):
            server.stdin.write(json.dumps({"jsonrpc": "2.0", **message}) + "\n")
            server.stdin.flush()

        def reply(request_id):
            for line in server.stdout:
                message = json.loads(line)
                if message.get("id") == request_id:
                    if "error" in message:
                        raise RuntimeError(message["error"])
                    return message["result"]
            raise RuntimeError("Engram terminó antes de responder")

        try:
            send({
                "id": 1, "method": "initialize",
                "params": {
                    "protocolVersion": "2024-11-05",
                    "clientInfo": {"name": "r01-pilot", "version": "1"},
                    "capabilities": {},
                },
            })
            reply(1)
            send({"method": "notifications/initialized"})
            send({"id": 2, "method": method, "params": params})
            return reply(2)
        finally:
            server.stdin.close()
            try:
                server.wait(timeout=5)
            except subprocess.TimeoutExpired:
                server.kill()


def tool(repo, env, name, **arguments):
    result = call_mcp(repo, env, "tools/call", {"name": name, "arguments": arguments})
    text = result["content"][0]["text"]
    if result.get("isError"):
        raise RuntimeError(f"{name}: {text}")
    return json.loads(text)


def examine_projects(base, env):
    repo = base / "checkout-original"
    clone = base / "different-clone"
    worktree = base / "different-worktree"

    def git(*args):
        subprocess.run(["git", *args], cwd=base, env=env, check=True)

    git("init", "-q", "-b", "main", repo)
    (repo / CONFIG).parent.mkdir()
    (repo / CONFIG).write_text(WRITE_CONFIG)
    git("-C", repo, "add", CONFIG)
    git("-C", repo, "commit", "-qm", "Synthetic pilot fixture")
    git("clone", "-q", repo, clone)
    git("-C", repo, "worktree", "add", "-qb", "pilot", worktree)

    for configured in (True, False):
        for directory in (repo, clone, worktree):
            if not configured:
                (directory / CONFIG).unlink()
            result = tool(directory, env, "mem_current_project")
            show(f"Q10 config={configured} {directory.name}",
                 [result["project"], result["project_source"]])
    (repo / CONFIG).write_text(WRITE_CONFIG)
    return repo


def examine_observations(repo, env):
    def call(name, **arguments):
        return tool(repo, env, name, **arguments)

    inventory = call_mcp(repo, env, "tools/list", {})["tools"]
    show("Q7/Q8 nombres MCP", [entry["name"] for entry in inventory])
    inspect = {"mem_save", "mem_update", "mem_search", "mem_capture_passive",
               "mem_session_summary", "mem_session_finalize", "mem_save_prompt"}
    for entry in inventory:
        if entry["name"] in inspect:
            show("Campos de " + entry["name"], sorted(entry["inputSchema"]["properties"]))

    saved = []
    branch_key = "branch/pilotfixture/K"
    observations = [
        ("K", "base", "alpha"),
        (branch_key, "[pilotfixture] branch", "beta"),
        (branch_key, "[pilotfixture] revision", "beta"),
    ]
    for index, (key, title, org) in enumerate(observations):
        result = call("mem_save", title=title, topic_key=key, org=org,
                      type="discovery", content=f"synthetic indexing canary {index}")
        saved.append(result["id"])
    show("Q2 IDs al guardar", saved)

    def search_ids(**arguments):
        return [row["id"] for row in call("mem_search", **arguments).get("results", [])]

    show("Q2 búsqueda literal [pilotfixture]", search_ids(query="[pilotfixture]"))
    for org in ("alpha", "beta", "gamma"):
        show(f"Q5 búsqueda canary con org={org}", search_ids(query="canary", org=org))

    base_id, branch_id = saved[:2]
    result = call("mem_update", id=branch_id, topic_key="K", expected_project=PROJECT)
    show("Q4 cambio a clave existente", result["result"])
    for number in (base_id, branch_id):
        show("Q4 lectura tras cambio", call("mem_get_observation", id=number)["result"])
    result = call("mem_save", title="post-collision", topic_key="K",
                  content="synthetic next save", type="discovery")
    show("Q4 ID del guardado posterior", result["id"])


def main():
    for program in ("engram", "git"):
        if not shutil.which(program):
            raise SystemExit(f"Falta {program} en PATH")
    with TemporaryDirectory(prefix="r01-engram-") as directory:
        base = Path(directory).resolve()
        profile = base / "profile"
        profile.mkdir()
        env = {
            "PATH": os.environ["PATH"],
            "HOME": str(profile),
            "USERPROFILE": str(profile),
            "ENGRAM_DATA_DIR": str(base / "db"),
            "ENGRAM_CLOUD_AUTOSYNC": "0",
            "XDG_CONFIG_HOME": str(profile / "config"),
            "GIT_CONFIG_GLOBAL": os.devnull,
            "GIT_CONFIG_NOSYSTEM": "1",
            "GIT_AUTHOR_NAME": "Pilot",
            "GIT_COMMITTER_NAME": "Pilot",
            "GIT_AUTHOR_EMAIL": "pilot@example.invalid",
            "GIT_COMMITTER_EMAIL": "pilot@example.invalid",
        }
        if "SystemRoot" in os.environ:
            env["SystemRoot"] = os.environ["SystemRoot"]
        subprocess.run(["engram", "version"], env=env, cwd=base, check=True)
        subprocess.run(["git", "--version"], env=env, cwd=base, check=True)
        examine_observations(examine_projects(base, env), env)


if __name__ == "__main__":
    main()
