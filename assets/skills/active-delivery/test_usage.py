import copy
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

from usage import EvidenceError, calculate


def evidence():
    return {
        "branch": "C-01-work",
        "pr": {"state": "MERGED", "headRefName": "C-01-work", "baseRefName": "main",
               "mergedAt": "2026-11-25T14:00:00Z", "mergeCommit": {"oid": "a" * 40},
               "url": "https://github.com/org/repo/pull/1"},
        "intervals": [
            {"start": "2026-11-25T10:00:00-03:00", "end": "2026-11-25T13:30:00Z", "source": "branch timer"},
            {"start": "2026-11-25T13:20:00Z", "end": "2026-11-25T13:50:00Z", "source": "review timer"},
        ],
        "usage_status": "complete",
        "rates": [
            {"id": "api-in", "provider": "provider-a", "model": "model-a", "tier": "standard",
             "category": "input", "usd_per_million": "3.50", "source": "https://example.com/pricing",
             "checked_on": "2026-11-25"},
            {"id": "sub-out", "provider": "provider-b", "model": "model-b", "tier": "standard",
             "category": "output", "usd_per_million": "0.000000012345", "source": "https://example.org/pricing",
             "checked_on": "2026-11-25"},
        ],
        "usage": [
            {"id": "request-a-input", "billing": "api", "rate_id": "api-in", "source": "response usage",
             "tokens": 1},
            {"id": "session-b-output", "billing": "subscription", "rate_id": "sub-out",
             "source": "session counter", "counter_id": "session-b-output",
             "start": "2026-11-25T13:00:00Z", "end": "2026-11-25T13:30:00Z",
             "before": 90, "after": 97},
        ],
    }


class UsageTest(unittest.TestCase):
    def test_overlap_dedupe_and_decimal_rates(self):
        data = evidence()
        data["usage"].append(copy.deepcopy(data["usage"][0]))
        report = calculate(data)
        self.assertEqual(report["active_seconds"], "3000.000000")
        self.assertEqual(report["elapsed_to_merge_seconds"], "3600.000000")
        self.assertEqual(report["total_tokens"], 8)
        self.assertEqual([row["model"] for row in report["model_rates"]], ["model-a", "model-b"])
        self.assertEqual(report["api_rate_cost_usd"], "0.00000350")
        self.assertEqual(report["subscription_api_equivalent_usd"], "0.000000000000086415")
        self.assertEqual(report["total_at_api_rates_usd"], "0.000003500000086415")

    def test_adjacent_counter_windows_are_deltas(self):
        data = evidence()
        next_window = dict(data["usage"][1], id="continued", start="2026-11-25T13:30:00Z",
                           end="2026-11-25T13:45:00Z", before=97, after=100)
        data["usage"].append(next_window)
        self.assertEqual(calculate(data)["total_tokens"], 11)
        next_window["start"] = "2026-11-25T13:29:59Z"
        with self.assertRaisesRegex(EvidenceError, "overlap"):
            calculate(data)
        next_window["start"] = "2026-11-25T13:30:00Z"
        next_window["before"] = 98
        with self.assertRaisesRegex(EvidenceError, "disagree"):
            calculate(data)

    def test_invalid_or_missing_evidence_blocks_archive(self):
        cases = (
            ("open PR", lambda data: data["pr"].update(state="CLOSED"), "confirm a merge"),
            ("wrong branch", lambda data: data["pr"].update(headRefName="main"), "confirm a merge"),
            ("missing time", lambda data: data.update(intervals=[]), "measured interval"),
            ("post-merge time", lambda data: data["intervals"][0].update(end="2026-11-25T14:01:00Z"), "on or before merge"),
            ("missing rates", lambda data: data.update(rates=[]), "missing a verified rate"),
            ("missing usage", lambda data: data.update(usage=[]), "usage is required"),
            ("unknown usage", lambda data: data.update(usage_status="unknown"), "incomplete evidence"),
            ("counter falls", lambda data: data["usage"][1].update(after=89), "decreasing"),
            ("conflicting duplicate", lambda data: data["usage"].append(dict(data["usage"][0], tokens=10)), "conflicting"),
            ("not a count", lambda data: data["usage"][0].update(tokens=True), "integer token count"),
            ("sensitive source", lambda data: data["usage"][0].update(source="sk-demo"), "non-sensitive"),
        )
        for name, change, error in cases:
            with self.subTest(name=name):
                data = evidence()
                change(data)
                with self.assertRaisesRegex(EvidenceError, error):
                    calculate(data)

    def test_no_ai_must_be_explicit(self):
        data = evidence()
        data.update(usage_status="none", rates=[], usage=[])
        self.assertEqual(calculate(data)["total_at_api_rates_usd"], "0")

    def test_cli_rejects_duplicate_keys_and_outputs_json_errors(self):
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / "evidence.json"
            source.write_text('{"branch":"one","branch":"two"}', encoding="utf-8")
            script = Path(__file__).resolve().with_name("usage.py")
            result = subprocess.run([sys.executable, "-B", str(script), str(source)],
                                    capture_output=True, text=True)
        self.assertEqual(result.returncode, 1)
        self.assertEqual(json.loads(result.stdout), {"ok": False, "error": "JSON contains a duplicate field"})


if __name__ == "__main__":
    unittest.main()
