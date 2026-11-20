"""Mide [EN] Q1 y Q13 con dos perfiles Engram temporales en macOS/Linux.

Control sin red: python3 sonda_cloud.py --offline
Piloto: python3 sonda_cloud.py --server URL_DE_LABORATORIO --project PROYECTO_DE_PRUEBA
Lee un objeto JSON por stdin. Usar un servidor de pruebas propio o autorizado.

Un gestor de secretos debe entregar el objeto con los campos "a" y "b":
"a" identifica al emisor, "b" al receptor. Cada valor es un token distinto
con acceso al proyecto de prueba. Para invertir la dirección, invertir los
valores antes de iniciar la sonda. Cada ejecución crea dos observaciones
sintéticas nuevas; la salida incluye solo el marcador y evidencia de la prueba.
"""
import argparse
from contextlib import ExitStack, contextmanager
import json
import os
from pathlib import Path
import select
import subprocess
import sys
import tempfile
import time
import uuid

SCOPES = ("project", "personal")
AUTHOR_FIELDS = {
    "author", "user", "username", "principal", "principal_id", "created_by",
    "actor", "actor_id",
}


def emit(**fields):
    print(json.dumps(fields), flush=True)


class Client:
    def __init__(self, proc):
        self.proc = proc
        self.seq = 0

    def send(self, msg):
        self.proc.stdin.write(json.dumps({"jsonrpc": "2.0", **msg}) + "\n")
        self.proc.stdin.flush()

    def rpc(self, method, params):
        self.seq += 1
        self.send({"method": method, "params": params, "id": self.seq})
        deadline = time.monotonic() + 20
        while time.monotonic() < deadline:
            timeout = max(0, deadline - time.monotonic())
            readable, _, _ = select.select([self.proc.stdout], [], [], timeout)
            if not readable:
                break
            line = self.proc.stdout.readline()
            if not line:
                raise RuntimeError("MCP terminó antes de responder")
            msg = json.loads(line)
            if msg.get("id") == self.seq:
                if "error" in msg:
                    raise RuntimeError("Error de protocolo MCP")
                return msg["result"]
        raise RuntimeError("Tiempo MCP agotado")

    def tool(self, name, **args):
        result = self.rpc("tools/call", {"name": name, "arguments": args})
        if result.get("isError"):
            raise RuntimeError("Engram rechazó " + name)
        return json.loads(result["content"][0]["text"])


@contextmanager
def client(root, token, server, project):
    live = bool(server)
    root.mkdir()
    repo = root / "repo"
    marker = repo / ".engram"
    marker.mkdir(parents=True)
    (marker / "config.json").write_text(json.dumps({"project_name": project}))
    env = {
        "PATH": os.environ["PATH"], "HOME": str(root),
        "ENGRAM_DATA_DIR": str(root / "store"), "ENGRAM_PROJECT": project,
        "XDG_CONFIG_HOME": str(root / "config"), "ENGRAM_CLOUD_AUTOSYNC": "0",
        "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.devnull,
    }
    if live:
        env.update(ENGRAM_CLOUD_TOKEN=token, ENGRAM_CLOUD_SERVER=server)
    # Enroll only the fresh local profile. Never invoke sync/import manually.
    run = subprocess.run(["engram", "cloud", "enroll", project], cwd=repo, env=env,
                         stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=15)
    if run.returncode:
        raise RuntimeError("Falló enroll del perfil temporal")
    if live:
        env["ENGRAM_CLOUD_AUTOSYNC"] = "1"
    with subprocess.Popen(["engram", "mcp"], env=env, cwd=repo, stdin=subprocess.PIPE,
                          stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True) as proc:
        try:
            mcp = Client(proc)
            mcp.rpc("initialize", {"protocolVersion":"2024-11-05",
                    "clientInfo":{"name":"r01-sync", "version":"1"}, "capabilities":{}})
            mcp.send({"method":"notifications/initialized"})
            yield mcp
        finally:
            proc.stdin.close()
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()


