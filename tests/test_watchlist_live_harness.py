import io
import json
import pathlib
import sqlite3
import subprocess
import tarfile
import tempfile
import unittest
from unittest.mock import patch

from scripts import test_watchlist_live as w


def _create_seerr_config(config_dir: pathlib.Path, *, wal_mode: bool = True,
                         with_existing_user: bool = False,
                         include_timestamps: bool = True) -> None:
    config_dir.mkdir(parents=True, exist_ok=True)
    db_dir = config_dir / "db"
    db_dir.mkdir(parents=True, exist_ok=True)

    (config_dir / "settings.json").write_text(
        json.dumps({"main": {}, "public": {"initialized": False}}),
        encoding="utf-8",
    )

    db_path = db_dir / "db.sqlite3"
    db = sqlite3.connect(db_path)
    if wal_mode:
        db.execute("PRAGMA journal_mode=WAL")
    db.execute(
        'CREATE TABLE "user" ('
        "id INTEGER PRIMARY KEY, email TEXT, username TEXT, "
        "permissions INTEGER, avatar TEXT, userType INTEGER)"
    )
    db.execute(
        "CREATE TABLE user_settings ("
        "userId INTEGER PRIMARY KEY, locale TEXT, "
        "watchlistSyncMovies INTEGER, watchlistSyncTv INTEGER)"
    )
    timestamp_cols = ", createdAt TEXT, updatedAt TEXT" if include_timestamps else ""
    db.execute(
        f"CREATE TABLE watchlist ("
        f"id INTEGER PRIMARY KEY AUTOINCREMENT, tmdbId INTEGER, "
        f"mediaType TEXT, title TEXT, ratingKey TEXT, requestedById INTEGER{timestamp_cols})"
    )
    if with_existing_user:
        db.execute(
            'INSERT INTO "user" (id, email, username, permissions, avatar, userType) '
            "VALUES (1, 'old@example.invalid', 'old', 0, '', 1)"
        )
        db.execute(
            "INSERT INTO user_settings (userId, locale, watchlistSyncMovies, watchlistSyncTv) "
            "VALUES (1, 'fr', 1, 1)"
        )
    db.commit()
    db.close()


