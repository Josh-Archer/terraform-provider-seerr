import io
import json
import pathlib
import re
import sqlite3
import subprocess
import tarfile
import tempfile
import unittest
from unittest.mock import patch

from scripts import ci_docker_helpers as h

ROOT = pathlib.Path(__file__).resolve().parent.parent
DIGEST_RE = re.compile(r"@sha256:[0-9a-f]{64}$")
VAR_DEFAULT_RE = re.compile(r"^\$\{([A-Za-z0-9_]+):-([^}]+)\}$")


def validate_compose_images(compose_text: str) -> list[str]:
    """Ensure mock compose services do not use unpinned python:3.12-alpine pulls."""
    errors: list[str] = []
    current_service: str | None = None
    in_services = False

    for raw_line in compose_text.splitlines():
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue
        if not raw_line.startswith(" "):
            in_services = raw_line.strip() == "services:"
            current_service = None
            continue
        if not in_services:
            continue

        service_match = re.match(r"^  ([A-Za-z0-9_.-]+):\s*$", raw_line)
        if service_match:
            current_service = service_match.group(1)
            continue

        image_match = re.match(r"^\s+image:\s*(\S+)\s*$", raw_line)
        if image_match and current_service is not None:
            image_spec = image_match.group(1).strip("\"'")
            var_match = VAR_DEFAULT_RE.match(image_spec)
            default_ref = var_match.group(2) if var_match else image_spec

            if current_service.endswith("-mock"):
                if var_match is None or var_match.group(1) != "SEERR_MOCK_PYTHON_IMAGE":
                    errors.append(
                        f"{current_service}: must reference ${{SEERR_MOCK_PYTHON_IMAGE:-...}}, got {image_spec}"
                    )
                if not DIGEST_RE.search(default_ref):
                    errors.append(
                        f"{current_service}: fallback image must be pinned by @sha256 digest, got {default_ref}"
                    )
                if default_ref.startswith("python:"):
                    errors.append(
                        f"{current_service}: fallback image must not default to unpinned Docker Hub {default_ref}"
                    )
            elif "python:" in default_ref and not DIGEST_RE.search(default_ref):
                errors.append(
                    f"{current_service}: python image must be pinned by @sha256 digest, got {default_ref}"
                )

    return errors


def validate_no_container_apk(script_text: str, name: str) -> list[str]:
    """Ensure CI bootstrap scripts do not run apk inside the Seerr container."""
    errors: list[str] = []
    for lineno, line in enumerate(script_text.splitlines(), start=1):
        stripped = line.strip()
        if stripped.startswith("#"):
            continue
        if re.search(r"\bapk\s+add\b", line) or re.search(r"\bapk\s+--", line):
            errors.append(f"{name}:{lineno}: forbidden runtime apk invocation: {stripped}")
    return errors


def _create_seerr_config(
    config_dir: pathlib.Path,
    *,
    wal_mode: bool = True,
    with_existing_user: bool = False,
) -> None:
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


class ComposeAndScriptRegressionTests(unittest.TestCase):
    def test_repository_compose_pins_mock_python_and_avoids_docker_hub_tag(self):
        compose_text = (ROOT / "tests" / "docker-compose.yml").read_text(encoding="utf-8")
        self.assertEqual(validate_compose_images(compose_text), [])
        self.assertIn(h.PINNED_MOCK_PYTHON_DIGEST, compose_text)

    def test_compose_validator_rejects_unpinned_python_alpine(self):
        bad_plain = "services:\n  plex-mock:\n    image: python:3.12-alpine\n"
        self.assertTrue(validate_compose_images(bad_plain))

        bad_env = "services:\n  plex-mock:\n    image: ${SEERR_MOCK_PYTHON_IMAGE:-python:3.12-alpine}\n"
        self.assertTrue(validate_compose_images(bad_env))

    def test_ci_docker_up_and_helpers_never_run_apk_add(self):
        up_script = (ROOT / "scripts" / "ci-docker-up.sh").read_text(encoding="utf-8")
        helpers = (ROOT / "scripts" / "ci_docker_helpers.py").read_text(encoding="utf-8")
        self.assertEqual(validate_no_container_apk(up_script, "scripts/ci-docker-up.sh"), [])
        self.assertEqual(validate_no_container_apk(helpers, "scripts/ci_docker_helpers.py"), [])
        self.assertTrue(validate_no_container_apk("until apk add --no-cache python3 sqlite; do", "bad.sh"))

    def test_workflow_prepares_cached_mock_image_for_all_compose_jobs(self):
        workflow_text = (ROOT / ".github" / "workflows" / "test.yml").read_text(encoding="utf-8")
        self.assertIn("test_ci_docker_setup.py", workflow_text)
        # All 5 compose-backed jobs (tofu-test-stable, terraform-test-stable,
        # consumer-replay, tofu-test-stable-trusted, tofu-test-full) must prepare the mock image.
        self.assertEqual(
            workflow_text.count("uses: ./.github/actions/ensure-mock-python-image"),
            5,
        )


