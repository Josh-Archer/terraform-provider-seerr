#!/usr/bin/env python3
"""Helpers for CI Docker image pre-pull/caching and runner-side Seerr seeding."""

from __future__ import annotations

import argparse
import hashlib
import io
import json
import os
import pathlib
import re
import shlex
import shutil
import sqlite3
import subprocess
import sys
import tarfile
import tempfile
import time

ROOT = pathlib.Path(__file__).resolve().parent.parent

# Multi-arch OCI index digest for python:3.12-alpine (3.12.16-alpine3.24).
PINNED_MOCK_PYTHON_DIGEST = "sha256:c5ba3deef00c7758561a0f03c26f44f10589263d40e732b8748e0d1f38f6dd89"
DEFAULT_MOCK_PYTHON_IMAGE = f"mirror.gcr.io/library/python:3.12-alpine@{PINNED_MOCK_PYTHON_DIGEST}"
LOCAL_CACHED_MOCK_PYTHON_TAG = "seerr-ci/python:3.12-alpine"
MOCK_PYTHON_MIRRORS = (
    f"mirror.gcr.io/library/python:3.12-alpine@{PINNED_MOCK_PYTHON_DIGEST}",
    f"public.ecr.aws/docker/library/python:3.12-alpine@{PINNED_MOCK_PYTHON_DIGEST}",
    f"docker.io/library/python:3.12-alpine@{PINNED_MOCK_PYTHON_DIGEST}",
)

DOCKER_IO_TIMEOUT_SECONDS = 60
DOCKER_PULL_TIMEOUT_SECONDS = 300
DEFAULT_PULL_ATTEMPTS = 3
DEFAULT_RETRY_DELAY_SECONDS = 2.0

SEERR_UID = 1000
SEERR_GID = 1000

CHECKSUM_MISMATCH_PATTERNS = (
    "checksum mismatch",
    "digest mismatch",
    "invalid checksum",
    "failed to verify",
    "size mismatch",
    "verification failed",
)

TRANSIENT_PULL_PATTERNS = (
    "short read",
    "unexpected eof",
    "eof",
    "timeout",
    "timed out",
    "context deadline exceeded",
    "connection reset",
    "connection refused",
    "broken pipe",
    "temporary failure",
    "try again",
    "toomanyrequests",
    "502",
    "503",
    "504",
    "429",
)

DIGEST_RE = re.compile(r"@(sha256:[0-9a-f]{64})$")


class DockerCIError(RuntimeError):
    """Raised when Docker image preparation or runner-side seeding fails."""

    def __init__(self, message: str, *, checksum_mismatch: bool = False) -> None:
        super().__init__(message)
        self.checksum_mismatch = checksum_mismatch


def run_cmd(
    command: list[str],
    *,
    cwd: pathlib.Path | None = None,
    input_bytes: bytes | None = None,
    capture: bool = False,
    allowed_codes: tuple[int, ...] = (0,),
    timeout: int = DOCKER_IO_TIMEOUT_SECONDS,
) -> subprocess.CompletedProcess:
    """Run a subprocess with a bounded timeout and fail clearly on unexpected exit."""
    print("+", " ".join(command), file=sys.stderr, flush=True)
    use_text = input_bytes is None
    result = subprocess.run(
        command,
        cwd=cwd,
        input=input_bytes,
        text=use_text,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.STDOUT if capture else None,
        check=False,
        timeout=timeout,
    )
    if result.returncode not in allowed_codes:
        output = ""
        if capture and result.stdout is not None:
            output = (
                result.stdout
                if isinstance(result.stdout, str)
                else result.stdout.decode("utf-8", errors="replace")
            )
        raise DockerCIError(
            f"command exited {result.returncode}: {' '.join(command)}"
            + (f"\n{output.strip()}" if output.strip() else "")
        )
    return result


def is_checksum_mismatch(output: str) -> bool:
    """Return True if output indicates an integrity/checksum/digest verification failure."""
    lowered = output.lower()
    if any(pattern in lowered for pattern in CHECKSUM_MISMATCH_PATTERNS):
        return True
    if "sha256" in lowered and ("mismatch" in lowered or "does not match" in lowered):
        return True
    return False


