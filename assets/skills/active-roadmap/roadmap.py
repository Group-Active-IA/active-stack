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

DELIVERY_STATUSES = {
    "pending": "Pendiente",
    "developing": "En desarrollo",
    "review": "En revisión",
    "changes_requested": "Cambios solicitados",
    "merged": "Fusionado",
    "archived": "Archivado",
}
PREVIOUS_STAGES = {
    "developing": {"pending", "changes_requested"},
    "review": {"developing"},
    "changes_requested": {"review"},
    "merged": {"review"},
    "archived": {"merged"},
}


def delivery_statuses(names):
    if not isinstance(names, list) or any(not isinstance(name, str) for name in names):
        raise RoadmapError("Ingresá los nombres de estado verificados de la lista")
    by_normalized = {}
    for name in names:
        normalized = name.strip().casefold()
        if normalized in by_normalized:
            raise RoadmapError("La lista contiene nombres de estado ambiguos")
        by_normalized[normalized] = name
    missing = [name for name in DELIVERY_STATUSES.values() if name.casefold() not in by_normalized]
    if missing:
        raise RoadmapError("Faltan estados en ClickUp: " + ", ".join(missing))
    return {stage: by_normalized[name.casefold()] for stage, name in DELIVERY_STATUSES.items()}



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


def delivery_update(task, previous):
    item_id = task["id"]
    branch, pr_url = task.get("branch"), task.get("pr_url")
    branch_pattern = rf"{re.escape(item_id)}-[a-z0-9]+(?:-[a-z0-9]+)*"
    if not isinstance(branch, str) or not re.fullmatch(branch_pattern, branch):
        raise RoadmapError("La rama del PR debe incluir el ID exacto de la ficha")
    if not isinstance(pr_url, str) or not re.fullmatch(
        r"https://[A-Za-z0-9.-]+/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/pull/[1-9][0-9]*", pr_url
    ):
        raise RoadmapError("Falta el enlace confirmado al PR de la ficha")
    delivery = {**previous, "pr_url": pr_url}
    if task["operation"] == "close":
        sha, handoff = task.get("merge_sha"), task.get("handoff")
        if not isinstance(sha, str) or not re.fullmatch(r"[0-9a-f]{40}|[0-9a-f]{64}", sha):
            raise RoadmapError("Falta el SHA verificado del merge")
        if handoff != f"knowledge-base/entrega/{item_id}.md":
            raise RoadmapError("Falta la ruta del traspaso único de la ficha")
        delivery.update(merge_sha=sha, handoff=handoff)
    elif previous.get("merge_sha") and previous.get("pr_url") != pr_url:
        delivery.pop("merge_sha", None)
        delivery.pop("handoff", None)
    return delivery


