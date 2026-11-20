"""Validador del contrato del modo equipo (.active-stack/team.json y fichas).

Solo biblioteca estandar, Python 3.8+. Uso:
    team.py --root <proyecto> validate [--file <team.json>]
    team.py --root <proyecto> ficha --file <ficha.json>
Imprime siempre un JSON por stdout. Codigos: 0 ok, 1 invalido/missing, 2 uso.
"""
import argparse
import json
import os
import re
import subprocess
import sys

ROOT_FIELDS = {"version", "size", "tracking", "members", "names_versioned"}
MEMBER_FIELDS = {"id", "name", "github", "clickup_id"}
FICHA_FIELDS = ("Resultado", "Depende de", "Zona", "Requiere", "Modo", "Responsable")

SUPPORTED_VERSION = 1
TRACKING_VALUES = ("local", "clickup")
MODO_VALUES = ("odd", "sdd")
LEGACY_DEPENDENCIES = "Dependencias"
TEAM_MARKER = "<!-- active-stack:team-mode -->"

TEAM_PATH = ".active-stack/team.json"
TEAM_LOCAL_PATH = ".active-stack/team.local.json"
ME_LOCAL_PATH = ".active-stack/me.local.json"

MEMBER_ID_RE = re.compile(r"^[a-z0-9][a-z0-9-]{0,31}$")
GITHUB_RE = re.compile(r"^[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})$")
FICHA_ID_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._-]*$")
DEP_ISSUE_RE = re.compile(r"^#[0-9]+(?:@[0-9a-fA-F]{7,40})?$")
NO_DEPENDENCIES = ("ninguna", "—")
CREDENTIAL_RE = re.compile(
    r"(?<![A-Za-z0-9])(?:gh[pousr]_|github_pat_|pk_[0-9]+_|sk-|xox[abp]-|Bearer\s)"
)
GIT_TIMEOUT = 5


# ---------------------------------------------------------------- utilidades

def _is_int(value):
    return isinstance(value, int) and not isinstance(value, bool)


def _type_name(value):
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "booleano"
    if isinstance(value, int):
        return "entero"
    if isinstance(value, float):
        return "numero decimal"
    if isinstance(value, str):
        return "texto"
    if isinstance(value, list):
        return "lista"
    if isinstance(value, dict):
        return "objeto"
    return type(value).__name__


def _credential_errors(value, path):
    """Recorre cualquier valor y reporta los textos que parecen credenciales."""
    errors = []
    if isinstance(value, str):
        if CREDENTIAL_RE.search(value):
            errors.append("%s: parece una credencial; los tokens no se guardan en estos archivos" % path)
    elif isinstance(value, dict):
        for key, item in value.items():
            errors.extend(_credential_errors(item, "%s.%s" % (path, key) if path else str(key)))
    elif isinstance(value, list):
        for index, item in enumerate(value):
            errors.extend(_credential_errors(item, "%s[%d]" % (path, index)))
    return errors


def _read_json(path):
    """Devuelve (doc, error). error es None si se pudo leer."""
    try:
        with open(path, "r", encoding="utf-8-sig") as handle:
            text = handle.read()
    except OSError as exc:
        return None, "%s: no se pudo leer (%s)" % (path, exc.strerror or exc)
    try:
        return json.loads(text), None
    except ValueError as exc:
        return None, "%s: JSON invalido (%s)" % (path, exc)


# ---------------------------------------------------------------- team.json