def is_transient_pull_error(output: str) -> bool:
    """Return True if pull output looks like a transient network or short-read failure."""
    if is_checksum_mismatch(output):
        return False
    lowered = output.lower()
    return any(pattern in lowered for pattern in TRANSIENT_PULL_PATTERNS)


def sha256_file(path: pathlib.Path) -> str:
    """Compute hex SHA-256 digest of a file."""
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        while True:
            chunk = handle.read(1024 * 1024)
            if not chunk:
                break
            digest.update(chunk)
    return digest.hexdigest()


def extract_digest(image_ref: str) -> str | None:
    """Return the pinned sha256:... digest from an image reference, if present."""
    match = DIGEST_RE.search(image_ref.strip())
    return match.group(1) if match else None


def inspect_local_image(image_ref: str) -> tuple[str, list[str]] | None:
    """Return (image_id, repo_digests) if image_ref is present locally, else None."""
    result = subprocess.run(
        ["docker", "image", "inspect", "--format", "{{.Id}}\t{{json .RepoDigests}}", image_ref],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )
    if result.returncode != 0:
        return None
    line = result.stdout.strip()
    if not line:
        return None
    image_id, _, raw_digests = line.partition("\t")
    try:
        repo_digests = json.loads(raw_digests or "[]")
    except json.JSONDecodeError:
        repo_digests = []
    if not isinstance(repo_digests, list):
        repo_digests = []
    return image_id.strip(), [str(d) for d in repo_digests]


def verify_local_image(image_ref: str, *, expected_digest: str | None = None) -> str:
    """Verify image_ref exists locally and, when expected_digest is set, matches RepoDigests."""
    inspected = inspect_local_image(image_ref)
    if inspected is None:
        raise DockerCIError(f"Docker image {image_ref} is not present locally after pull/load")
    image_id, repo_digests = inspected
    if not image_id:
        raise DockerCIError(f"Docker image {image_ref} has an empty image ID")
    if expected_digest is not None:
        expected_suffix = f"@{expected_digest}"
        if not any(d.endswith(expected_suffix) for d in repo_digests):
            raise DockerCIError(
                f"Digest mismatch for {image_ref}: expected {expected_digest} in RepoDigests, "
                f"got {repo_digests}",
                checksum_mismatch=True,
            )
    return image_id


def verify_mock_python_runtime(image_ref: str) -> None:
    """Verify the mock Python image can run the stdlib HTTP server used by tests/fixtures."""
    run_cmd(
        [
            "docker",
            "run",
            "--rm",
            image_ref,
            "python3",
            "-c",
            "import http.server, json, os; print('ok')",
        ],
        capture=True,
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )


def pull_image_with_verification(
    image_ref: str,
    *,
    expected_digest: str | None = None,
    max_attempts: int = DEFAULT_PULL_ATTEMPTS,
    retry_delay: float = DEFAULT_RETRY_DELAY_SECONDS,
    timeout: int = DOCKER_PULL_TIMEOUT_SECONDS,
) -> str:
    """Pull image_ref with retries on transient errors while never hiding checksum mismatches."""
    if expected_digest is None:
        expected_digest = extract_digest(image_ref)

    last_error = ""
    for attempt in range(1, max_attempts + 1):
        print(
            f"Pulling {image_ref} (attempt {attempt}/{max_attempts})...",
            file=sys.stderr,
            flush=True,
        )
        completed = subprocess.run(
            ["docker", "pull", image_ref],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            check=False,
            timeout=timeout,
        )
        output = (completed.stdout or "").strip()
        if completed.returncode == 0:
            return verify_local_image(image_ref, expected_digest=expected_digest)

        last_error = output or f"docker pull exited {completed.returncode}"
        if is_checksum_mismatch(last_error):
            raise DockerCIError(
                f"Checksum/digest mismatch pulling {image_ref} on attempt {attempt}/{max_attempts}:\n"
                f"{last_error}",
                checksum_mismatch=True,
            )

        if attempt < max_attempts and is_transient_pull_error(last_error):
            print(
                f"Transient pull failure for {image_ref} (attempt {attempt}/{max_attempts}): "
                f"{last_error.splitlines()[-1]}",
                file=sys.stderr,
                flush=True,
            )
            time.sleep(retry_delay * attempt)
            continue

        break

    raise DockerCIError(
        f"Failed to pull {image_ref} after {attempt} attempt(s): {last_error}"
    )