class Roadmap:
    def __init__(self, root):
        self.root = root
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

    def team(self):
        path = self.config_path.with_name("team.json")
        if not path.exists():
            return None
        settings = load(path)
        tracking = settings.get("tracking", "local")
        if tracking not in {"clickup", "local"}:
            raise RoadmapError("Seguimiento de equipo inválido")
        members = {}
        for member in settings.get("members", []):
            member_id = member["id"]
            if member_id in members:
                raise RoadmapError("Identidad de equipo duplicada")
            members[member_id] = member.get("clickup_id")
        private = path.with_name("team.local.json")
        if settings.get("names_versioned") is False and private.is_file():
            private_ids = set()
            for member in load(private).get("members", []):
                member_id = member["id"]
                if member_id not in members or member_id in private_ids:
                    raise RoadmapError("Identidad privada inválida")
                private_ids.add(member_id)
                members[member_id] = member.get("clickup_id")
        if any(value is not None and (type(value) is not int or value <= 0)
               for value in members.values()):
            raise RoadmapError("ID de ClickUp inválido en el contrato de equipo")
        return {"tracking": tracking, "members": members}

    def journal_path(self):
        try:
            output = subprocess.run(
                ["git", "-C", str(self.root), "rev-parse", "--git-path", "active-roadmap/state.json"],
                check=True, capture_output=True, timeout=10,
            )
        except (OSError, subprocess.SubprocessError) as error:
            raise RoadmapError("No se pudo localizar el diario local de este worktree Git") from error
        return self.root / os.fsdecode(output.stdout.rstrip(b"\r\n"))

    def read(self):
        shared = load(self.state_path) if self.state_path.exists() else {"tasks": {}}
        if not self.team():
            return shared
        journal = self.journal_path()
        if not journal.exists():
            return shared
        local = load(journal)
        shared_destination = shared.get("destination")
        if shared_destination and local.get("destination") != shared_destination:
            raise RoadmapError("El destino del diario local difiere del vínculo compartido")
        if "destination" in local:
            shared.setdefault("destination", local["destination"])
        for task_id, entry in local["tasks"].items():
            shared["tasks"].setdefault(task_id, {}).update(entry)
        return shared

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
        path = self.journal_path() if self.team() else self.state_path
        lock = path.with_suffix(".lock")
        path.parent.mkdir(parents=True, exist_ok=True)
        try:
            lock.mkdir()
        except FileExistsError as error:
            raise RoadmapError("Registro ocupado. Reconciliá el proceso anterior antes de liberar el bloqueo.") from error
        try:
            yield
        finally:
            lock.rmdir()

    def write(self, state):
        if self.team() is None:
            self.write_state(self.state_path, state)
            return
        shared = {"tasks": {}}
        local = {**state, "tasks": {}}
        link_fields = {"remote_id", "url", "fingerprint", "delivery"}
        for task_id, entry in state["tasks"].items():
            link = {key: value for key, value in entry.items() if key in link_fields}
            if link:
                shared["tasks"][task_id] = link
            local["tasks"][task_id] = {key: value for key, value in entry.items() if key not in link_fields}
        if shared["tasks"] and "destination" in state:
            shared["destination"] = state["destination"]
        journal = self.journal_path()
        if self.state_path.exists() or shared["tasks"]:
            old_shared = load(self.state_path) if self.state_path.exists() else None
            if shared != old_shared:
                if old_shared:
                    # Keep legacy observations and pending work recoverable until
                    # the shared update succeeds. A new create already has a
                    # pending local intent from begin().
                    legacy = {
                        task_id: {key: value for key, value in entry.items() if key not in link_fields}
                        for task_id, entry in old_shared["tasks"].items()
                    }
                    if any(legacy.values()):
                        old_local = load(journal) if journal.exists() else {"tasks": {}}
                        before = {**old_shared, **old_local, "tasks": {}}
                        for task_id in legacy.keys() | old_local["tasks"].keys():
                            before["tasks"][task_id] = {
                                **legacy.get(task_id, {}), **old_local["tasks"].get(task_id, {}),
                            }
                        self.write_state(journal, before)
                self.write_state(self.state_path, shared)
        self.write_state(journal, local)

    @staticmethod
    def write_state(path, state):
        with tempfile.NamedTemporaryFile(
            mode="w", dir=path.parent, encoding="utf-8", delete=False
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
            temporary.replace(path)
        finally:
            temporary.unlink(missing_ok=True)

    def preview(self, tasks, *, selected=None, access="ok"):
        return self._preview(tasks, self.config(), self.read(), self.team(), selected, access)

    def _preview(self, tasks, settings, state, team=None, selected=None, access="ok"):
        mode = settings["mode"]
        if (team and team["tracking"] == "local") or access != "ok":
            mode = "local"
        if selected is not None and (not selected or len(set(selected)) != len(selected)):
            raise RoadmapError("Seleccioná IDs únicos de la vista previa")
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
            if selected is not None and task_id not in selected:
                continue
            title = task["title"].strip()
            if not title:
                raise RoadmapError(f"Título vacío: {task_id}")
            current = state["tasks"].get(task_id, {})
            marker = "active-stack-" + digest([destination["project_id"], task_id])
            dependencies = ", ".join(task.get("dependencies", [])) or "Ninguna"
            responsible = task.get("responsible")
            accounts = team["members"] if team else settings.get("assignees", {})
            assignee = accounts.get(responsible) if mode == "clickup" or not team else None
            remote_id = current.get("remote_id")
            assignment_pending = bool(responsible and not assignee)
            if remote_id and not team:
                assignment_pending = current.get("assignment_pending", False)
            delivery = current.get("delivery", {})
            operation = task.get("operation", "publish") if team else "publish"
            if task.get("operation") == "status":
                operation = "status"
            if mode == "clickup" and team and operation in {"link_pr", "close"}:
                delivery = delivery_update(task, delivery)
            delivery_text = "".join(f"{label}: {delivery[key]}\n" for key, label in (
                ("pr_url", "PR"), ("merge_sha", "SHA de merge"), ("handoff", "Traspaso"),
            ) if key in delivery)
            description = (
                f"[{marker}]\n"
                f"Ficha: CHANGES.md · {task_id}\n"
                f"Alcance: {task.get('description', '')}\n"
                f"Dependencias: {dependencies}\n"
                f"Responsable previsto: {responsible or 'Sin definir'}\n"
                f"{delivery_text}[/{marker}]"
            )
            managed_description = description
            content_fingerprint = digest({"name": title, "description": managed_description})
            current_description = task.get("current_description")
            if mode == "local":
                action = "local"
            elif current.get("pending"):
                action = "recover"
            elif operation == "status":
                action = "inspect"
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
            row = {
                "id": task_id, "action": action, "title": title, "description": description,
                "dependencies": task.get("dependencies", []), "responsible": responsible,
                "assignee": assignee if not remote_id else None,
                "assignment_pending": assignment_pending, "remote_id": remote_id,
                "url": current.get("url"), "marker": marker,
                "current_description_hash": digest(current_description) if action == "update" else None,
                "content_fingerprint": content_fingerprint,
            }
            if operation == "status" and mode == "clickup":
                self.status_row(row, task, current, destination)
            elif team and mode == "clickup":
                self.team_row(row, task, current, team, delivery, operation, managed_description)
            rows.append(row)
        if selected is not None:
            by_id = {row["id"]: row for row in rows}
            if any(task_id not in by_id for task_id in selected):
                raise RoadmapError("La selección contiene fichas inexistentes")
            rows = [by_id[task_id] for task_id in selected]
        preview = {"mode": mode, **destination, "rows": rows}
        if access != "ok":
            preview["access"] = access
        return {**preview, "approval": digest(preview)}

    def status_row(self, row, task, current, destination):
        row["operation"] = "status"
        if row["action"] == "recover":
            return
        if (not row["remote_id"] or task.get("current_remote_id") != row["remote_id"]
                or task.get("current_list_id") != destination["list_id"]
                or merge_description(task.get("current_text"), row["marker"], "") is None):
            raise RoadmapError("Confirmá la tarea vinculada, su lista de origen y su marcador")
        stage = task.get("stage")
        if stage not in PREVIOUS_STAGES:
            raise RoadmapError("Hito de entrega inválido")
        statuses = delivery_statuses(task.get("list_statuses"))
        current_status = task.get("current_status")
        if not isinstance(current_status, str) or not current_status.strip():
            raise RoadmapError("Consultá el estado actual de la tarea vinculada")
        normalized = current_status.strip().casefold()
        current_stage = next((key for key, name in statuses.items()
                              if name.strip().casefold() == normalized), None)
        target = statuses[stage]
        if current_stage != stage and current_stage not in PREVIOUS_STAGES[stage]:
            raise RoadmapError(f"Transición no permitida de {current_status} a {target}")
        unchanged = current_stage == stage
        row.update(
            action="unchanged" if unchanged else "update",
            stage=stage,
            current_status=current_status,
            changes={} if unchanged else {"status": target},
            content_fingerprint=current.get("fingerprint"),
        )

    def team_row(self, row, task, current, team, delivery, operation, managed_description):
        if row["action"] == "recover":
            return
        if operation == "publish":
            row["delivery"] = delivery
            return
        if not row["remote_id"]:
            raise RoadmapError("La operación requiere una ficha vinculada")
        if (task.get("current_remote_id") != row["remote_id"]
                or merge_description(task.get("current_text"), row["marker"], "") is None):
            raise RoadmapError("Confirmá la identidad de la tarea vinculada y su marcador")
        if operation in {"take", "assign"}:
            assignees = task.get("current_assignees")
            if not isinstance(assignees, list) or any(
                type(value) is not int or value <= 0 for value in assignees
            ):
                raise RoadmapError("Consultá primero los responsables actuales de ClickUp")
            if operation == "take":
                me_path = self.config_path.with_name("me.local.json")
                member = load(me_path).get("id") if me_path.is_file() else None
            else:
                member = row["responsible"]
            account = team["members"].get(member)
            if account is None and operation == "take":
                raise RoadmapError("Identificate con un integrante que tenga un ID confirmado de ClickUp")
            row.update(
                action="unchanged" if account is None or assignees == [account] else "update",
                current_assignees=assignees,
                content_fingerprint=current.get("fingerprint"),
                changes={"assignees": [account]} if account is not None else {},
                assignment_pending=bool(member and account is None),
            )
        elif operation in {"link_pr", "close"}:
            title = task.get("current_title")
            if not isinstance(title, str) or not title.strip():
                raise RoadmapError("Consultá el título actual de la tarea vinculada")
            current_description = task.get("current_description")
            final_description = merge_description(current_description, row["marker"], managed_description)
            changes = {}
            if final_description is None:
                row["action"] = "inspect"
            elif final_description == current_description:
                row["action"] = "unchanged"
            else:
                row["action"] = "update"
                changes["description"] = final_description
            row["current_description_hash"] = digest(current_description)
            row["content_fingerprint"] = digest({"name": title, "description": managed_description})
            if operation == "close":
                status = task.get("current_status")
                closing = task.get("closing_status")
                if any(not isinstance(value, str) or not value.strip() for value in (status, closing)):
                    raise RoadmapError("Consultá el estado actual y el estado de cierre disponible")
                if row["action"] != "inspect" and status != closing:
                    row["action"] = "update"
                    changes["status"] = closing
                row["current_status"] = status
            row.update(changes=changes, delivery=delivery)
        else:
            raise RoadmapError("Operación de equipo inválida")
        row["operation"] = operation

    def begin(self, tasks, approval, selected):
        if not selected or len(set(selected)) != len(selected):
            raise RoadmapError("Seleccioná IDs únicos de la vista previa")
        with self.locked():
            state = self.read()
            team = self.team()
            plan = self._preview(tasks, self.config(), state, team, selected if team else None)
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

    def result(self, task_id, outcome, *, remote_id=None, url=None, reason=None, status=None):
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
                if pending.get("operation") == "status" and (
                    not isinstance(status, str)
                    or status.strip().casefold() != pending["changes"]["status"].strip().casefold()
                ):
                    raise RoadmapError("Confirmá el estado objetivo antes de registrar la escritura")
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
                if pending.get("delivery"):
                    entry["delivery"] = pending["delivery"]
                if "assignees" in pending.get("changes", {}):
                    entry["assignment_pending"] = False
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
    preview.add_argument("--select", nargs="+")
    preview.add_argument("--access", choices=["ok", "unauthorized", "forbidden", "not_found", "error"], default="ok")
    begin = commands.add_parser("begin")
    begin.add_argument("tasks", type=Path)
    begin.add_argument("approval")
    begin.add_argument("ids", nargs="+")
    for name in ("result", "reconcile"):
        command = commands.add_parser(name)
        command.add_argument("task_id")
        command.add_argument("response", type=Path)
    statuses = commands.add_parser("statuses")
    statuses.add_argument("snapshot", type=Path)
    arguments = parser.parse_args()
    journal = Roadmap(arguments.root)
    try:
        if arguments.command == "statuses":
            output = delivery_statuses(load(arguments.snapshot))
        elif arguments.command == "read":
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
            output = journal.preview(
                load(arguments.tasks), selected=arguments.select, access=arguments.access,
            )
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