def _validate_member(member, path, tracking, shared):
    errors = []
    if not isinstance(member, dict):
        return ["%s: debe ser un objeto, no %s" % (path, _type_name(member))]
    for key in sorted(set(member) - MEMBER_FIELDS):
        errors.append("%s.%s: campo desconocido (permitidos: %s)"
                      % (path, key, ", ".join(sorted(MEMBER_FIELDS))))
    if "id" not in member:
        errors.append("%s.id: es obligatorio" % path)
    else:
        member_id = member["id"]
        if not isinstance(member_id, str) or not MEMBER_ID_RE.match(member_id):
            errors.append("%s.id: %r no es valido; usa minusculas, digitos y guiones (hasta 32, p. ej. m1)"
                          % (path, member_id))
    if "name" in member and member["name"] is not None and not isinstance(member["name"], str):
        errors.append("%s.name: debe ser texto o null, no %s" % (path, _type_name(member["name"])))
    if "github" in member and member["github"] is not None:
        handle = member["github"]
        if not isinstance(handle, str) or not GITHUB_RE.match(handle):
            errors.append("%s.github: %r no es un usuario de GitHub valido" % (path, handle))
    if "clickup_id" in member and member["clickup_id"] is not None:
        if tracking != "clickup":
            errors.append("%s.clickup_id: solo se usa con tracking \"clickup\"" % path)
        elif not _is_int(member["clickup_id"]):
            errors.append("%s.clickup_id: debe ser un entero, no %s"
                          % (path, _type_name(member["clickup_id"])))
    if shared is not None and shared is False:
        for key in ("name", "github", "clickup_id"):
            if member.get(key) is not None:
                errors.append("%s.%s: con names_versioned false los datos personales van en %s"
                              % (path, key, TEAM_LOCAL_PATH))
    return errors


def validate_team(doc):
    """Valida el contenido de team.json. Devuelve la lista de errores."""
    if not isinstance(doc, dict):
        return ["team.json: la raiz debe ser un objeto JSON, no %s" % _type_name(doc)]
    errors = []

    if "version" in doc:
        version = doc["version"]
        if not _is_int(version):
            errors.append("version: debe ser el entero %d, no %s" % (SUPPORTED_VERSION, _type_name(version)))
        elif version != SUPPORTED_VERSION:
            errors.append("version: %d no esta soportada; solo se soporta la version %d"
                          % (version, SUPPORTED_VERSION))

    for key in sorted(set(doc) - ROOT_FIELDS):
        errors.append("%s: campo desconocido (permitidos: %s)" % (key, ", ".join(sorted(ROOT_FIELDS))))

    size = doc.get("size")
    if "size" not in doc:
        errors.append("size: es obligatorio (cantidad de personas del equipo, minimo 2)")
        size = None
    elif not _is_int(size):
        errors.append("size: debe ser un entero, no %s" % _type_name(size))
        size = None
    elif size < 2:
        errors.append("size: %d no alcanza; un equipo tiene al menos 2 personas" % size)

    tracking = doc.get("tracking", "local")
    if tracking not in TRACKING_VALUES:
        errors.append("tracking: %r no es valido (opciones: %s)" % (tracking, ", ".join(TRACKING_VALUES)))

    names_versioned = doc.get("names_versioned", True)
    if not isinstance(names_versioned, bool):
        errors.append("names_versioned: debe ser true o false, no %s" % _type_name(names_versioned))
        names_versioned = None

    members = doc.get("members", [])
    if not isinstance(members, list):
        errors.append("members: debe ser una lista, no %s" % _type_name(members))
    else:
        if _is_int(size) and len(members) > size:
            errors.append("members: hay %d miembros y size es %d" % (len(members), size))
        seen = set()
        for index, member in enumerate(members):
            path = "members[%d]" % index
            errors.extend(_validate_member(member, path, tracking, names_versioned))
            if isinstance(member, dict) and isinstance(member.get("id"), str):
                if member["id"] in seen:
                    errors.append("%s.id: %s esta repetido" % (path, member["id"]))
                seen.add(member["id"])

    errors.extend(_credential_errors(doc, ""))
    return errors


def _member_ids(doc):
    members = doc.get("members") if isinstance(doc, dict) else None
    if not isinstance(members, list):
        return []
    return [m["id"] for m in members if isinstance(m, dict) and isinstance(m.get("id"), str)]


# ---------------------------------------------------------------- fichas

def _validate_dependencies(value, field):
    if not isinstance(value, str):
        return ["%s: debe ser texto, no %s" % (field, _type_name(value))]
    text = value.strip()
    if text.lower() in NO_DEPENDENCIES or text in NO_DEPENDENCIES:
        return []
    errors = []
    for item in text.split(","):
        item = item.strip()
        if not item or not (FICHA_ID_RE.match(item) or DEP_ISSUE_RE.match(item)):
            errors.append("%s: %r no es una dependencia valida (usa ninguna, —, IDs de ficha, #N o #N@sha "
                          "separados por comas)" % (field, item))
    return errors