class DockerPullAndCacheTests(unittest.TestCase):
    def test_pull_retries_transient_short_read_and_verifies_digest(self):
        attempts = [
            subprocess.CompletedProcess(
                ["docker", "pull"],
                1,
                "short read: expected 3853673 bytes but got 0: unexpected EOF",
                None,
            ),
            subprocess.CompletedProcess(["docker", "pull"], 0, "Status: Downloaded", None),
        ]
        ref = f"mirror.gcr.io/library/python:3.12-alpine@{h.PINNED_MOCK_PYTHON_DIGEST}"
        with (
            patch.object(h.subprocess, "run", side_effect=attempts) as mock_run,
            patch.object(h.time, "sleep") as mock_sleep,
            patch.object(
                h,
                "inspect_local_image",
                return_value=("sha256:abc", [f"mirror.gcr.io/library/python@{h.PINNED_MOCK_PYTHON_DIGEST}"]),
            ),
        ):
            image_id = h.pull_image_with_verification(ref, max_attempts=3, retry_delay=1.0)
        self.assertEqual(image_id, "sha256:abc")
        self.assertEqual(mock_run.call_count, 2)
        mock_sleep.assert_called_once_with(1.0)

    def test_pull_exhausting_retries_reports_short_read_error(self):
        err = subprocess.CompletedProcess(
            ["docker", "pull"],
            1,
            "short read: expected 3853673 bytes but got 0: unexpected EOF",
            None,
        )
        ref = f"mirror.gcr.io/library/python:3.12-alpine@{h.PINNED_MOCK_PYTHON_DIGEST}"
        with (
            patch.object(h.subprocess, "run", side_effect=[err, err, err]),
            patch.object(h.time, "sleep"),
        ):
            with self.assertRaisesRegex(h.DockerCIError, "short read.*unexpected EOF"):
                h.pull_image_with_verification(ref, max_attempts=3, retry_delay=0.1)

    def test_pull_does_not_retry_or_hide_checksum_mismatch(self):
        mismatch = subprocess.CompletedProcess(
            ["docker", "pull"],
            1,
            "filesystem layer verification failed for digest sha256:881fc932: checksum mismatch",
            None,
        )
        ref = f"mirror.gcr.io/library/python:3.12-alpine@{h.PINNED_MOCK_PYTHON_DIGEST}"
        with (
            patch.object(h.subprocess, "run", return_value=mismatch) as mock_run,
            patch.object(h.time, "sleep") as mock_sleep,
        ):
            with self.assertRaises(h.DockerCIError) as ctx:
                h.pull_image_with_verification(ref, max_attempts=3, retry_delay=1.0)
        self.assertTrue(ctx.exception.checksum_mismatch)
        self.assertIn("checksum mismatch", str(ctx.exception))
        self.assertEqual(mock_run.call_count, 1)
        mock_sleep.assert_not_called()

    def test_verify_local_image_rejects_missing_expected_digest(self):
        with patch.object(
            h,
            "inspect_local_image",
            return_value=("sha256:abc", ["mirror.gcr.io/library/python@sha256:" + ("0" * 64)]),
        ):
            with self.assertRaises(h.DockerCIError) as ctx:
                h.verify_local_image(
                    f"mirror.gcr.io/library/python:3.12-alpine@{h.PINNED_MOCK_PYTHON_DIGEST}",
                    expected_digest=h.PINNED_MOCK_PYTHON_DIGEST,
                )
        self.assertTrue(ctx.exception.checksum_mismatch)

    def test_ensure_mock_python_image_loads_valid_cache_without_pull(self):
        with tempfile.TemporaryDirectory() as tmp:
            cache_dir = pathlib.Path(tmp)
            tar_path = cache_dir / "python-3.12-alpine.tar"
            tar_path.write_bytes(b"fake-docker-tar")
            (cache_dir / "python-3.12-alpine.tar.sha256").write_text(
                h.sha256_file(tar_path) + "\n", encoding="utf-8"
            )
            (cache_dir / "python-3.12-alpine.image-id").write_text("sha256:goodid\n", encoding="utf-8")

            with (
                patch.object(h, "run_cmd") as mock_run_cmd,
                patch.object(h, "inspect_local_image", return_value=("sha256:goodid", [])),
                patch.object(h, "verify_mock_python_runtime") as mock_verify,
                patch.object(h, "pull_image_with_verification") as mock_pull,
            ):
                resolved = h.ensure_mock_python_image(cache_dir=cache_dir)

            self.assertEqual(resolved, h.LOCAL_CACHED_MOCK_PYTHON_TAG)
            mock_pull.assert_not_called()
            mock_verify.assert_called_once_with(h.LOCAL_CACHED_MOCK_PYTHON_TAG)
            mock_run_cmd.assert_called_once()
            self.assertEqual(mock_run_cmd.call_args.args[0][:3], ["docker", "load", "-i"])

    def test_ensure_mock_python_image_fails_loudly_on_corrupt_cache(self):
        with tempfile.TemporaryDirectory() as tmp:
            cache_dir = pathlib.Path(tmp)
            (cache_dir / "python-3.12-alpine.tar").write_bytes(b"corrupted-bytes")
            (cache_dir / "python-3.12-alpine.tar.sha256").write_text("0" * 64 + "\n", encoding="utf-8")

            with self.assertRaises(h.DockerCIError) as ctx:
                h.ensure_mock_python_image(cache_dir=cache_dir)
            self.assertTrue(ctx.exception.checksum_mismatch)
            self.assertIn("checksum mismatch", str(ctx.exception))

    def test_ensure_mock_python_image_falls_back_on_transient_error_but_not_checksum_mismatch(self):
        c1 = f"mirror.gcr.io/library/python:3.12-alpine@{h.PINNED_MOCK_PYTHON_DIGEST}"
        c2 = f"public.ecr.aws/docker/library/python:3.12-alpine@{h.PINNED_MOCK_PYTHON_DIGEST}"

        with (
            patch.object(h, "inspect_local_image", return_value=None),
            patch.object(
                h,
                "pull_image_with_verification",
                side_effect=[h.DockerCIError("short read: unexpected EOF"), "sha256:good"],
            ) as mock_pull,
            patch.object(h, "run_cmd") as mock_run_cmd,
            patch.object(h, "verify_mock_python_runtime"),
        ):
            resolved = h.ensure_mock_python_image(candidates=(c1, c2))
        self.assertEqual(resolved, h.LOCAL_CACHED_MOCK_PYTHON_TAG)
        self.assertEqual(mock_pull.call_count, 2)
        mock_run_cmd.assert_called_once_with(
            ["docker", "tag", c2, h.LOCAL_CACHED_MOCK_PYTHON_TAG],
            timeout=h.DOCKER_IO_TIMEOUT_SECONDS,
        )

        with (
            patch.object(h, "inspect_local_image", return_value=None),
            patch.object(
                h,
                "pull_image_with_verification",
                side_effect=h.DockerCIError("checksum mismatch", checksum_mismatch=True),
            ) as mock_pull_bad,
        ):
            with self.assertRaises(h.DockerCIError) as ctx:
                h.ensure_mock_python_image(candidates=(c1, c2))
        self.assertTrue(ctx.exception.checksum_mismatch)
        self.assertEqual(mock_pull_bad.call_count, 1)

    def test_ensure_mock_python_image_rejects_unpinned_requested_image(self):
        with patch.object(h, "inspect_local_image", return_value=None):
            with self.assertRaisesRegex(h.DockerCIError, "unpinned SEERR_MOCK_PYTHON_IMAGE"):
                h.ensure_mock_python_image(requested_image="python:3.12-alpine")