class WatchlistLiveTests(unittest.TestCase):
    def test_script_does_not_install_packages_in_container(self):
        source = (w.ROOT / "scripts" / "test_watchlist_live.py").read_text(encoding="utf-8")
        self.assertNotIn("apk add", source)
        self.assertNotIn("apk ", source)

    def test_seed_config_dir_populates_settings_and_42_watchlist_rows_from_wal(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_dir = pathlib.Path(tmp) / "config"
            _create_seerr_config(config_dir, wal_mode=True, with_existing_user=False)

            w.seed_config_dir(config_dir)

            settings = json.loads((config_dir / "settings.json").read_text(encoding="utf-8"))
            self.assertEqual(settings["main"]["apiKey"], w.API_KEY)
            self.assertTrue(settings["public"]["initialized"])

            # Verify WAL was checkpointed into db.sqlite3 itself.
            self.assertFalse((config_dir / "db" / "db.sqlite3-wal").exists())

            db = sqlite3.connect(config_dir / "db" / "db.sqlite3")
            try:
                user = db.execute('SELECT id, email, username, permissions FROM "user" WHERE id=1').fetchone()
                self.assertEqual(user, (1, "watchlist-ci@example.invalid", "watchlist-ci-admin", 2))
                user_settings = db.execute(
                    "SELECT userId, locale, watchlistSyncMovies, watchlistSyncTv FROM user_settings WHERE userId=1"
                ).fetchone()
                self.assertEqual(user_settings, (1, "en", 0, 0))
                rows = db.execute(
                    "SELECT tmdbId, mediaType, requestedById FROM watchlist ORDER BY tmdbId"
                ).fetchall()
                self.assertEqual(len(rows), w.FIXTURE_COUNT)
                self.assertEqual(rows[0], (w.FIRST_TMDB_ID, "movie", 1))
                self.assertEqual(rows[-1], (w.FIRST_TMDB_ID + w.FIXTURE_COUNT - 1, "movie", 1))
            finally:
                db.close()

    def test_seed_config_dir_updates_existing_user_and_rejects_tmdb_collision(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_dir = pathlib.Path(tmp) / "config"
            _create_seerr_config(config_dir, wal_mode=False, with_existing_user=True, include_timestamps=False)
            w.seed_config_dir(config_dir)

            with self.assertRaisesRegex(w.IntegrationFailure, "fixture TMDB IDs already exist"):
                w.seed_config_dir(config_dir)

    def test_build_config_tar_sets_seerr_uid_gid_and_mode(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_dir = pathlib.Path(tmp) / "config"
            _create_seerr_config(config_dir, wal_mode=False)
            w.seed_config_dir(config_dir)

            tar_bytes = w.build_config_tar(config_dir)

            with tarfile.open(fileobj=io.BytesIO(tar_bytes), mode="r") as archive:
                members = {m.name: m for m in archive.getmembers()}
                self.assertEqual(set(members), {"settings.json", "db/db.sqlite3"})
                for name, member in members.items():
                    self.assertEqual(member.uid, w.SEERR_UID)
                    self.assertEqual(member.gid, w.SEERR_GID)
                    self.assertEqual(member.uname, "node")
                    self.assertEqual(member.gname, "node")
                    self.assertEqual(member.mode, 0o644)
                    extracted = archive.extractfile(member)
                    self.assertIsNotNone(extracted)
                    self.assertEqual(extracted.read(), (config_dir / name).read_bytes())

    def test_seed_container_database_stops_copies_seeds_and_streams_tar_back(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = pathlib.Path(tmp)
            compose = ["docker", "compose", "-p", "proj", "-f", "compose.yml"]
            recorded = []

            def fake_run(command, *, input_bytes=None, **kwargs):
                recorded.append((command, input_bytes, kwargs))
                if command[:2] == ["docker", "cp"] and command[2] != "-":
                    _create_seerr_config(pathlib.Path(command[3]))
                return subprocess.CompletedProcess(command, 0, "", "")

            with patch.object(w, "run", side_effect=fake_run):
                w.seed_container_database(compose, "cid-123", work_dir)

            commands = [entry[0] for entry in recorded]
            self.assertEqual(commands[0], compose + ["stop", "seerr"])
            self.assertEqual(commands[1], ["docker", "cp", "cid-123:/app/config", str(work_dir / "seed-config")])
            self.assertEqual(
                commands[2],
                compose + [
                    "run", "--rm", "--no-deps", "-T", "--entrypoint", "sh", "seerr",
                    "-c", "rm -f /app/config/db/db.sqlite3-wal /app/config/db/db.sqlite3-shm",
                ],
            )
            self.assertEqual(commands[3], ["docker", "cp", "-", "cid-123:/app/config"])
            self.assertIsInstance(recorded[3][1], bytes)
            self.assertGreater(len(recorded[3][1]), 0)
            for _, _, kwargs in recorded:
                self.assertEqual(kwargs.get("timeout"), w.DOCKER_IO_TIMEOUT_SECONDS)

    def test_tail_lines_returns_at_most_50_lines(self):
        lines = "\n".join(f"line {i}" for i in range(1, 81))
        tailed = w.tail_lines(lines, 50).splitlines()
        self.assertEqual(len(tailed), 50)
        self.assertEqual(tailed[0], "line 31")
        self.assertEqual(tailed[-1], "line 80")
        self.assertEqual(w.tail_lines("", 50), "")

    def test_capture_failure_diagnostics_writes_full_logs_and_prints_last_50_lines(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = pathlib.Path(tmp)
            compose = ["docker", "compose", "-p", "proj", "-f", "compose.yml"]
            container_log = "\n".join(f"seerr-log-{i}" for i in range(1, 76)) + "\n"
            compose_log = "compose-log-line\n"
            stderr = io.StringIO()

            def fake_subprocess_run(command, **kwargs):
                self.assertEqual(kwargs.get("timeout"), w.LOG_CAPTURE_TIMEOUT_SECONDS)
                if command[:2] == ["docker", "logs"]:
                    return subprocess.CompletedProcess(command, 0, container_log, None)
                return subprocess.CompletedProcess(command, 0, compose_log, None)

            with patch.object(w.subprocess, "run", side_effect=fake_subprocess_run):
                w.capture_failure_diagnostics(compose, "cid-123", work_dir, stream=stderr)

            self.assertEqual((work_dir / "seerr-container.log").read_text(encoding="utf-8"), container_log)
            self.assertEqual((work_dir / "compose.log").read_text(encoding="utf-8"), compose_log)
            printed = stderr.getvalue()
            self.assertIn("Last 50 lines of Seerr container logs", printed)
            self.assertNotIn("seerr-log-25\n", printed)
            self.assertIn("seerr-log-26\n", printed)
            self.assertIn("seerr-log-75\n", printed)
            self.assertIn(f"Diagnostics saved under {work_dir}", printed)

    def test_wait_for_server_and_auth_have_bounded_timeouts(self):
        with patch.object(w.urllib.request, "urlopen", side_effect=OSError("refused")):
            with patch.object(w.time, "monotonic", side_effect=[0.0, 1.0, 2.0]), patch.object(w.time, "sleep"):
                with self.assertRaisesRegex(w.IntegrationFailure, "did not become ready"):
                    w.wait_for_server("http://127.0.0.1:5055", "cid-123", timeout_seconds=1.5)

        with patch.object(w, "api_request", side_effect=OSError("refused")):
            with patch.object(w.time, "monotonic", side_effect=[0.0, 1.0, 2.0]), patch.object(w.time, "sleep"):
                with self.assertRaisesRegex(w.IntegrationFailure, "did not authenticate"):
                    w.wait_for_auth("http://127.0.0.1:5055", timeout_seconds=1.5)


if __name__ == "__main__":
    unittest.main()
