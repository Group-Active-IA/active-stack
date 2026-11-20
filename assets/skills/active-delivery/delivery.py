"""Read-only team delivery inspection and explicitly confirmed local actions."""
import errno
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

import team


def _failure(message, **details):
    return {"ok": False, "errors": [message], **details}


def identity(root, args, state):
    doc, error = team._read_json(root / team.TEAM_PATH)
    if error:
        return _failure(error)
    members = {member["id"]: dict(member) for member in doc.get("members", [])}
    local_path = root / team.TEAM_LOCAL_PATH
    if local_path.is_file():
        local, error = team._read_json(local_path)
        if error:
            return _failure(error)
        for member in local.get("members", []):
            if member["id"] in members:
                members[member["id"]].update(member)

    me = state["me"]
    if args.select is not None:
        me = args.select
        if me not in members:
            return _failure("Select a member ID from the validated team.")
        tracked, _ = team._git(root, "ls-files", "--error-unmatch", "--", team.ME_LOCAL_PATH)
        ignored, _ = team._git(root, "check-ignore", "-q", "--", team.ME_LOCAL_PATH)
        if tracked != 1 or ignored != 0:
            return _failure("Identity must be untracked and ignored by Git.")
        target = root / team.ME_LOCAL_PATH
        if target.is_symlink() or target.parent.is_symlink():
            return _failure("Select identity in a regular project-local file.")
        temporary = None
        try:
            with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", dir=target.parent,
                                             delete=False) as handle:
                temporary = Path(handle.name)
                json.dump({"id": me}, handle)
                handle.write("\n")
            temporary.replace(target)
        except OSError:
            return _failure("Could not save the project-local identity.")
        finally:
            if temporary is not None and temporary.exists():
                temporary.unlink()

    accounts = [
        member["id"] for member in members.values()
        if (args.github and args.github.casefold() == (member.get("github") or "").casefold())
        or (args.clickup_id is not None and args.clickup_id == member.get("clickup_id"))
    ]
    suggested = accounts[0] if me is None and len(accounts) == 1 else None
    return {
        "ok": True,
        "team_mode": True,
        "me": me,
        "suggested": suggested,
        "mode": "identified" if me else "confirm" if suggested else "consultation",
        "members": [{"id": member["id"], "name": member.get("name")} for member in members.values()],
    }


def _worktrees(root):
    code, listing = team._git(root, "worktree", "list", "--porcelain", "-z")
    if code != 0:
        return None
    records = []
    for record in listing.split("\0\0"):
        fields = dict(line.partition(" ")[::2] for line in record.split("\0") if line)
        if "worktree" in fields:
            records.append((Path(fields["worktree"]).resolve(), fields.get("branch")))
    return records


def _preview_workspace(root, args):
    branch_ref = "refs/heads/" + args.branch
    checkout = root.parent / args.branch
    info = {"ok": True, "branch": args.branch, "path": str(checkout)}
    code, base = team._git(root, "rev-parse", "--verify", "--end-of-options",
                           args.base + "^{commit}")
    if code != 0:
        return _failure("Resolve the agreed base before starting a workspace.", **{
            "action": "unknown-base", "branch": args.branch, "path": str(checkout),
        })
    info["base_sha"] = base
    worktrees = _worktrees(root)
    if worktrees is None:
        return _failure("Could not inspect existing Git worktrees.")
    for path, ref in worktrees:
        if ref == branch_ref:
            code, changes = team._git(path, "status", "--porcelain", "--untracked-files=all")
            if code != 0:
                return _failure("Could not inspect the branch's existing worktree.",
                                action="occupied", path=str(path))
            return {**info, "action": "dirty" if changes else "occupied", "path": str(path)}
        if path == checkout:
            return {**info, "action": "collision"}
    if checkout.exists() or checkout.is_symlink():
        return {**info, "action": "collision"}

    code, changes = team._git(root, "status", "--porcelain", "--untracked-files=all")
    if code != 0:
        return _failure("Could not check for changes in the starting checkout.")
    if changes:
        return {**info, "action": "dirty", "path": str(root)}

    code, output = team._git(root, "for-each-ref", "--format=%(refname)", "refs/heads",
                             "refs/remotes")
    if code != 0:
        return _failure("Could not inspect existing branches.")
    refs = output.splitlines()
    prefix = args.id.casefold() + "-"
    # Split once past the remote name for remote-tracking branches.
    names = [ref.split("/", 2)[2] if ref.startswith("refs/heads/")
             else ref.split("/", 3)[3] for ref in refs]
    conflicting = sorted({name for name in names
                          if name.casefold().startswith(prefix) and name != args.branch})
    if conflicting:
        return {**info, "action": "branch-conflict", "branches": conflicting}
    if branch_ref in refs:
        _, tip = team._git(root, "rev-parse", "--verify", branch_ref)
        if not tip:
            return _failure("Could not resolve the existing local branch.")
        return {**info, "action": "attach", "tip_sha": tip}
    if args.branch in names:
        return {**info, "action": "remote-branch"}
    return {**info, "action": "create"}


