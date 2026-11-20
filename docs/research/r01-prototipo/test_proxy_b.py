import json
import os
from pathlib import Path
import subprocess
import sys
import unittest

import test_hook


class ProxyB(unittest.TestCase):
    def setUp(self):
        self.fixture = test_hook.SaveFixture()
        self.fixture.setUp()
        self.addCleanup(self.fixture.doCleanups)

    def run_proxy(self, request):
        proxy = Path(__file__).resolve().with_name("proxy_b.py")
        res = subprocess.run([sys.executable, str(proxy), sys.executable, "-c",
                              "import sys; sys.stdout.buffer.write(sys.stdin.buffer.read())"],
                             cwd=self.fixture.repo, input=request, capture_output=True, check=True)
        return res.stdout

    def test_preserva_argumentos_y_sobre_del_mcp(self):
        self.fixture.git("switch", "-qc", "feat/R-01")
        args = self.fixture.event()["tool_input"]
        args["type"] = "decision"
        req = {"jsonrpc": "2.0", "method": "tools/call", "id": 8,
               "params": {"name": "mem_save", "arguments": args, "_meta": {"progressToken": 12}}}
        actual = json.loads(self.run_proxy((json.dumps(req) + "\n").encode()))
        expected = {**args, "title": "[feat-r-01] T", "topic_key": "branch/feat-r-01/K",
                    "content": "Branch: feat/R-01 · Author: Persona Piloto\nC"}
        self.assertEqual(actual, {**req, "params": {**req["params"], "arguments": expected}})

    def test_vias_narrativas_se_etiquetan_por_el_proxy(self):
        self.fixture.git("switch", "-qc", "feat/R-05")
        for name, campo in (("mem_session_summary", "content"), ("mem_session_finalize", "summary"),
                            ("mem_save_prompt", "content")):
            args = {"session_id": "S", campo: "texto"}
            req = {"jsonrpc": "2.0", "method": "tools/call", "id": 3, "params": {"name": name, "arguments": args}}
            actual = json.loads(self.run_proxy((json.dumps(req) + "\n").encode()))
            esperado = {"session_id": "S", campo: "Branch: feat/R-05 · Author: Persona Piloto\ntexto"}
            self.assertEqual(actual["params"]["arguments"], esperado, name)

    def test_reenvia_otros_metodos_y_guardado_sin_marcador(self):
        (self.fixture.repo / ".engram" / "active-stack.json").unlink()
        requests = [
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
            {"jsonrpc": "2.0", "method": "notifications/initialized"},
            {"jsonrpc": "2.0", "id": 2, "result": {"roots": []}},
            {"method": "tools/call", "params": {"name": "mem_search", "arguments": {"query": "T"}}},
            {"method": "tools/call", "params": {"name": "mem_save", "arguments": self.fixture.event()["tool_input"]}},
        ]
        payload = ("\n".join(map(json.dumps, requests)) + "\n").encode()
        self.assertEqual(self.run_proxy(payload), payload)

    def test_traza_acotada_a_dos_marcadores_y_campos_permitidos(self):
        from unittest.mock import patch
        marker = "r01b-" + "0" * 32
        trace = self.fixture.repo.parent / "trace.jsonl"
        req = {"method": "tools/call", "params": {"name": "mem_save", "arguments": {
            **self.fixture.event()["tool_input"], "title": marker, "private_extra": "not-for-trace"}}}
        with patch.dict(os.environ, {"R01_TRACE_MARKERS": marker + ",r01b-" + "1"*32,
                                     "R01_TRACE_FILE": str(trace)}):
            self.run_proxy((json.dumps(req) + "\n").encode())
            req["params"]["arguments"]["title"] = "OTHER"
            self.run_proxy((json.dumps(req) + "\n").encode())
        entries = trace.read_text().splitlines()
        self.assertEqual(len(entries), 1)
        saved = json.loads(entries[0])
        self.assertEqual(saved["incoming"]["title"], marker)
        self.assertEqual(saved["upstream"]["session_id"], "original")
        self.assertNotIn("not-for-trace", entries[0])

    def test_guardado_fuera_de_git_permite_continuar_la_sesion(self):
        self.fixture.repo = self.fixture.repo.parent
        payload = json.dumps({"method": "tools/call", "params": {"name": "mem_save",
                              "arguments": self.fixture.event()["tool_input"]}}).encode() + b"\n"
        payload += b'{"method":"ping","id":2}\n'
        self.assertEqual(self.run_proxy(payload), payload)