def search(client, marker, scope, project):
    result = client.tool("mem_search", query=marker, project=project, scope=scope)
    return next((r for r in result.get("results", [])
                 if r.get("title") == marker and r.get("scope") == scope
                 and r.get("project") == project), None)


def verify_observation(client, number, title, scope, project):
    """Verifica el registro por ID y evita volcar la respuesta del servidor."""
    observation = client.tool("mem_get_observation", id=number)
    lines = set(observation.get("result", "").splitlines())
    expected = {
        f"#{number} [discovery] {title}",
        "Synthetic R01 pilot " + title,
        f"Project: {project}",
        f"Scope: {scope}",
    }
    if observation.get("project") != project or not expected.issubset(lines):
        raise RuntimeError("La observación recuperada difiere del marcador de prueba")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--offline", action="store_true")
    parser.add_argument("--server", help="URL del servidor de prueba autorizado")
    parser.add_argument("--project", help="Nombre del proyecto de prueba autorizado")
    args = parser.parse_args()
    if not args.offline and not (args.server and args.project):
        parser.error("--server y --project son obligatorios para el ensayo remoto")
    if args.offline and (args.server or args.project):
        parser.error("--offline utiliza dos perfiles locales sin servidor")
    project = args.project or "r01-laboratorio-local"
    tokens = {} if args.offline else json.load(sys.stdin)
    if not isinstance(tokens, dict):
        raise ValueError("La entrada debe ser un objeto JSON")
    if not args.offline and not all(
        isinstance(tokens.get(k), str) and tokens[k].strip() for k in ("a", "b")
    ):
        raise ValueError("Se necesitan tokens a y b")
    if not args.offline and tokens["a"] == tokens["b"]:
        raise ValueError("Los dos tokens deben ser distintos")
    with tempfile.TemporaryDirectory(prefix="r01-sync-") as dirname, ExitStack() as stack:
        sender, receiver = (
            stack.enter_context(client(Path(dirname) / key, tokens.get(key), args.server, project))
            for key in ("a", "b")
        )
        marker = "r01probe" + uuid.uuid4().hex
        emit(mode="offline" if args.offline else "live", marker=marker, max_wait_s=150)
        started = {}
        for scope in SCOPES:
            title = marker + scope
            if search(receiver, title, scope, project):
                raise RuntimeError("Marcador preexistente")
            result = sender.tool("mem_save", title=title, content="Synthetic R01 pilot " + title,
                                     scope=scope, topic_key="r01/pilot/" + title, type="discovery",
                                     capture_prompt=False)
            started[scope] = time.monotonic()
            if not result.get("id") or not search(sender, title, scope, project):
                raise RuntimeError("Faltó lectura local de control")
            verify_observation(sender, result["id"], title, scope, project)
            emit(local_read=scope, sender_id=result["id"], verified_by_id=True)
        remaining = set(SCOPES)
        while remaining:
            for scope in sorted(remaining):
                match = search(receiver, marker + scope, scope, project)
                if match:
                    elapsed = round(time.monotonic() - started[scope], 2)
                    verify_observation(receiver, match["id"], marker + scope, scope, project)
                    emit(received_scope=scope, receiver_id=match["id"],
                         verified_by_id=True, latency_upper_s=elapsed,
                         search_author_fields=sorted(AUTHOR_FIELDS.intersection(match)))
                    remaining.remove(scope)
            if args.offline:
                if remaining != set(SCOPES):
                    raise RuntimeError("Perfiles offline no aislados")
                emit(offline_verified=True)
                return 0
            if remaining and time.monotonic()-min(started.values()) >= 150:
                emit(timeout_scopes=sorted(remaining))
                return 2
            if remaining:
                time.sleep(2)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (BrokenPipeError, ValueError, KeyError, RuntimeError, subprocess.SubprocessError) as error:
        # Errors from MCP and input parsing can include untrusted data; whitelist our own labels.
        print(f"Pilot failed: {type(error).__name__}", file=sys.stderr)
        if type(error) is RuntimeError:
            print(str(error), file=sys.stderr)
        sys.exit(1)
