import unittest

from roadmap import DELIVERY_STATUSES, RoadmapError, delivery_statuses
import test_roadmap


class DeliveryStatusTests(unittest.TestCase):
    setUp = test_roadmap.RoadmapTests.setUp
    begin = test_roadmap.RoadmapTests.begin

    def linked_task(self):
        published = self.begin("F1")["rows"][0]
        self.roadmap.result("F1", "created", remote_id="a1", url="https://app.clickup.com/t/a1")
        task = self.tasks[0]
        task.update(operation="status", stage="developing", current_remote_id="a1",
                    current_list_id="42", current_text=published["description"],
                    current_status="pendiente", list_statuses=list(DELIVERY_STATUSES.values()))
        return task

    def test_workflow_and_retry(self):
        task = self.linked_task()
        for stage in ("developing", "review", "changes_requested", "developing", "review", "merged", "archived"):
            with self.subTest(stage=stage):
                task["stage"] = stage
                preview = self.roadmap.preview(self.tasks)
                self.assertEqual(preview["rows"][0]["changes"], {"status": DELIVERY_STATUSES[stage]})
                self.roadmap.begin(self.tasks, preview["approval"], ["F1"])
                with self.assertRaises(RoadmapError):
                    self.roadmap.result("F1", "updated", remote_id="a1", url="https://app.clickup.com/t/a1")
                task["current_status"] = DELIVERY_STATUSES[stage].lower()
                self.roadmap.result("F1", "updated", remote_id="a1", url="https://app.clickup.com/t/a1",
                                    status=task["current_status"])
                self.assertEqual(self.roadmap.preview(self.tasks)["rows"][0]["action"], "unchanged")

    def test_status_update_checks_home_list_stage_and_available_statuses(self):
        task = self.linked_task()
        for field, value in (
            ("current_list_id", "99"), ("current_remote_id", "other"),
            ("current_text", "no marker"), ("stage", "merged"), ("list_statuses", ["Pendiente"]),
        ):
            with self.subTest(field=field):
                old = task[field]
                task[field] = value
                with self.assertRaises(RoadmapError):
                    self.roadmap.preview(self.tasks)
                task[field] = old
        self.assertEqual(self.roadmap.preview(self.tasks, access="error")["rows"][0]["action"], "local")

    def test_status_labels_preserve_clickup_spelling(self):
        expected = {stage: label.lower() for stage, label in DELIVERY_STATUSES.items()}
        self.assertEqual(delivery_statuses(list(expected.values())), expected)
        with self.assertRaisesRegex(RoadmapError, "ambiguos"):
            delivery_statuses([*expected.values(), "PENDIENTE"])
