"""Local preview and publication journal for Active Roadmap."""

import argparse
from contextlib import contextmanager
from datetime import datetime, timezone
from hashlib import sha256
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import uuid


class RoadmapError(ValueError):
    pass


def digest(value):
    return sha256(json.dumps(value, sort_keys=True, ensure_ascii=False).encode("utf-8")).hexdigest()


def timestamp():
    return datetime.now(timezone.utc).isoformat(timespec="seconds")


def load(path):
    with path.open(encoding="utf-8-sig") as document:
        return json.load(document)


CLICKUP_MCP_URL = "https://mcp.clickup.com/mcp"


def diagnose_clickup(registrations, probe=None, name=None):
    official = [
        server for server in registrations
        if server["transport"]["type"] == "streamable_http"
        and server["transport"].get("url", "").rstrip("/") == CLICKUP_MCP_URL
    ]
    if name is not None:
        official = [server for server in official if server["name"] == name]
        if not official:
            return {"status": "missing", "name": name}
    enabled = [server for server in official if server["enabled"]]
    candidates = enabled or official
    reachable = [server for server in candidates if server.get("reachable", True)]
    candidates = reachable or candidates
    if not candidates:
        status = "conflict" if any(server["name"] == "clickup" for server in registrations) else "missing"
        return {"status": status}
    server = next((item for item in candidates if item["name"] == "clickup"), None)
    if server is None:
        if len(candidates) > 1:
            return {"status": "ambiguous", "names": [item["name"] for item in candidates]}
        server = candidates[0]
    if not server["enabled"]:
        status = "disabled"
    elif server.get("reachable") is False and probe is None:
        status = "verification_pending"
    elif server["auth_status"] == "not_logged_in" or probe == "unauthorized":
        status = "login_required"
    else:
        status = {
            None: "check_list",
            "forbidden": "list_unavailable",
            "not_found": "list_unavailable",
            "ok": "ready",
            "error": "verification_pending",
        }[probe]
    return {"status": status, "name": server["name"]}


def client_command(name, *arguments):
    """Resolve npm shims such as codex.cmd or claude.cmd, which Windows needs by full path."""
    return [shutil.which(name) or name, *arguments]


def codex_registrations():
    try:
        process = subprocess.run(client_command("codex", "mcp", "list", "--json"), capture_output=True, check=True, timeout=30)
        return json.loads(process.stdout)
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        raise RoadmapError("No se pudo consultar la configuración MCP de Codex") from error


CLAUDE_MCP_LINE = re.compile(r"^(?P<name>[^\s].*?): (?P<target>.+?)(?: \((?P<kind>HTTP|SSE)\))? - (?P<health>[^-]+)$")


def parse_claude_mcp_list(text):
    """Translate `claude mcp list` output into the registration shape used by Codex."""
    registrations = []
    for line in text.splitlines():
        match = CLAUDE_MCP_LINE.match(line.strip())
        if not match:
            continue
        target = match["target"]
        health = match["health"].lower()
        if not target.startswith(("http://", "https://")):
            transport = {"type": "stdio", "command": target}
        elif match["kind"] == "SSE":
            transport = {"type": "sse", "url": target}
        else:
            transport = {"type": "streamable_http", "url": target}
        registrations.append({
            "name": match["name"],
            "enabled": True,
            "auth_status": "not_logged_in" if "auth" in health else "oauth",
            "reachable": "connected" in health or "auth" in health,
            "transport": transport,
        })
    return registrations


def claude_registrations():
    try:
        process = subprocess.run(client_command("claude", "mcp", "list"), capture_output=True, check=True, timeout=120)
        return parse_claude_mcp_list(process.stdout.decode("utf-8", errors="replace"))
    except (OSError, subprocess.SubprocessError) as error:
        raise RoadmapError("No se pudo consultar la configuración MCP de Claude Code") from error


def merge_description(current, marker, replacement):
    if not isinstance(current, str):
        return None
    pattern = rf"(\\?)\[(/?){re.escape(marker)}(\\?)\]"
    matches = list(re.finditer(pattern, current))
    if len(matches) != 2:
        return None
    opening, closing = matches
    style = opening[1]
    if opening[2] or closing[2] != "/" or (opening[3], closing[1], closing[3]) != (style, style, style):
        return None
    if style:
        replacement = replacement.replace(f"[{marker}]", opening[0], 1)
        replacement = replacement.replace(f"[/{marker}]", closing[0], 1)
    return current[:opening.start()] + replacement + current[closing.end():]