class RunnerSeedingTests(unittest.TestCase):
    def test_seed_config_dir_populates_settings_and_sqlite_from_wal(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_dir = pathlib.Path(tmp) / "config"
            _create_seerr_config(config_dir, wal_mode=True, with_existing_user=False)

            h.seed_config_dir(
                config_dir,
                api_key="test-api-key",
                admin_email="ci-admin@example.invalid",
                admin_username="ci-admin",
                admin_avatar="/logo_full.svg",
                local_port=15055,
            )

            settings = json.loads((config_dir / "settings.json").read_text(encoding="utf-8"))
            self.assertEqual(settings["main"]["apiKey"], "test-api-key")
            self.assertTrue(settings["public"]["initialized"])
            self.assertTrue(settings["notifications"]["email"]["enabled"])
            self.assertEqual(settings["notifications"]["email"]["options"]["emailUrlBase"], "http://127.0.0.1:15055")
            self.assertFalse((config_dir / "db" / "db.sqlite3-wal").exists())

            db = sqlite3.connect(config_dir / "db" / "db.sqlite3")
            try:
                user = db.execute('SELECT id, email, username, permissions FROM "user" WHERE id=1').fetchone()
                self.assertEqual(user, (1, "ci-admin@example.invalid", "ci-admin", 2))
                user_settings = db.execute(
                    "SELECT userId, locale, watchlistSyncMovies, watchlistSyncTv FROM user_settings WHERE userId=1"
                ).fetchone()
                self.assertEqual(user_settings, (1, "en", 0, 0))
            finally:
                db.close()

    def test_seed_config_dir_updates_existing_user_and_requires_db(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_dir = pathlib.Path(tmp) / "config"
            _create_seerr_config(config_dir, wal_mode=False, with_existing_user=True)
            h.seed_config_dir(
                config_dir,
                api_key="new-key",
                admin_email="ci-admin@example.invalid",
                admin_username="ci-admin",
                admin_avatar="/logo_full.svg",
                local_port=15055,
            )
            db = sqlite3.connect(config_dir / "db" / "db.sqlite3")
            try:
                user = db.execute('SELECT email, username, permissions FROM "user" WHERE id=1').fetchone()
                self.assertEqual(user, ("ci-admin@example.invalid", "ci-admin", 2))
            finally:
                db.close()

            missing_dir = pathlib.Path(tmp) / "missing"
            with self.assertRaisesRegex(h.DockerCIError, "did not create"):
                h.seed_config_dir(
                    missing_dir,
                    api_key="k",
                    admin_email="e",
                    admin_username="u",
                    admin_avatar="/a",
                    local_port=15055,
                )

    def test_build_config_tar_sets_seerr_uid_gid_and_mode(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_dir = pathlib.Path(tmp) / "config"
            _create_seerr_config(config_dir, wal_mode=False)
            h.seed_config_dir(
                config_dir,
                api_key="test-api-key",
                admin_email="ci-admin@example.invalid",
                admin_username="ci-admin",
                admin_avatar="/logo_full.svg",
                local_port=15055,
            )

            tar_bytes = h.build_config_tar(config_dir)
            with tarfile.open(fileobj=io.BytesIO(tar_bytes), mode="r") as archive:
                members = {m.name: m for m in archive.getmembers()}
                self.assertEqual(set(members), {"settings.json", "db/db.sqlite3"})
                for name, member in members.items():
                    self.assertEqual(member.uid, h.SEERR_UID)
                    self.assertEqual(member.gid, h.SEERR_GID)
                    self.assertEqual(member.uname, "node")
                    self.assertEqual(member.gname, "node")
                    self.assertEqual(member.mode, 0o644)
                    extracted = archive.extractfile(member)
                    self.assertIsNotNone(extracted)
                    self.assertEqual(extracted.read(), (config_dir / name).read_bytes())

    def test_seed_seerr_container_orchestrates_runner_seed_without_apk_add(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = pathlib.Path(tmp)
            compose = ["docker", "compose"]
            recorded = []

            def fake_run(command, *, input_bytes=None, **kwargs):
                recorded.append((command, input_bytes, kwargs))
                if command[:2] == ["docker", "cp"] and command[2] != "-":
                    _create_seerr_config(pathlib.Path(command[3]))
                return subprocess.CompletedProcess(command, 0, "", "")

            with patch.object(h, "run_cmd", side_effect=fake_run):
                h.seed_seerr_container(
                    compose,
                    "seerr-test",
                    work_dir,
                    compose_cwd=work_dir,
                    api_key="test-api-key",
                    admin_email="ci-admin@example.invalid",
                    admin_username="ci-admin",
                    admin_avatar="/logo_full.svg",
                    local_port=15055,
                )

            commands = [entry[0] for entry in recorded]
            self.assertEqual(commands[0], compose + ["stop", "seerr"])
            self.assertEqual(commands[1], ["docker", "cp", "seerr-test:/app/config", str(work_dir / "seed-config")])
            self.assertEqual(
                commands[2],
                compose
                + [
                    "run",
                    "--rm",
                    "--no-deps",
                    "-T",
                    "--entrypoint",
                    "sh",
                    "seerr",
                    "-c",
                    "rm -f /app/config/db/db.sqlite3-wal /app/config/db/db.sqlite3-shm",
                ],
            )
            self.assertEqual(commands[3], ["docker", "cp", "-", "seerr-test:/app/config"])
            self.assertIsInstance(recorded[3][1], bytes)
            self.assertGreater(len(recorded[3][1]), 0)
            self.assertEqual(commands[4], compose + ["start", "seerr"])
            for command, _, kwargs in recorded:
                self.assertNotIn("apk", " ".join(command))
                self.assertEqual(kwargs.get("timeout"), h.DOCKER_IO_TIMEOUT_SECONDS)


if __name__ == "__main__":
    unittest.main()