def _load_cached_mock_image(cache_dir: pathlib.Path, target_tag: str) -> str | None:
    """Load target_tag from cache_dir if present, raising on checksum mismatch."""
    tar_path = cache_dir / "python-3.12-alpine.tar"
    sha_path = cache_dir / "python-3.12-alpine.tar.sha256"
    id_path = cache_dir / "python-3.12-alpine.image-id"
    if not tar_path.is_file():
        return None

    if not sha_path.is_file():
        raise DockerCIError(
            f"Cached image tarball {tar_path} is missing sidecar checksum {sha_path}",
            checksum_mismatch=True,
        )
    expected_sha = sha_path.read_text(encoding="utf-8").strip()
    actual_sha = sha256_file(tar_path)
    if actual_sha != expected_sha:
        raise DockerCIError(
            f"Cached mock Python tarball checksum mismatch for {tar_path}: "
            f"expected {expected_sha}, got {actual_sha}",
            checksum_mismatch=True,
        )

    print(f"Loading mock Python image from cache: {tar_path}", file=sys.stderr, flush=True)
    run_cmd(["docker", "load", "-i", str(tar_path)], capture=True, timeout=DOCKER_IO_TIMEOUT_SECONDS)

    inspected = inspect_local_image(target_tag)
    if inspected is None:
        raise DockerCIError(f"Loaded cache archive {tar_path} did not populate tag {target_tag}")
    image_id, _ = inspected
    if id_path.is_file():
        expected_id = id_path.read_text(encoding="utf-8").strip()
        if image_id != expected_id:
            raise DockerCIError(
                f"Cached mock Python image ID mismatch for {target_tag}: "
                f"expected {expected_id}, got {image_id}",
                checksum_mismatch=True,
            )

    verify_mock_python_runtime(target_tag)
    return target_tag


def _save_cached_mock_image(cache_dir: pathlib.Path, target_tag: str, image_id: str) -> None:
    """Atomically save target_tag and its checksum metadata into cache_dir."""
    cache_dir.mkdir(parents=True, exist_ok=True)
    tar_path = cache_dir / "python-3.12-alpine.tar"
    tmp_tar = cache_dir / "python-3.12-alpine.tar.tmp"
    sha_path = cache_dir / "python-3.12-alpine.tar.sha256"
    id_path = cache_dir / "python-3.12-alpine.image-id"

    run_cmd(
        ["docker", "save", "-o", str(tmp_tar), target_tag],
        capture=True,
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )
    digest = sha256_file(tmp_tar)
    tmp_tar.replace(tar_path)
    sha_path.write_text(digest + "\n", encoding="utf-8")
    id_path.write_text(image_id + "\n", encoding="utf-8")


def ensure_mock_python_image(
    *,
    requested_image: str = "",
    target_tag: str = LOCAL_CACHED_MOCK_PYTHON_TAG,
    cache_dir: pathlib.Path | None = None,
    candidates: tuple[str, ...] = MOCK_PYTHON_MIRRORS,
    expected_digest: str = PINNED_MOCK_PYTHON_DIGEST,
) -> str:
    """Ensure the pinned mock Python image is available locally without live Docker Hub flakes."""
    requested = requested_image.strip()
    if requested:
        req_digest = extract_digest(requested)
        inspected = inspect_local_image(requested)
        if inspected is not None:
            verify_local_image(requested, expected_digest=req_digest)
            verify_mock_python_runtime(requested)
            return requested
        if req_digest is None and not requested.startswith(LOCAL_CACHED_MOCK_PYTHON_TAG):
            raise DockerCIError(
                f"Refusing unpinned SEERR_MOCK_PYTHON_IMAGE '{requested}' that is not already local"
            )
        if req_digest is not None:
            pull_image_with_verification(requested, expected_digest=req_digest)
            verify_mock_python_runtime(requested)
            return requested

    if cache_dir is not None:
        loaded = _load_cached_mock_image(cache_dir, target_tag)
        if loaded is not None:
            return loaded

    for candidate in candidates:
        inspected = inspect_local_image(candidate)
        if inspected is not None:
            image_id = verify_local_image(candidate, expected_digest=expected_digest)
            if target_tag != candidate:
                run_cmd(["docker", "tag", candidate, target_tag], timeout=DOCKER_IO_TIMEOUT_SECONDS)
            verify_mock_python_runtime(target_tag)
            if cache_dir is not None and not (cache_dir / "python-3.12-alpine.tar").is_file():
                _save_cached_mock_image(cache_dir, target_tag, image_id)
            return target_tag

    errors: list[str] = []
    for candidate in candidates:
        try:
            image_id = pull_image_with_verification(candidate, expected_digest=expected_digest)
            if target_tag != candidate:
                run_cmd(["docker", "tag", candidate, target_tag], timeout=DOCKER_IO_TIMEOUT_SECONDS)
            verify_mock_python_runtime(target_tag)
            if cache_dir is not None:
                _save_cached_mock_image(cache_dir, target_tag, image_id)
            return target_tag
        except DockerCIError as error:
            if error.checksum_mismatch:
                raise
            errors.append(f"{candidate}: {error}")

    raise DockerCIError(
        "Unable to prepare pinned mock Python image from any mirror:\n" + "\n".join(errors)
    )


