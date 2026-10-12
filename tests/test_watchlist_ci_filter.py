import unittest
from unittest.mock import patch

from scripts.watchlist_ci_filter import pr_paths, should_run


class WatchlistCIFilterTests(unittest.TestCase):
    def test_main_push_always_runs_even_for_documentation(self):
        self.assertTrue(should_run("push", ["README.md"]))
        self.assertTrue(should_run("push", []))

    def test_documentation_only_pr_skips_integration(self):
        self.assertFalse(should_run("pull_request", ["README.md", "docs/resources/watchlist.md", "templates/guides/watchlist-live-integration.md"]))

    def test_provider_and_harness_changes_run_integration(self):
        for path in ["internal/provider/client.go", "internal/provider/watchlist_api.go", "main.go", "go.mod", "go.sum", "scripts/test_watchlist_live.py", "scripts/watchlist_ci_filter.py", "tests/docker-compose.watchlist-live.yml", ".github/workflows/test.yml"]:
            with self.subTest(path=path):
                self.assertTrue(should_run("pull_request", ["README.md", path]))

    def test_similar_unrelated_paths_do_not_match(self):
        self.assertFalse(should_run("pull_request", ["scripts-docs.md", "tests-notes.md", "internal/provider-notes.md", ".github/workflows/release.yml"]))

    @patch("scripts.watchlist_ci_filter.subprocess.check_output")
    def test_comparison_keeps_renamed_old_paths_and_unusual_names(self, output):
        output.side_effect = ["base head", b"docs/moved.go\0internal/provider/deleted.go\0docs/odd\nname.md\0"]
        paths = pr_paths()
        self.assertEqual(paths, ["docs/moved.go", "internal/provider/deleted.go", "docs/odd\nname.md"])
        self.assertTrue(should_run("pull_request", paths))
        self.assertIn("--no-renames", output.call_args.args[0])
        self.assertEqual(output.call_args.args[0][-2:], ["HEAD^1", "HEAD"])

    @patch("scripts.watchlist_ci_filter.subprocess.check_output", return_value="one-parent")
    def test_unexpected_checkout_fails_instead_of_silently_skipping(self, output):
        with self.assertRaisesRegex(RuntimeError, "two-parent"):
            pr_paths()


if __name__ == "__main__":
    unittest.main()
