import json
import unittest
from observador_autoria import extract


class AuthorTest(unittest.TestCase):
    def event(self):
        return {"hook_event_name": "PostToolUse", "session_id": "S", "agent_id": "H",
                "tool_name": "mcp__engram__mem_save", "tool_use_id": "call1",
                "tool_input": {"title":"identico", "topic_key":"alfa"},
                "tool_response": {"content":[{"type":"text","text":json.dumps(
                    {"candidates":[{"id":999}],"id":7,"result":"Memory saved: test"})}]}}

    def test_post_elige_id_de_la_observacion(self):
        self.assertEqual(extract(self.event(), ["alfa","beta"]), {"phase":"PostToolUse",
            "session_id":"S","agent_id":"H","tool_use_id":"call1","observation_id":7})

    def test_pre_omite_autor_ausente_y_no_registra_datos(self):
        event = self.event()
        event["hook_event_name"] = "PreToolUse"
        del event["agent_id"]
        self.assertEqual(extract(event, ["alfa","beta"]), {"phase":"PreToolUse",
            "session_id":"S","tool_use_id":"call1"})

    def test_descarta_otras_llamadas_y_no_infiere_id(self):
        event = self.event()
        self.assertIsNone(extract(event, ["otra"]))
        event["tool_response"] = {"content":[{"type":"text","text":'{"candidates":[{"id":999}]}'}]}
        self.assertNotIn("observation_id",extract(event, ["alfa","beta"]))