def ensure_seerr_image(image_ref: str) -> str:
    """Ensure the Seerr/Jellyseerr/Overseerr test image is pulled once with safe retries."""
    if inspect_local_image(image_ref) is not None:
        return image_ref
    pull_image_with_verification(image_ref, expected_digest=extract_digest(image_ref))
    return image_ref


def seed_config_dir(
    config_dir: pathlib.Path,
    *,
    api_key: str,
    admin_email: str,
    admin_username: str,
    admin_avatar: str,
    local_port: int | str,
) -> None:
    """Seed Seerr settings.json and SQLite admin/user_settings on the runner."""
    config_dir.mkdir(parents=True, exist_ok=True)
    db_dir = config_dir / "db"
    db_dir.mkdir(parents=True, exist_ok=True)

    settings_path = config_dir / "settings.json"
    if settings_path.is_file():
        settings = json.loads(settings_path.read_text(encoding="utf-8"))
    else:
        settings = {}

    settings.setdefault("main", {})["apiKey"] = api_key
    settings.setdefault("public", {})["initialized"] = True
    email_opts = {
        "emailFrom": "test@example.com",
        "emailUrlBase": f"http://127.0.0.1:{local_port}",
        "senderName": "Seerr Test",
        "smtpHost": "127.0.0.1",
        "smtpPort": 25,
        "secure": False,
        "ignoreTls": True,
        "requireTls": False,
        "allowSelfSigned": True,
    }
    settings.setdefault("notifications", {}).setdefault("email", {})["enabled"] = True
    settings["notifications"]["email"]["options"] = email_opts
    settings_path.write_text(json.dumps(settings, indent=1) + "\n", encoding="utf-8")

    db_path = db_dir / "db.sqlite3"
    if not db_path.is_file():
        raise DockerCIError(f"Seerr bootstrap did not create {db_path}")

    db = sqlite3.connect(db_path)
    try:
        user_result = db.execute(
            'UPDATE "user" SET email = ?, username = ?, permissions = 2, avatar = ?, userType = 1 WHERE id = 1',
            (admin_email, admin_username, admin_avatar),
        )
        if user_result.rowcount == 0:
            db.execute(
                'INSERT INTO "user" (id, email, username, permissions, avatar, userType) VALUES (1, ?, ?, 2, ?, 1)',
                (admin_email, admin_username, admin_avatar),
            )

        settings_result = db.execute(
            "UPDATE user_settings SET locale = 'en', watchlistSyncMovies = 0, watchlistSyncTv = 0 WHERE userId = 1"
        )
        if settings_result.rowcount == 0:
            db.execute(
                "INSERT INTO user_settings (userId, locale, watchlistSyncMovies, watchlistSyncTv) VALUES (1, 'en', 0, 0)"
            )
        db.commit()
        db.execute("PRAGMA wal_checkpoint(TRUNCATE)")
    finally:
        db.close()


def build_config_tar(config_dir: pathlib.Path) -> bytes:
    """Package seeded settings.json and db/db.sqlite3 with UID/GID 1000 ownership."""
    buffer = io.BytesIO()
    with tarfile.open(fileobj=buffer, mode="w") as archive:
        for rel_path in ("settings.json", "db/db.sqlite3"):
            source = config_dir / rel_path
            data = source.read_bytes()
            info = tarfile.TarInfo(rel_path)
            info.size = len(data)
            info.mode = 0o644
            info.uid = SEERR_UID
            info.gid = SEERR_GID
            info.uname = "node"
            info.gname = "node"
            info.mtime = int(source.stat().st_mtime)
            archive.addfile(info, io.BytesIO(data))
    return buffer.getvalue()