def validate_ficha(fields, team_doc=None):
    """Valida los campos de una ficha. Devuelve {"ok", "errors", "warnings"}."""
    errors = []
    warnings = []
    if not isinstance(fields, dict):
        return {"ok": False, "errors": ["ficha: debe ser un objeto, no %s" % _type_name(fields)], "warnings": []}

    if "id" not in fields:
        errors.append("id: es obligatorio")
    elif not isinstance(fields["id"], str) or not FICHA_ID_RE.match(fields["id"]):
        errors.append("id: %r no es valido (letras, digitos, punto, guion o guion bajo, sin espacios)"
                      % (fields["id"],))

    for field in FICHA_FIELDS:
        if field not in fields and not (field == "Depende de" and LEGACY_DEPENDENCIES in fields):
            warnings.append("%s: falta en la ficha" % field)

    if "Modo" in fields and fields["Modo"] not in MODO_VALUES:
        errors.append("Modo: %r no es valido (opciones: %s)" % (fields["Modo"], ", ".join(MODO_VALUES)))

    for field in ("Depende de", LEGACY_DEPENDENCIES):
        if field in fields:
            errors.extend(_validate_dependencies(fields[field], field))

    if "Responsable" in fields:
        owner = fields["Responsable"]
        if not isinstance(owner, str):
            errors.append("Responsable: debe ser texto, no %s" % _type_name(owner))
        elif team_doc is not None:
            ids = _member_ids(team_doc)
            if owner not in ids:
                errors.append("Responsable: %s no es un miembro del equipo (miembros: %s)"
                              % (owner, ", ".join(ids) or "ninguno"))

    errors.extend(_credential_errors(fields, ""))
    return {"ok": not errors, "errors": errors, "warnings": warnings}


# ---------------------------------------------------------------- proyecto

def _validate_team_local(doc, tracking, known_ids):
    if not isinstance(doc, dict):
        return ["%s: la raiz debe ser un objeto JSON" % TEAM_LOCAL_PATH]
    errors = []
    for key in sorted(set(doc) - {"members"}):
        errors.append("%s: %s es un campo desconocido (permitido: members)" % (TEAM_LOCAL_PATH, key))
    members = doc.get("members", [])
    if not isinstance(members, list):
        errors.append("%s: members debe ser una lista" % TEAM_LOCAL_PATH)
        members = []
    seen = set()
    for index, member in enumerate(members):
        path = "%s: members[%d]" % (TEAM_LOCAL_PATH, index)
        errors.extend(_validate_member(member, path, tracking, None))
        if isinstance(member, dict) and isinstance(member.get("id"), str):
            member_id = member["id"]
            if member_id not in known_ids:
                errors.append("%s.id: %s no existe en %s" % (path, member_id, TEAM_PATH))
            if member_id in seen:
                errors.append("%s.id: %s esta repetido" % (path, member_id))
            seen.add(member_id)
    errors.extend("%s: %s" % (TEAM_LOCAL_PATH, e) for e in _credential_errors(doc, ""))
    return errors


def _validate_me_local(doc, known_ids):
    if not isinstance(doc, dict):
        return None, ["%s: la raiz debe ser un objeto JSON" % ME_LOCAL_PATH]
    errors = []
    for key in sorted(set(doc) - {"id"}):
        errors.append("%s: %s es un campo desconocido (permitido: id)" % (ME_LOCAL_PATH, key))
    errors.extend("%s: %s" % (ME_LOCAL_PATH, e) for e in _credential_errors(doc, ""))
    me = doc.get("id")
    if not isinstance(me, str):
        errors.append("%s: id es obligatorio y debe ser texto" % ME_LOCAL_PATH)
        return None, errors
    if me not in known_ids:
        errors.append("%s: id %s no es un miembro de %s (miembros: %s)"
                      % (ME_LOCAL_PATH, me, TEAM_PATH, ", ".join(known_ids) or "ninguno"))
    return me, errors


