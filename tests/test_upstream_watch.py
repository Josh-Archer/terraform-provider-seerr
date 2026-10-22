import json
import sys
import unittest
from pathlib import Path
from unittest.mock import call, patch


sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))
import upstream_watch


class UpstreamWatchTests(unittest.TestCase):
    def status(self, seerr="v3.5.0", overseerr="v1.35.0"):
        return upstream_watch.ReleaseStatus(
            latest={
                "seerr": seerr,
                "jellyseerr": "v2.7.3",
                "overseerr": overseerr,
            },
            verified={
                "seerr": "v3.5.0",
                "jellyseerr": "v2.7.3",
                "overseerr": "v1.35.0",
            },
        )

    @patch.object(upstream_watch, "run_gh")
    def test_repeated_status_creates_once_then_leaves_comment_unchanged(self, run_gh):
        run_gh.side_effect = ["[]", "{\"id\": 41}", json.dumps([
            {"id": 41, "body": upstream_watch.render_status(self.status())}
        ])]
        body = upstream_watch.render_status(self.status())

        self.assertEqual(upstream_watch.reconcile_comment(body), "created")
        self.assertEqual(upstream_watch.reconcile_comment(body), "unchanged")
        self.assertEqual(run_gh.call_count, 3)
        self.assertIn("POST", run_gh.call_args_list[1].args[0])
        self.assertNotIn("POST", run_gh.call_args_list[2].args[0])
        self.assertIn("no compatibility action is needed", body)
        self.assertEqual(body.count(upstream_watch.STATUS_MARKER), 1)

    @patch.object(upstream_watch, "run_gh")
    def test_changed_tag_updates_the_marked_comment_in_place(self, run_gh):
        old_body = upstream_watch.render_status(self.status())
        run_gh.return_value = json.dumps([{"id": 41, "body": old_body}])

        new_body = upstream_watch.render_status(self.status(seerr="v3.6.0"))
        self.assertEqual(upstream_watch.reconcile_comment(new_body), "updated")
        self.assertEqual(run_gh.call_count, 2)
        self.assertIn("Verification needed", new_body)
        self.assertIn("v3.6.0", new_body)
        self.assertIn("OpenAPI specification", new_body)
        self.assertIn("issues/comments/41", run_gh.call_args.args[0][1])
        self.assertIn("PATCH", run_gh.call_args.args[0])

    @patch.object(upstream_watch, "run_gh")
    def test_successful_verification_transitions_existing_comment_to_current(self, run_gh):
        old_body = upstream_watch.render_status(self.status(seerr="v3.6.0"))
        run_gh.return_value = json.dumps([{"id": 41, "body": old_body}])

        current_body = upstream_watch.render_status(self.status())
        self.assertEqual(upstream_watch.reconcile_comment(current_body), "updated")
        self.assertIn("no compatibility action is needed", current_body)
        self.assertNotIn("Verification needed", current_body)

    @patch.object(upstream_watch.subprocess, "run")
    def test_release_api_failure_and_empty_tag_are_errors(self, run):
        run.side_effect = upstream_watch.subprocess.CalledProcessError(1, ["gh"])
        with self.assertRaises(upstream_watch.UpstreamCheckError):
            upstream_watch.fetch_latest_tag("seerr-team/seerr")

        run.side_effect = None
        run.return_value = upstream_watch.subprocess.CompletedProcess(["gh"], 0, "", "")
        with self.assertRaises(upstream_watch.UpstreamCheckError):
            upstream_watch.fetch_latest_tag("seerr-team/seerr")

        run.return_value = upstream_watch.subprocess.CompletedProcess(["gh"], 0, "null\n", "")
        with self.assertRaises(upstream_watch.UpstreamCheckError):
            upstream_watch.fetch_latest_tag("seerr-team/seerr")

    @patch.object(upstream_watch, "fetch_latest_tag")
    @patch.object(upstream_watch, "load_verified")
    def test_legacy_jellyseerr_redirect_is_not_queried(self, load_verified, fetch_latest_tag):
        load_verified.return_value = {
            "seerr": "v3.5.0",
            "jellyseerr": "v2.7.3",
            "overseerr": "v1.35.0",
        }
        fetch_latest_tag.side_effect = ["v3.5.0", "v1.35.0"]

        result = upstream_watch.check_releases()

        self.assertEqual(result.latest["jellyseerr"], "v2.7.3")
        self.assertEqual(
            fetch_latest_tag.call_args_list,
            [call("seerr-team/seerr"), call("sct/overseerr")],
        )

    @patch.object(upstream_watch.subprocess, "run")
    def test_api_error_does_not_read_or_create_status_comment(self, run):
        run.side_effect = upstream_watch.subprocess.CalledProcessError(1, ["gh"])
        with self.assertRaises(upstream_watch.UpstreamCheckError):
            upstream_watch.main()
        self.assertEqual(run.call_count, 1)


if __name__ == "__main__":
    unittest.main()
