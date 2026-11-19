"""Contrato del modo equipo: tests de team.py y sus lanzadores.

Correr desde esta carpeta con: python -B -m unittest test_team
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))

import team  # noqa: E402

GITHUB_TOKEN = "ghp_" + "a" * 36
CLICKUP_TOKEN = "pk_12345678_" + "A" * 32


def errors_text(result):
    return " ".join(result.get("errors", []))


class Project:
    """Proyecto temporal con helpers para escribir los archivos del contrato."""

    def __init__(self, root):
        self.root = Path(root)

    def write(self, relative, content):
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        if not isinstance(content, str):
            content = json.dumps(content, ensure_ascii=False)
        path.write_text(content, encoding="utf-8")
        return path

    def team(self, doc):
        return self.write(".active-stack/team.json", doc)

    def git(self, *args):
        return subprocess.run(
            ["git", "-C", str(self.root), *args],
            check=True, capture_output=True, text=True,
            env={**os.environ, "GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@t",
                 "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@t"},
        )


class TempProjectTest(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.project = Project(self._tmp.name)

    def tearDown(self):
        self._tmp.cleanup()

    def validate(self):
        return team.validate_project(self.project.root)


class TeamDocumentTest(unittest.TestCase):
    """Reglas de team.json: validate_team(doc) devuelve la lista de errores."""

    def assertValid(self, doc):
        self.assertEqual(team.validate_team(doc), [])

    def assertInvalid(self, doc, fragment):
        errors = team.validate_team(doc)
        self.assertTrue(errors, f"debía rechazarse: {doc}")
        self.assertIn(fragment, " ".join(errors))

    # T-F1-1
    def test_minimal_only_size_is_valid(self):
        self.assertValid({"size": 2})

    def test_complete_document_is_valid(self):
        self.assertValid({
            "version": 1, "size": 3, "tracking": "clickup", "names_versioned": True,
            "members": [
                {"id": "m1", "name": "Ana", "github": "ana-dev", "clickup_id": 12345},
                {"id": "m2", "name": None},
            ],
        })

    # T-F1-2
    def test_size_below_two_is_rejected(self):
        for size in (1, 0, -3):
            self.assertInvalid({"size": size}, "size")

    def test_size_must_be_an_integer(self):
        for size in (2.5, "3", True, None):
            self.assertInvalid({"size": size}, "size")

    def test_size_is_required(self):
        self.assertInvalid({"version": 1}, "size")

    # T-F1-3
    def test_members_without_name_are_valid(self):
        self.assertValid({"size": 2, "members": [{"id": "m1", "name": None}, {"id": "m2"}]})

    def test_member_ids_are_unique(self):
        self.assertInvalid({"size": 2, "members": [{"id": "m1"}, {"id": "m1"}]}, "m1")

    def test_member_id_format(self):
        for bad in ("M1", "", "m 1", 1, None):
            self.assertInvalid({"size": 2, "members": [{"id": bad}]}, "id")

    def test_fewer_members_than_size_is_valid(self):
        self.assertValid({"size": 4, "members": [{"id": "m1"}]})

    def test_more_members_than_size_is_rejected(self):
        self.assertInvalid({"size": 2, "members": [{"id": "m1"}, {"id": "m2"}, {"id": "m3"}]}, "size")

    # T-F1-4
    def test_clickup_id_requires_clickup_tracking(self):
        self.assertInvalid({"size": 2, "tracking": "local", "members": [{"id": "m1", "clickup_id": 1}]}, "clickup_id")
        self.assertInvalid({"size": 2, "members": [{"id": "m1", "clickup_id": 1}]}, "clickup_id")

    def test_clickup_id_must_be_an_integer(self):
        self.assertInvalid({"size": 2, "tracking": "clickup", "members": [{"id": "m1", "clickup_id": "12"}]}, "clickup_id")

    def test_unknown_tracking_is_rejected(self):
        self.assertInvalid({"size": 2, "tracking": "jira"}, "tracking")

    # T-F1-5
    def test_unknown_version_gives_a_clear_error(self):
        errors = team.validate_team({"version": 2, "size": 2})
        self.assertTrue(errors)
        self.assertIn("version", errors[0])
        self.assertIn("1", errors[0])

    def test_version_must_be_an_integer(self):
        self.assertInvalid({"version": "1", "size": 2}, "version")

    # T-F1-6
    def test_private_names_keep_personal_data_out_of_shared_file(self):
        for field, value in (("name", "Ana"), ("github", "ana-dev")):
            self.assertInvalid({"size": 2, "names_versioned": False,
                                "members": [{"id": "m1", field: value}]}, "names_versioned")
        self.assertInvalid({"size": 2, "tracking": "clickup", "names_versioned": False,
                            "members": [{"id": "m1", "clickup_id": 7}]}, "names_versioned")

    def test_private_names_allow_ids_only(self):
        self.assertValid({"size": 4, "names_versioned": False})
        self.assertValid({"size": 2, "names_versioned": False, "members": [{"id": "m1"}, {"id": "m2"}]})

    def test_names_versioned_must_be_boolean(self):
        self.assertInvalid({"size": 2, "names_versioned": "no"}, "names_versioned")

    # T-F1-8
    def test_unknown_fields_are_rejected(self):
        self.assertInvalid({"size": 2, "token": "abc"}, "token")
        self.assertInvalid({"size": 2, "members": [{"id": "m1", "password": "x"}]}, "password")

    def test_values_that_look_like_credentials_are_rejected(self):
        for secret in (GITHUB_TOKEN, CLICKUP_TOKEN, "github_pat_11ABCDEF", "Bearer abc.def", "sk-proj-abc123"):
            self.assertInvalid({"size": 2, "members": [{"id": "m1", "name": secret}]}, "credencial")

    def test_github_must_be_a_handle(self):
        for bad in ("-ana", "ana_dev", "a" * 40, "github_pat_11ABC"):
            self.assertInvalid({"size": 2, "members": [{"id": "m1", "github": bad}]}, "github")

    def test_root_must_be_an_object(self):
        self.assertInvalid([1, 2], "objeto")
        self.assertInvalid({"size": 2, "members": "m1"}, "members")


class FichaTest(unittest.TestCase):
    """validate_ficha(campos, team) -> {"ok", "errors", "warnings"}."""

    OLD = {
        "id": "C-01",
        "Estado": "[ ] pendiente",
        "Scope": "Scaffolding del monorepo",
        "Dependencias": "ninguna",
        "Governance": "BAJO",
        "Leer antes": "knowledge-base/01_vision_y_objetivos.md",
    }
    TEAM = {"size": 2, "members": [{"id": "m1"}, {"id": "m2"}]}

    # T-F1-7
    def test_old_ficha_without_new_fields_is_still_valid(self):
        result = team.validate_ficha(self.OLD)
        self.assertTrue(result["ok"], result)
        self.assertEqual(result["errors"], [])
        warnings = " ".join(result["warnings"])
        for field in ("Resultado", "Zona", "Modo", "Responsable", "Requiere"):
            self.assertIn(field, warnings)

    def test_new_ficha_with_all_fields_is_valid(self):
        ficha = {
            "id": "R-11", "Resultado": "El agente muestra qué me toca", "Depende de": "R-06, #12@1a2b3c4",
            "Zona": "assets/skills/active-delivery/", "Requiere": "claude o codex · windows y linux",
            "Modo": "odd", "Responsable": "m2",
        }
        result = team.validate_ficha(ficha, self.TEAM)
        self.assertEqual(result, {"ok": True, "errors": [], "warnings": []})

    def test_id_is_required_and_well_formed(self):
        self.assertFalse(team.validate_ficha({"Scope": "x"})["ok"])
        self.assertFalse(team.validate_ficha({"id": "C 01"})["ok"])

    def test_modo_is_odd_or_sdd(self):
        result = team.validate_ficha({**self.OLD, "Modo": "rapido"})
        self.assertFalse(result["ok"])
        self.assertIn("Modo", errors_text(result))

    def test_responsable_must_be_a_team_member(self):
        result = team.validate_ficha({**self.OLD, "Responsable": "m9"}, self.TEAM)
        self.assertFalse(result["ok"])
        self.assertIn("m9", errors_text(result))

    def test_dependency_format(self):
        for good in ("ninguna", "—", "C-01", "C-01, C-02", "#12", "#12@1a2b3c4"):
            self.assertTrue(team.validate_ficha({**self.OLD, "Dependencias": good})["ok"], good)
        for bad in ("C-01 y C-02", "#abc", "#12@xyz", "@1a2b3c4"):
            self.assertFalse(team.validate_ficha({**self.OLD, "Depende de": bad})["ok"], bad)

    # T-F1-8
    def test_ficha_fields_do_not_accept_credentials(self):
        result = team.validate_ficha({**self.OLD, "Scope": "usar " + GITHUB_TOKEN})
        self.assertFalse(result["ok"])
        self.assertIn("credencial", errors_text(result))


class ProjectWithTeamTest(TempProjectTest):
    """validate_project(root) con .active-stack/team.json presente."""

    def test_valid_team_returns_team_data(self):
        self.project.team({"size": 3, "tracking": "local", "members": [{"id": "m1"}, {"id": "m2"}]})
        result = self.validate()
        self.assertEqual(result, {
            "ok": True, "team_mode": True, "size": 3, "tracking": "local",
            "members": ["m1", "m2"], "me": None,
        })

    def test_invalid_team_returns_errors(self):
        self.project.team({"size": 1})
        result = self.validate()
        self.assertFalse(result["ok"])
        self.assertIs(result["team_mode"], True)
        self.assertIn("size", errors_text(result))

    def test_broken_json_is_reported(self):
        self.project.team("{size: 2")
        result = self.validate()
        self.assertFalse(result["ok"])
        self.assertIn("JSON", errors_text(result))

    def test_me_local_identifies_the_person(self):
        self.project.team({"size": 2, "members": [{"id": "m1"}, {"id": "m2"}]})
        self.project.write(".active-stack/me.local.json", {"id": "m2"})
        self.assertEqual(self.validate()["me"], "m2")

    def test_me_local_with_unknown_id_is_rejected(self):
        self.project.team({"size": 2, "members": [{"id": "m1"}]})
        self.project.write(".active-stack/me.local.json", {"id": "m7"})
        result = self.validate()
        self.assertFalse(result["ok"])
        self.assertIn("m7", errors_text(result))

    def test_team_local_holds_private_names(self):
        self.project.team({"size": 2, "names_versioned": False, "members": [{"id": "m1"}, {"id": "m2"}]})
        self.project.write(".active-stack/team.local.json",
                           {"members": [{"id": "m1", "name": "Ana", "github": "ana-dev"}]})
        result = self.validate()
        self.assertTrue(result["ok"], result)
        self.assertEqual(result["members"], ["m1", "m2"])

    def test_team_local_with_unknown_id_is_rejected(self):
        self.project.team({"size": 2, "names_versioned": False, "members": [{"id": "m1"}]})
        self.project.write(".active-stack/team.local.json", {"members": [{"id": "m5", "name": "Ana"}]})
        result = self.validate()
        self.assertFalse(result["ok"])
        self.assertIn("m5", errors_text(result))

    def test_local_files_do_not_accept_credentials(self):
        self.project.team({"size": 2, "names_versioned": False, "members": [{"id": "m1"}]})
        self.project.write(".active-stack/team.local.json", {"members": [{"id": "m1", "name": CLICKUP_TOKEN}]})
        result = self.validate()
        self.assertFalse(result["ok"])
        self.assertIn("credencial", errors_text(result))


class ProjectWithoutTeamTest(TempProjectTest):
    """R06-2: sin team.json se decide con señales, no solo con la ausencia."""

    def assertMissing(self, signal_fragment, hint_fragment):
        result = self.validate()
        self.assertFalse(result["ok"], result)
        self.assertEqual(result["team_mode"], "missing")
        self.assertIn(signal_fragment, " ".join(result["signals"]))
        self.assertIn(hint_fragment, result["hint"])

    def test_no_file_and_no_signals_is_an_individual_project(self):
        self.project.write("README.md", "proyecto individual")
        self.project.write(".active-orchestrator-state.json", {"version": 5, "foundation": {"version": 1, "phases": {}}})
        self.assertEqual(self.validate(), {"ok": True, "team_mode": False})

    def test_foundation_state_marks_team_project(self):
        self.project.write(".active-orchestrator-state.json",
                           {"version": 5, "foundation": {"version": 1, "phases": {"team-mode": {"status": "done"}}}})
        self.assertMissing("fundación", "actualizar equipo")

    def test_skipped_or_deferred_team_phase_is_not_a_signal(self):
        for status in ("skipped", "deferred", "pending"):
            self.project.write(".active-orchestrator-state.json",
                               {"version": 5, "foundation": {"version": 1, "phases": {"team-mode": {"status": status}}}})
            self.assertEqual(self.validate(), {"ok": True, "team_mode": False}, status)

    def test_malformed_foundation_state_is_ignored(self):
        self.project.write(".active-orchestrator-state.json", "{no es json")
        self.assertEqual(self.validate(), {"ok": True, "team_mode": False})

    def test_team_rules_marker_marks_team_project(self):
        for name in ("CLAUDE.md", "AGENTS.md"):
            with self.subTest(name):
                (self.project.root / "CLAUDE.md").unlink(missing_ok=True)
                (self.project.root / "AGENTS.md").unlink(missing_ok=True)
                self.project.write(name, "# Reglas\n\n<!-- active-stack:team-mode -->\nreglas\n<!-- /active-stack:team-mode -->\n")
                self.assertMissing("reglas de equipo", "actualizar equipo")

    def test_leftover_local_files_mark_team_project(self):
        self.project.write(".active-stack/me.local.json", {"id": "m1"})
        self.assertMissing("archivos locales", "rama")

    def test_file_in_last_commit_but_deleted_in_working_tree(self):
        self.project.git("init", "-q", "-b", "main")
        self.project.team({"size": 2})
        self.project.git("add", ".")
        self.project.git("commit", "-q", "-m", "equipo")
        (self.project.root / ".active-stack" / "team.json").unlink()
        self.assertMissing("último commit", "git checkout HEAD -- .active-stack/team.json")

    def test_file_in_default_branch_but_not_in_current_branch(self):
        self.project.git("init", "-q", "-b", "main")
        self.project.write("README.md", "base")
        self.project.git("add", ".")
        self.project.git("commit", "-q", "-m", "base")
        self.project.git("switch", "-q", "-c", "feat/vieja")
        self.project.git("switch", "-q", "main")
        self.project.team({"size": 2})
        self.project.git("add", ".")
        self.project.git("commit", "-q", "-m", "equipo")
        self.project.git("switch", "-q", "feat/vieja")
        self.assertMissing("main", "actualizá tu rama")

    def test_not_a_git_repository_still_answers(self):
        self.assertEqual(self.validate(), {"ok": True, "team_mode": False})


class CommandLineTest(TempProjectTest):
    """team.py --root <proyecto> validate|ficha: JSON por stdout y código de salida."""

    def run_cli(self, *args):
        completed = subprocess.run(
            [sys.executable, "-B", str(HERE / "team.py"), *args],
            capture_output=True, text=True, encoding="utf-8",
        )
        return completed.returncode, json.loads(completed.stdout)

    def test_individual_project_exits_zero(self):
        code, out = self.run_cli("--root", str(self.project.root), "validate")
        self.assertEqual((code, out), (0, {"ok": True, "team_mode": False}))

    def test_invalid_team_exits_one(self):
        self.project.team({"size": 1})
        code, out = self.run_cli("--root", str(self.project.root), "validate")
        self.assertEqual(code, 1)
        self.assertFalse(out["ok"])

    def test_missing_team_exits_one(self):
        self.project.write(".active-stack/me.local.json", {"id": "m1"})
        code, out = self.run_cli("--root", str(self.project.root), "validate")
        self.assertEqual((code, out["team_mode"]), (1, "missing"))

    def test_validate_a_draft_file(self):
        draft = self.project.write("borrador.json", {"size": 1})
        code, out = self.run_cli("--root", str(self.project.root), "validate", "--file", str(draft))
        self.assertEqual(code, 1)
        self.assertIn("size", errors_text(out))
        self.assertFalse((self.project.root / ".active-stack" / "team.json").exists())

    def test_ficha_command(self):
        self.project.team({"size": 2, "members": [{"id": "m1"}]})
        ficha = self.project.write("ficha.json", {"id": "C-01", "Responsable": "m4"})
        code, out = self.run_cli("--root", str(self.project.root), "ficha", "--file", str(ficha))
        self.assertEqual(code, 1)
        self.assertIn("m4", errors_text(out))


class LauncherTest(TempProjectTest):
    """Los lanzadores eligen el intérprete; el agente siempre llama al mismo comando."""

    def test_platform_launcher_runs_the_validator(self):
        self.project.team({"size": 1})
        if os.name == "nt":
            command = ["cmd", "/c", str(HERE / "team.bat"), "--root", str(self.project.root), "validate"]
        else:
            command = ["sh", str(HERE / "team.sh"), "--root", str(self.project.root), "validate"]
        completed = subprocess.run(command, capture_output=True, text=True, encoding="utf-8")
        self.assertEqual(completed.returncode, 1, completed.stderr)
        self.assertFalse(json.loads(completed.stdout)["ok"])

    def test_launchers_report_missing_python(self):
        empty = Path(self._tmp.name) / "sin-python"
        empty.mkdir()
        env = {**os.environ, "PATH": str(empty)}
        if os.name == "nt":
            env["PATH"] = os.path.join(os.environ.get("SystemRoot", r"C:\Windows"), "System32")
            command = ["cmd", "/c", str(HERE / "team.bat"), "--root", str(self.project.root), "validate"]
        else:
            command = [shutil.which("sh") or "/bin/sh", str(HERE / "team.sh"), "--root", str(self.project.root), "validate"]
        completed = subprocess.run(command, capture_output=True, text=True, encoding="utf-8", env=env)
        self.assertEqual(completed.returncode, 3, completed.stdout + completed.stderr)
        self.assertIn("Python", json.loads(completed.stdout)["errors"][0])

    def test_batch_launcher_uses_crlf(self):
        data = (HERE / "team.bat").read_bytes()
        self.assertTrue(data.endswith(b"\r\n"))
        self.assertEqual(data.count(b"\n"), data.count(b"\r\n"))


class ContractDocumentTest(unittest.TestCase):
    """references/team-mode.md documenta todo lo que team.py hace cumplir."""

    def test_reference_documents_every_rule(self):
        text = (HERE / "references" / "team-mode.md").read_text(encoding="utf-8")
        required = [
            ".active-stack/team.json", ".active-stack/team.local.json", ".active-stack/me.local.json",
            *sorted(team.ROOT_FIELDS), *sorted(team.MEMBER_FIELDS),
            "team-mode", "<!-- active-stack:team-mode -->", "team_mode", "missing",
            "team.sh", "team.bat", "validate", "ficha", "--file",
            *team.FICHA_FIELDS,
        ]
        for fragment in required:
            self.assertIn(fragment, text, f"falta documentar {fragment!r}")


if __name__ == "__main__":
    unittest.main()