def workspace(root, args, state):
    if state["me"] is None:
        return {"ok": True, "action": "consultation"}
    if (not team.FICHA_ID_RE.fullmatch(args.id)
            or not re.fullmatch(re.escape(args.id) + r"-[a-z0-9]+(?:-[a-z0-9]+)*", args.branch)):
        return _failure("Use a branch named <id>-<lowercase-slug>.")
    code, top = team._git(root, "rev-parse", "--show-toplevel")
    if code != 0 or Path(top).resolve() != root:
        return _failure("Run workspace from the root of a Git worktree.")
    if not args.apply:
        return _preview_workspace(root, args)

    code, common = team._git(root, "rev-parse", "--git-common-dir")
    if code != 0:
        return _failure("Could not locate the shared Git directory.")
    lock_path = root / common / ("active-delivery-" + args.id.casefold() + ".lock")
    try:
        with lock_path.open("a+b") as lock:
            lock.seek(0)
            try:
                if os.name == "nt":
                    import msvcrt
                    msvcrt.locking(lock.fileno(), msvcrt.LK_NBLCK, 1)
                else:
                    import fcntl
                    fcntl.flock(lock.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError as exc:
                if exc.errno in (errno.EACCES, errno.EAGAIN, errno.EDEADLK):
                    return {"ok": True, "action": "busy", "branch": args.branch}
                raise
            # The OS releases the lock on close or process exit. Retain the file
            # so another process always locks the same inode.
            return _apply_workspace(root, args)
    except OSError:
        return _failure("Could not lock this ficha in the shared Git directory.")


def _apply_workspace(root, args):
    preview = _preview_workspace(root, args)
    if preview.get("action") not in ("create", "attach"):
        return preview
    if preview["base_sha"] != args.expected_base:
        return _failure("The agreed base changed. Show a fresh preview and ask again.",
                        action="base-changed", base_sha=preview["base_sha"])
    if preview.get("tip_sha") != args.expected_tip:
        return _failure("The local branch changed. Show a fresh preview and ask again.",
                        action="branch-changed", tip_sha=preview.get("tip_sha"))
    if preview["action"] == "create":
        options = ["-b", args.branch, preview["path"], preview["base_sha"]]
    else:
        options = [preview["path"], args.branch]
    env = {**os.environ, "GIT_TERMINAL_PROMPT": "0"}
    try:
        result = subprocess.run(["git", "-C", str(root), "worktree", "add", *options],
                                stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL,
                                stderr=subprocess.DEVNULL, timeout=30, env=env)
    except (OSError, subprocess.SubprocessError):
        result = None
    if result is None or result.returncode != 0:
        observed = _preview_workspace(root, args)
        return _failure("Git did not confirm creation. Inspect this workspace before retrying.",
                        observed=observed)
    return {**preview, "action": "ready"}


def _observation(source, value):
    if not source.get("available") or value is None:
        return None
    return {"value": value, "source": source["source"], "checked_at": source["checked_at"]}


def _dependencies(ficha):
    field = ficha.get("Depende de", ficha.get(team.LEGACY_DEPENDENCIES, "ninguna"))
    return [] if field.strip().lower() in team.NO_DEPENDENCIES else [
        part.strip() for part in field.split(",")
    ]


def overview(root, evidence, state):
    sources = evidence["sources"]
    roadmap = sources.get("roadmap", {})
    github = sources.get("github", {})
    fichas = evidence.get("fichas", [])
    observations = evidence.get("observations", {})
    doc, error = team._read_json(root / team.TEAM_PATH)
    if error:
        return _failure(error)
    by_id = {}
    for ficha in fichas:
        validation = team.validate_ficha(ficha, doc)
        if not validation["ok"]:
            return _failure("Invalid roadmap ficha.", errors=validation["errors"])
        item = ficha["id"]
        if item in by_id:
            return _failure("Roadmap ficha ID appears more than once.", ficha=item)
        by_id[item] = ficha

    def observed(item, field):
        data = observations.get(item, {})
        return _observation(sources.get(data.get("source"), {}), data.get(field))

    def ficha_info(item):
        ficha = by_id[item]
        return {
            "id": item,
            "reference": _observation(roadmap, item),
            "assignee": (observed(item, "assignee") if "assignee" in observations.get(item, {})
                         and sources.get(observations[item].get("source"), {}).get("available")
                         else _observation(roadmap, ficha.get("Responsable"))),
            "status": observed(item, "status"),
        }

    def complete(status, *, is_pr=False):
        states = ("merged",) if is_pr else ("merged", "done", "omitted")
        return status and status["value"] in states

    by_id_info = {item: ficha_info(item) for item in by_id}
    mine = []
    waiting = []
    blocks = []
    for item, ficha in by_id.items():
        info = by_id_info[item]
        owner = info["assignee"]
        is_mine = bool(owner and owner["value"] == state["me"])
        if is_mine:
            mine.append(info)
        if complete(info["status"]):
            continue
        for dependency in _dependencies(ficha):
            # The pin is context for the base. Use just the PR number to read status.
            key = dependency.split("@", 1)[0]
            status = observed(key, "status")
            if complete(status, is_pr=key.startswith("#")):
                continue
            dependency_ficha = by_id_info.get(key)
            dependency_owner = (dependency_ficha["assignee"] if dependency_ficha
                                else observed(key, "assignee"))
            evidence = {"id": item, "reference": info["reference"],
                        "dependency": _observation(roadmap, dependency)}
            if is_mine:
                waiting.append({**evidence, "status": status, "assignee": dependency_owner})
            if dependency_ficha and dependency_owner and dependency_owner["value"] == state["me"]:
                blocks.append({**evidence, "assignee": owner})

    return {"ok": True, "me": state["me"], "sources": sources, "mine": mine,
            "waiting": waiting, "blocks": blocks,
            "reviews": [_observation(github, review) for review in evidence["reviews"]]
            if github.get("available") and "reviews" in evidence else None}


def _overlap(left, right):
    return any(a == b or a.startswith(b + "/") or b.startswith(a + "/")
               for a in left for b in right)


def notices(evidence, state):
    findings = {}
    uncertain = set()
    assignment = evidence.get("assignment")
    if assignment is None:
        uncertain.add(1)
    else:
        occupied = {field: assignment[field] for field in ("branch", "pr")
                    if assignment.get(field)}
        if assignment.get("assignee") and assignment["assignee"] != state["me"]:
            occupied["assignee"] = assignment["assignee"]
        if occupied:
            findings[1] = occupied

    dependency = evidence.get("dependency") or {}
    expected, actual = dependency.get("expected"), dependency.get("actual")
    if expected:
        if not actual:
            uncertain.add(2)
        elif not actual.lower().startswith(expected.lower()):
            findings[2] = {"expected": expected, "actual": actual}

    zone = [path.rstrip("/") for path in evidence.get("zone", [])]
    open_prs = evidence.get("open_prs")
    mode = evidence.get("mode")
    if open_prs is None or not zone:
        uncertain.add(3)
        if mode == "odd":
            uncertain.add(4)
    else:
        overlaps = []
        sdd_overlaps = []
        for pr in open_prs:
            if pr["number"] == evidence.get("current_pr"):
                continue
            paths = pr.get("paths")
            if paths is None:
                uncertain.add(3)
                if mode == "odd":
                    uncertain.add(4)
                continue
            if _overlap(zone, paths):
                overlaps.append(pr["number"])
            if mode == "odd" and any(
                    path.startswith("openspec/changes/")
                    and not path.startswith("openspec/changes/archive/") for path in paths):
                sdd_zone = pr.get("sdd_zone")
                if sdd_zone is None:
                    uncertain.add(4)
                elif _overlap(zone, sdd_zone):
                    sdd_overlaps.append(pr["number"])
        if overlaps:
            findings[3] = {"prs": overlaps}
        if sdd_overlaps:
            findings[4] = {"prs": sdd_overlaps}

    meets_requirements = evidence.get("requirements_met")
    if meets_requirements is False:
        findings[5] = {}
    elif meets_requirements is not True:
        uncertain.add(5)

    return {"ok": True, "notices": [{"id": key, **details} for key, details in findings.items()],
            "unconfirmed": sorted(uncertain)}


def inspect(root, args, state):
    evidence, error = team._read_json(args.file)
    if error:
        return _failure(error)
    try:
        if args.command == "overview":
            return overview(root, evidence, state)
        return notices(evidence, state)
    except (AttributeError, KeyError, TypeError):
        return _failure("Evidence does not match the team delivery input format.")


def add_commands(sub):
    who = sub.add_parser("identity", help="inspect or explicitly select the local team identity")
    who.add_argument("--github")
    who.add_argument("--clickup-id", type=int)
    who.add_argument("--select")
    space = sub.add_parser("workspace", help="preview or explicitly create a ficha worktree")
    space.add_argument("--id", required=True)
    space.add_argument("--branch", required=True)
    space.add_argument("--base", required=True)
    space.add_argument("--apply", action="store_true")
    space.add_argument("--expected-base")
    space.add_argument("--expected-tip")
    for name in ("overview", "notices"):
        inspection = sub.add_parser(name, help="summarize read-only team delivery evidence")
        inspection.add_argument("--file", required=True)


def run(args):
    root = Path(args.root).resolve()
    state = team.validate_project(root)
    if not state["ok"] or not state["team_mode"]:
        return state
    action = {"identity": identity, "workspace": workspace}.get(args.command, inspect)
    return action(root, args, state)