class Roadmap:
    def __init__(self, root):
        self.config_path = root / ".active-stack" / "roadmap.json"
        self.state_path = root / ".active-stack" / "roadmap-state.json"

    def config(self):
        settings = load(self.config_path) if self.config_path.exists() else {"mode": "local"}
        if settings["mode"] not in {"local", "clickup"}:
            raise RoadmapError("Modo de roadmap inválido")
        if settings["mode"] == "clickup":
            for field in ("project_id", "list_id"):
                if not isinstance(settings.get(field), str) or not settings[field].strip():
                    raise RoadmapError(f"Falta configurar {field}")
        return settings

    def read(self):
        return load(self.state_path) if self.state_path.exists() else {"tasks": {}}

    def configure(self, list_id, workspace_id=None):
        """Save the approved ClickUp destination, keeping a stable project_id and known assignees."""
        if not isinstance(list_id, str) or not list_id.strip():
            raise RoadmapError("Falta configurar list_id")
        current = load(self.config_path) if self.config_path.exists() else {}
        settings = {
            "mode": "clickup",
            "project_id": current.get("project_id") or str(uuid.uuid4()),
            "list_id": list_id.strip(),
        }
        workspace = workspace_id or current.get("workspace_id")
        if workspace:
            settings["workspace_id"] = workspace
        if current.get("assignees"):
            settings["assignees"] = current["assignees"]
        self.config_path.parent.mkdir(parents=True, exist_ok=True)
        if self.config_path.exists():
            backup = self.config_path.with_suffix(".json.bak")
            backup.write_bytes(self.config_path.read_bytes())
        with tempfile.NamedTemporaryFile(
            mode="w", dir=self.config_path.parent, encoding="utf-8", delete=False
        ) as output:
            json.dump(settings, output, ensure_ascii=False, indent=2)
            output.write("\n")
        Path(output.name).replace(self.config_path)
        return settings

    @contextmanager
    def locked(self):
        lock = self.state_path.with_suffix(".lock")
        self.state_path.parent.mkdir(parents=True, exist_ok=True)
        try:
            lock.mkdir()
        except FileExistsError as error:
            raise RoadmapError("Registro ocupado. Reconciliá el proceso anterior antes de liberar el bloqueo.") from error
        try:
            yield
        finally:
            lock.rmdir()

    def write(self, state):
        with tempfile.NamedTemporaryFile(
            mode="w", dir=self.state_path.parent, encoding="utf-8", delete=False
        ) as output:
            temporary = Path(output.name)
            try:
                json.dump(state, output, ensure_ascii=False, indent=2)
                output.write("\n")
                output.flush()
                os.fsync(output.fileno())
            except Exception:
                output.close()
                temporary.unlink(missing_ok=True)
                raise
        try:
            temporary.replace(self.state_path)
        finally:
            temporary.unlink(missing_ok=True)

    def preview(self, tasks):
        return self._preview(tasks, self.config(), self.read())

    def _preview(self, tasks, settings, state):
        mode = settings["mode"]
        destination = {field: settings.get(field) for field in ("project_id", "workspace_id", "list_id")}
        if mode == "clickup" and state.get("destination", destination) != destination:
            raise RoadmapError("El destino cambió. Reconciliá los vínculos existentes antes de publicar.")
        rows = []
        seen = set()
        for task in tasks:
            task_id = task["id"]
            if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*", task_id) or task_id in seen:
                raise RoadmapError(f"ID de tarea inválido o repetido: {task_id}")
            seen.add(task_id)
            title = task["title"].strip()
            if not title:
                raise RoadmapError(f"Título vacío: {task_id}")
            current = state["tasks"].get(task_id, {})
            marker = "active-stack-" + digest([destination["project_id"], task_id])
            dependencies = ", ".join(task.get("dependencies", [])) or "Ninguna"
            responsible = task.get("responsible")
            assignee = settings.get("assignees", {}).get(responsible)
            remote_id = current.get("remote_id")
            assignment_pending = current.get("assignment_pending", False) if remote_id else bool(responsible and not assignee)
            description = (
                f"[{marker}]\n"
                f"Ficha: CHANGES.md · {task_id}\n"
                f"Alcance: {task.get('description', '')}\n"
                f"Dependencias: {dependencies}\n"
                f"Responsable previsto: {responsible or 'Sin definir'}\n"
                f"[/{marker}]"
            )
            content_fingerprint = digest({"name": title, "description": description})
            current_description = task.get("current_description")
            if mode == "local":
                action = "local"
            elif current.get("pending"):
                action = "recover"
            elif remote_id and current.get("fingerprint") == content_fingerprint:
                action = "unchanged"
            elif remote_id:
                merged = merge_description(current_description, marker, description)
                action = "inspect"
                if merged is not None:
                    action = "update"
                    description = merged
            else:
                action = "create"
            rows.append({
                "id": task_id, "action": action, "title": title, "description": description,
                "dependencies": task.get("dependencies", []), "responsible": responsible,
                "assignee": assignee if not remote_id else None,
                "assignment_pending": assignment_pending, "remote_id": remote_id,
                "url": current.get("url"), "marker": marker,
                "current_description_hash": digest(current_description) if action == "update" else None,
                "content_fingerprint": content_fingerprint,
            })
        preview = {"mode": mode, **destination, "rows": rows}
        return {**preview, "approval": digest(preview)}

    def begin(self, tasks, approval, selected):
        if not selected or len(set(selected)) != len(selected):
            raise RoadmapError("Seleccioná IDs únicos de la vista previa")
        with self.locked():
            state = self.read()
            plan = self._preview(tasks, self.config(), state)
            if approval != plan["approval"]:
                raise RoadmapError("La vista previa cambió. Mostrá el nuevo plan y pedí confirmación.")
            rows = {row["id"]: row for row in plan["rows"]}
            if any(task_id not in rows or rows[task_id]["action"] not in {"create", "update"} for task_id in selected):
                raise RoadmapError("La selección contiene tareas sin una escritura autorizable")
            state["destination"] = {key: plan[key] for key in ("project_id", "workspace_id", "list_id")}
            for task_id in selected:
                entry = state["tasks"].setdefault(task_id, {})
                entry["pending"] = {**rows[task_id], "started_at": timestamp()}
            self.write(state)
            return {**state["destination"], "rows": [rows[task_id] for task_id in selected]}

    def result(self, task_id, outcome, *, remote_id=None, url=None, reason=None):
        with self.locked():
            state = self.read()
            entry = state["tasks"].get(task_id, {})
            pending = entry.get("pending")
            if not pending:
                raise RoadmapError(f"No hay escritura pendiente para {task_id}")
            if outcome in {"created", "updated"}:
                expected = "created" if pending["action"] == "create" else "updated"
                if outcome != expected or not remote_id or not url:
                    raise RoadmapError("El resultado exitoso requiere operación, ID y enlace coherentes")
                if entry.get("remote_id", remote_id) != remote_id:
                    raise RoadmapError("La respuesta cambió el ID remoto de la tarea")
                entry.update({
                    "remote_id": remote_id, "url": url,
                    "fingerprint": pending["content_fingerprint"],
                    "assignment_pending": (
                        pending["assignment_pending"] if outcome == "created"
                        else entry.get("assignment_pending", False)
                    ),
                })
            elif outcome not in {"failed", "uncertain", "skipped"} or not reason:
                raise RoadmapError("El resultado requiere un motivo de fallo, incertidumbre u omisión")
            if outcome != "uncertain":
                entry.pop("pending")
            entry["result"] = {"outcome": outcome, "reason": reason, "at": timestamp()}
            self.write(state)
            return entry

    def reconcile(self, task_id, *, status=None, assignees=None, reason=None):
        with self.locked():
            state = self.read()
            entry = state["tasks"].get(task_id, {})
            if not entry.get("remote_id"):
                raise RoadmapError(f"No hay vínculo a ClickUp para {task_id}")
            if reason:
                entry["verification_pending"] = True
                entry["verification_error"] = reason
            else:
                if not isinstance(status, str) or not status.strip() or not isinstance(assignees, list):
                    raise RoadmapError("La conciliación requiere estado y lista de responsables")
                entry["observed"] = {"status": status, "assignees": assignees, "at": timestamp()}
                entry["verification_pending"] = False
                entry.pop("verification_error", None)
                if assignees:
                    entry["assignment_pending"] = False
            self.write(state)
            return entry


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, required=True)
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("read")
    setup = commands.add_parser("setup")
    setup.add_argument("--probe", choices=["unauthorized", "forbidden", "not_found", "ok", "error"])
    setup.add_argument("--client", choices=["codex", "claude"], required=True)
    setup.add_argument("--name", help="MCP connection to use when several point to ClickUp")
    setup.add_argument("--list-id", help="Destination to diagnose before saving the configuration")
    setup.add_argument("--workspace-id")
    configure = commands.add_parser("configure")
    configure.add_argument("--list-id", required=True)
    configure.add_argument("--workspace-id")
    preview = commands.add_parser("preview")
    preview.add_argument("tasks", type=Path)
    begin = commands.add_parser("begin")
    begin.add_argument("tasks", type=Path)
    begin.add_argument("approval")
    begin.add_argument("ids", nargs="+")
    for name in ("result", "reconcile"):
        command = commands.add_parser(name)
        command.add_argument("task_id")
        command.add_argument("response", type=Path)
    arguments = parser.parse_args()
    journal = Roadmap(arguments.root)
    try:
        if arguments.command == "read":
            output = journal.read()
        elif arguments.command == "setup":
            settings = journal.config()
            if arguments.list_id:
                settings = {**settings, "mode": "clickup", "list_id": arguments.list_id,
                            "workspace_id": arguments.workspace_id or settings.get("workspace_id")}
            if settings["mode"] == "local":
                output = {"status": "local"}
            else:
                output = {
                    **diagnose_clickup(
                        claude_registrations() if arguments.client == "claude" else codex_registrations(),
                        arguments.probe,
                        arguments.name,
                    ),
                    "workspace_id": settings.get("workspace_id"),
                    "list_id": settings["list_id"],
                }
        elif arguments.command == "configure":
            output = journal.configure(arguments.list_id, arguments.workspace_id)
        elif arguments.command == "preview":
            output = journal.preview(load(arguments.tasks))
        elif arguments.command == "begin":
            output = journal.begin(load(arguments.tasks), arguments.approval, arguments.ids)
        else:
            operation = journal.result if arguments.command == "result" else journal.reconcile
            output = operation(arguments.task_id, **load(arguments.response))
    except (OSError, ValueError, KeyError, TypeError) as error:
        parser.exit(1, f"{error}\n")
    print(json.dumps(output, ensure_ascii=True, indent=2))


if __name__ == "__main__":
    main()