def _git(root, *args):
    """Comando git local de solo lectura. Devuelve (codigo, stdout) o (None, '')."""
    env = dict(os.environ)
    env["GIT_TERMINAL_PROMPT"] = "0"
    env["GIT_OPTIONAL_LOCKS"] = "0"
    try:
        completed = subprocess.run(
            ["git", "-C", str(root)] + list(args),
            stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL,
            timeout=GIT_TIMEOUT, env=env,
        )
    except (OSError, subprocess.SubprocessError):
        return None, ""
    return completed.returncode, completed.stdout.decode("utf-8", "replace").strip()


def _git_signal(root):
    """Devuelve (señal, hint) si Git sabe de un team.json, o (None, None)."""
    code, out = _git(root, "rev-parse", "--is-inside-work-tree")
    if code != 0 or out != "true":
        return None, None
    spec = "./" + TEAM_PATH
    code, _ = _git(root, "cat-file", "-e", "HEAD:" + spec)
    if code == 0:
        return ("%s existe en el último commit pero falta en la copia de trabajo" % TEAM_PATH,
                "Recuperalo con: git checkout HEAD -- %s" % TEAM_PATH)

    code, current = _git(root, "symbolic-ref", "-q", "--short", "HEAD")
    current = current if code == 0 else None
    candidates = []
    code, origin_head = _git(root, "symbolic-ref", "-q", "refs/remotes/origin/HEAD")
    if code == 0 and origin_head.startswith("refs/remotes/"):
        candidates.append(origin_head[len("refs/remotes/"):])
    candidates.extend(["origin/main", "origin/master", "main", "master"])
    for ref in candidates:
        if ref == current:
            continue
        code, _ = _git(root, "rev-parse", "--verify", "-q", ref + "^{commit}")
        if code != 0:
            continue
        code, _ = _git(root, "cat-file", "-e", ref + ":" + spec)
        if code == 0:
            return ("%s existe en la rama por defecto %s pero no en la rama actual" % (TEAM_PATH, ref),
                    "Tu rama no tiene el equipo: actualizá tu rama desde %s (según la copia local de Git; "
                    "conviene correr git fetch antes)." % ref)
        return None, None
    return None, None


def _foundation_signal(root):
    doc, error = _read_json(os.path.join(root, ".active-orchestrator-state.json"))
    if error or not isinstance(doc, dict):
        return None
    foundation = doc.get("foundation")
    phases = foundation.get("phases") if isinstance(foundation, dict) else None
    phase = phases.get("team-mode") if isinstance(phases, dict) else None
    if isinstance(phase, dict) and phase.get("status") == "done":
        return "la fundación registra la fase team-mode como hecha (.active-orchestrator-state.json)"
    return None


def _rules_signal(root):
    for name in ("CLAUDE.md", "AGENTS.md"):
        path = os.path.join(root, name)
        try:
            with open(path, "r", encoding="utf-8", errors="replace") as handle:
                if TEAM_MARKER in handle.read():
                    return "%s contiene las reglas de equipo (%s)" % (name, TEAM_MARKER)
        except OSError:
            continue
    return None


def _local_files_signal(root):
    present = [p for p in (ME_LOCAL_PATH, TEAM_LOCAL_PATH) if os.path.isfile(os.path.join(root, p))]
    if present:
        return "hay archivos locales del modo equipo sin %s: %s" % (TEAM_PATH, ", ".join(present))
    return None


def _missing_team(root):
    signals = []
    hint = None
    git_signal, git_hint = _git_signal(root)
    if git_signal:
        signals.append(git_signal)
        hint = git_hint
    state_signals = [s for s in (_foundation_signal(root), _rules_signal(root)) if s]
    signals.extend(state_signals)
    if state_signals and hint is None:
        hint = ("El proyecto es de equipo pero falta %s: pedile a quien fundó el equipo que lo suba, "
                "o recrealo con actualizar equipo." % TEAM_PATH)
    local_signal = _local_files_signal(root)
    if local_signal:
        signals.append(local_signal)
        if hint is None:
            hint = ("Puede que tu rama esté desactualizada: actualizá la rama desde la rama por defecto. "
                    "Si el equipo ya no existe, borrá los archivos locales (%s, %s)."
                    % (ME_LOCAL_PATH, TEAM_LOCAL_PATH))
    if not signals:
        return {"ok": True, "team_mode": False}
    return {
        "ok": False,
        "team_mode": "missing",
        "errors": ["falta %s en un proyecto con señales de modo equipo" % TEAM_PATH],
        "signals": signals,
        "hint": hint,
    }