def seed_seerr_container(
    compose_cmd: list[str],
    container_name: str,
    work_dir: pathlib.Path,
    *,
    compose_cwd: pathlib.Path | None = None,
    api_key: str,
    admin_email: str,
    admin_username: str,
    admin_avatar: str,
    local_port: int | str,
) -> None:
    """Stop Seerr, seed /app/config on the runner without container package installs, and start Seerr."""
    print("Seeding Seerr database and settings from the runner...", file=sys.stderr, flush=True)
    run_cmd(compose_cmd + ["stop", "seerr"], cwd=compose_cwd, timeout=DOCKER_IO_TIMEOUT_SECONDS)

    seed_dir = work_dir / "seed-config"
    shutil.rmtree(seed_dir, ignore_errors=True)
    run_cmd(
        ["docker", "cp", f"{container_name}:/app/config", str(seed_dir)],
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )

    seed_config_dir(
        seed_dir,
        api_key=api_key,
        admin_email=admin_email,
        admin_username=admin_username,
        admin_avatar=admin_avatar,
        local_port=local_port,
    )

    # Clear pre-stop WAL/SHM files using the already-pulled Seerr image so the
    # checkpointed DB is authoritative when Seerr starts back up.
    run_cmd(
        compose_cmd
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
        cwd=compose_cwd,
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )
    run_cmd(
        ["docker", "cp", "-", f"{container_name}:/app/config"],
        input_bytes=build_config_tar(seed_dir),
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )
    run_cmd(compose_cmd + ["start", "seerr"], cwd=compose_cwd, timeout=DOCKER_IO_TIMEOUT_SECONDS)


def _parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    mock_parser = subparsers.add_parser("prepare-mock-image")
    mock_parser.add_argument("--cache-dir", type=pathlib.Path, default=None)
    mock_parser.add_argument("--github-env", type=pathlib.Path, default=None)
    mock_parser.add_argument("--requested-image", default=os.environ.get("SEERR_MOCK_PYTHON_IMAGE", ""))

    seerr_parser = subparsers.add_parser("ensure-seerr-image")
    seerr_parser.add_argument("image")

    seed_parser = subparsers.add_parser("seed-seerr")
    seed_parser.add_argument("--compose-cmd", default="docker compose")
    seed_parser.add_argument("--tests-dir", type=pathlib.Path, required=True)
    seed_parser.add_argument("--container-name", default="seerr-test")
    seed_parser.add_argument("--api-key", required=True)
    seed_parser.add_argument("--admin-email", required=True)
    seed_parser.add_argument("--admin-username", required=True)
    seed_parser.add_argument("--admin-avatar", required=True)
    seed_parser.add_argument("--local-port", required=True)

    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(sys.argv[1:] if argv is None else argv)
    try:
        if args.command == "prepare-mock-image":
            resolved = ensure_mock_python_image(
                requested_image=args.requested_image,
                cache_dir=args.cache_dir,
            )
            if args.github_env is not None:
                with args.github_env.open("a", encoding="utf-8") as handle:
                    handle.write(f"SEERR_MOCK_PYTHON_IMAGE={resolved}\n")
            print(resolved)
            return 0

        if args.command == "ensure-seerr-image":
            resolved = ensure_seerr_image(args.image)
            print(resolved)
            return 0

        if args.command == "seed-seerr":
            compose_cmd = shlex.split(args.compose_cmd)
            with tempfile.TemporaryDirectory(prefix="seerr-seed-") as tmp_dir:
                seed_seerr_container(
                    compose_cmd,
                    args.container_name,
                    pathlib.Path(tmp_dir),
                    compose_cwd=args.tests_dir,
                    api_key=args.api_key,
                    admin_email=args.admin_email,
                    admin_username=args.admin_username,
                    admin_avatar=args.admin_avatar,
                    local_port=args.local_port,
                )
            return 0
    except DockerCIError as error:
        print(f"CI DOCKER ERROR: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
