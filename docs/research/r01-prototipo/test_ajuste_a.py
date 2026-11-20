import json
from pathlib import Path
import subprocess
import sys
import unittest

import test_hook


class AjusteA(unittest.TestCase):
    def ensayar(self, event):
        archivo = Path(__file__).with_name("hook_ajuste_a.py")
        p = subprocess.run([sys.executable, str(archivo)], input=json.dumps(event),
                           text=True, check=True, capture_output=True)
        return json.loads(p.stdout)["hookSpecificOutput"]["updatedInput"] if p.stdout else None

    def setUp(self):
        self.fixture = test_hook.SaveFixture()
        self.fixture.setUp()
        self.addCleanup(self.fixture.doCleanups)

    def test_solo_campos_del_stack(self):
        self.fixture.git("switch", "-qc", "feat/R-01")
        event = self.fixture.event()
        event["tool_input"]["type"] = "discovery"
        self.assertEqual(self.ensayar(event), {
            "title": "[feat-r-01] T", "topic_key": "branch/feat-r-01/K",
            "content": "Branch: feat/R-01 · Author: Persona Piloto\nC",
        })

    def test_sin_clave_en_main(self):
        event = self.fixture.event()
        del event["tool_input"]["topic_key"]
        self.assertEqual(self.ensayar(event), {"title": "T", "content": "Author: Persona Piloto\nC"})