def validate_project(root):
    """Valida el modo equipo del proyecto en root."""
    root = str(root)
    team_file = os.path.join(root, TEAM_PATH)
    if not os.path.isfile(team_file):
        return _missing_team(root)

    doc, error = _read_json(team_file)
    if error:
        return {"ok": False, "team_mode": True, "errors": [error.replace(team_file, TEAM_PATH)]}
    errors = validate_team(doc)
    if errors:
        return {"ok": False, "team_mode": True, "errors": errors}

    tracking = doc.get("tracking", "local")
    ids = _member_ids(doc)
    me = None

    local_file = os.path.join(root, TEAM_LOCAL_PATH)
    if os.path.isfile(local_file):
        local_doc, error = _read_json(local_file)
        if error:
            errors.append(error.replace(local_file, TEAM_LOCAL_PATH))
        else:
            errors.extend(_validate_team_local(local_doc, tracking, ids))

    me_file = os.path.join(root, ME_LOCAL_PATH)
    if os.path.isfile(me_file):
        me_doc, error = _read_json(me_file)
        if error:
            errors.append(error.replace(me_file, ME_LOCAL_PATH))
        else:
            me, me_errors = _validate_me_local(me_doc, ids)
            errors.extend(me_errors)

    if errors:
        return {"ok": False, "team_mode": True, "errors": errors}
    return {"ok": True, "team_mode": True, "size": doc["size"], "tracking": tracking,
            "members": ids, "me": me}


# ---------------------------------------------------------------- CLI

class _JsonArgumentParser(argparse.ArgumentParser):
    def error(self, message):
        _emit({"ok": False, "errors": ["uso incorrecto: %s" % message]})
        sys.stderr.write(self.format_usage())
        sys.exit(2)


def _emit(result):
    sys.stdout.write(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    sys.stdout.flush()


def _build_parser():
    parser = _JsonArgumentParser(prog="team.py", description="Valida el contrato del modo equipo.")
    parser.add_argument("--root", default=".", help="raiz del proyecto (por defecto, la carpeta actual)")
    sub = parser.add_subparsers(dest="command", parser_class=_JsonArgumentParser)
    sub.required = True
    validate = sub.add_parser("validate", help="valida .active-stack/team.json del proyecto")
    validate.add_argument("--file", help="valida este archivo como team.json (solo el contenido)")
    ficha = sub.add_parser("ficha", help="valida los campos de una ficha en JSON")
    ficha.add_argument("--file", required=True, help="JSON con los campos de la ficha")
    import delivery

    delivery.add_commands(sub)
    return parser


def main(argv=None):
    if hasattr(sys.stdout, "reconfigure"):
        try:
            sys.stdout.reconfigure(encoding="utf-8")
        except (ValueError, OSError):
            pass
    args = _build_parser().parse_args(argv)

    if args.command == "validate":
        if args.file:
            doc, error = _read_json(args.file)
            errors = [error] if error else validate_team(doc)
            result = {"ok": not errors, "errors": errors}
        else:
            result = validate_project(args.root)
    elif args.command == "ficha":
        fields, error = _read_json(args.file)
        if error:
            result = {"ok": False, "errors": [error], "warnings": []}
        else:
            team_doc = None
            team_file = os.path.join(args.root, TEAM_PATH)
            if os.path.isfile(team_file):
                loaded, team_error = _read_json(team_file)
                if not team_error and isinstance(loaded, dict):
                    team_doc = loaded
            result = validate_ficha(fields, team_doc)
    else:
        import delivery

        result = delivery.run(args)
    _emit(result)
    return 0 if result.get("ok") else 1


if __name__ == "__main__":
    sys.exit(main())
