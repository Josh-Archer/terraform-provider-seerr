#!/usr/bin/env python3
"""Exercise watchlist pagination and drift against an isolated real Seerr 3.5.0."""

from __future__ import annotations

import io
import json
import os
import pathlib
import shutil
import sqlite3
import subprocess
import sys
import tarfile
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid


ROOT = pathlib.Path(__file__).resolve().parent.parent
COMPOSE_FILE = ROOT / "tests" / "docker-compose.watchlist-live.yml"
IMAGE = "ghcr.io/seerr-team/seerr:v3.5.0"
API_KEY = "seerr-watchlist-live-ci-key"
FIXTURE_COUNT = 42
FIRST_TMDB_ID = 9_220_000
COMMAND_TIMEOUT_SECONDS = 300
DOCKER_IO_TIMEOUT_SECONDS = 60
LOG_CAPTURE_TIMEOUT_SECONDS = 30
SERVER_READY_TIMEOUT_SECONDS = 180
AUTH_READY_TIMEOUT_SECONDS = 60
FAILURE_LOG_TAIL_LINES = 50
SEERR_UID = 1000
SEERR_GID = 1000


class IntegrationFailure(RuntimeError):
    pass


def require(condition: bool, message: str) -> None:
    if not condition:
        raise IntegrationFailure(message)


def run(command: list[str], *, cwd: pathlib.Path | None = None, env=None,
        input_text: str | None = None, input_bytes: bytes | None = None,
        capture: bool = False, allowed_codes: tuple[int, ...] = (0,),
        timeout: int = COMMAND_TIMEOUT_SECONDS) -> subprocess.CompletedProcess:
    if input_text is not None and input_bytes is not None:
        raise ValueError("input_text and input_bytes are mutually exclusive")
    print("+", " ".join(command), flush=True)
    use_text = input_bytes is None
    result = subprocess.run(
        command,
        cwd=cwd,
        env=env,
        input=input_bytes if input_bytes is not None else input_text,
        text=use_text,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.STDOUT if capture else None,
        check=False,
        timeout=timeout,
    )
    if result.returncode not in allowed_codes:
        output = ""
        if capture and result.stdout is not None:
            output = result.stdout if isinstance(result.stdout, str) else result.stdout.decode("utf-8", errors="replace")
        raise IntegrationFailure(
            f"command exited {result.returncode}: {' '.join(command)}\n{output}"
        )
    return result


def write(path: pathlib.Path, content: str) -> None:
    path.write_text(content, encoding="utf-8")


def api_request(base_url: str, path: str, *, method: str = "GET") -> tuple[int, object]:
    request = urllib.request.Request(
        base_url + path,
        headers={"X-Api-Key": API_KEY, "Accept": "application/json"},
        method=method,
    )
    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            body = response.read()
            return response.status, json.loads(body) if body else None
    except urllib.error.HTTPError as error:
        body = error.read().decode("utf-8", errors="replace")
        raise IntegrationFailure(
            f"Seerr {method} {path} returned HTTP {error.code}: {body}"
        ) from error


def seed_config_dir(config_dir: pathlib.Path) -> None:
    """Seed settings.json and the SQLite watchlist fixtures in a host copy of /app/config."""
    settings_path = config_dir / "settings.json"
    if not settings_path.is_file():
        raise IntegrationFailure(f"Seerr bootstrap did not create {settings_path}")
    settings = json.loads(settings_path.read_text(encoding="utf-8"))
    settings.setdefault("main", {})["apiKey"] = API_KEY
    settings.setdefault("public", {})["initialized"] = True
    settings_path.write_text(json.dumps(settings, indent=1) + "\n", encoding="utf-8")

    db_path = config_dir / "db" / "db.sqlite3"
    if not db_path.is_file():
        raise IntegrationFailure(f"Seerr bootstrap did not create {db_path}")

    db = sqlite3.connect(db_path)
    try:
        db.row_factory = sqlite3.Row
        user_result = db.execute(
            'UPDATE "user" SET email=?, username=?, permissions=2, avatar=?, userType=1 WHERE id=1',
            ("watchlist-ci@example.invalid", "watchlist-ci-admin", "/logo_full.svg"),
        )
        if user_result.rowcount == 0:
            db.execute(
                'INSERT INTO "user" (id, email, username, permissions, avatar, userType) VALUES (1, ?, ?, 2, ?, 1)',
                ("watchlist-ci@example.invalid", "watchlist-ci-admin", "/logo_full.svg"),
            )
        settings_result = db.execute(
            "UPDATE user_settings SET locale='en', watchlistSyncMovies=0, watchlistSyncTv=0 WHERE userId=1"
        )
        if settings_result.rowcount == 0:
            db.execute(
                "INSERT INTO user_settings (userId, locale, watchlistSyncMovies, watchlistSyncTv) VALUES (1, 'en', 0, 0)"
            )
        tables = {row[0].lower(): row[0] for row in db.execute("SELECT name FROM sqlite_master WHERE type='table'")}
        table = tables.get("watchlist")
        if table is None:
            raise IntegrationFailure("Seerr 3.5.0 database has no watchlist table")
        columns = {row[1] for row in db.execute(f'PRAGMA table_info("{table}")')}
        required = {"tmdbId", "mediaType", "title", "ratingKey", "requestedById"}
        missing = required - columns
        if missing:
            raise IntegrationFailure(f"watchlist schema is missing columns: {sorted(missing)}")
        user_row = db.execute('SELECT id FROM "user" WHERE id=1').fetchone()
        if user_row is None:
            raise IntegrationFailure("expected the existing Seerr bootstrap to create admin user 1")
        start_id = FIRST_TMDB_ID
        end_id = start_id + FIXTURE_COUNT
        collision = db.execute(
            'SELECT tmdbId, mediaType FROM "' + table + '" WHERE tmdbId >= ? AND tmdbId < ?',
            (start_id, end_id),
        ).fetchall()
        if collision:
            raise IntegrationFailure(f"fixture TMDB IDs already exist: {[tuple(row) for row in collision]}")

        for offset in range(FIXTURE_COUNT):
            tmdb_id = start_id + offset
            row = {
                "tmdbId": tmdb_id,
                "mediaType": "movie",
                "title": f"Watchlist CI Fixture {offset + 1:02d}",
                "ratingKey": f"watchlist-ci-{offset + 1:02d}",
                "requestedById": 1,
            }
            if "createdAt" in columns:
                row["createdAt"] = f"2024-01-{(offset // 28) + 1:02d}T00:{offset % 60:02d}:00.000Z"
            if "updatedAt" in columns:
                row["updatedAt"] = f"2024-01-{(offset // 28) + 1:02d}T00:{offset % 60:02d}:00.000Z"
            names = list(row)
            quoted = ", ".join('"' + name + '"' for name in names)
            placeholders = ", ".join("?" for _ in names)
            db.execute(f'INSERT INTO "{table}" ({quoted}) VALUES ({placeholders})', [row[name] for name in names])

        db.commit()
        count = db.execute(
            'SELECT COUNT(*) FROM "' + table + '" WHERE requestedById=1 AND tmdbId >= ? AND tmdbId < ?',
            (start_id, end_id),
        ).fetchone()[0]
        if count != FIXTURE_COUNT:
            raise IntegrationFailure(f"expected {FIXTURE_COUNT} seeded watchlist rows, found {count}")
        db.execute("PRAGMA wal_checkpoint(TRUNCATE)")
    finally:
        db.close()
    print(f"Seeded {FIXTURE_COUNT} native watchlist rows for admin user 1")


def build_config_tar(config_dir: pathlib.Path) -> bytes:
    """Package seeded config files with Seerr container UID/GID ownership."""
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


def seed_container_database(compose: list[str], container_id: str, work_dir: pathlib.Path) -> None:
    """Stop Seerr, seed its config/SQLite DB on the runner, and copy it back."""
    print("Seeding Seerr configuration and SQLite watchlist fixtures from the runner...")
    run(compose + ["stop", "seerr"], timeout=DOCKER_IO_TIMEOUT_SECONDS)

    seed_dir = work_dir / "seed-config"
    shutil.rmtree(seed_dir, ignore_errors=True)
    run(["docker", "cp", f"{container_id}:/app/config", str(seed_dir)], timeout=DOCKER_IO_TIMEOUT_SECONDS)

    seed_config_dir(seed_dir)

    # Clear any pre-stop SQLite WAL/SHM files so the checkpointed DB is authoritative on restart.
    run(
        compose + [
            "run", "--rm", "--no-deps", "-T", "--entrypoint", "sh", "seerr",
            "-c", "rm -f /app/config/db/db.sqlite3-wal /app/config/db/db.sqlite3-shm",
        ],
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )
    run(
        ["docker", "cp", "-", f"{container_id}:/app/config"],
        input_bytes=build_config_tar(seed_dir),
        timeout=DOCKER_IO_TIMEOUT_SECONDS,
    )


def wait_for_server(base_url: str, container_id: str, *,
                    timeout_seconds: float = SERVER_READY_TIMEOUT_SECONDS) -> None:
    deadline = time.monotonic() + timeout_seconds
    attempt = 0
    while True:
        try:
            with urllib.request.urlopen(base_url + "/api/v1/status", timeout=3) as response:
                if response.status == 200:
                    return
        except (OSError, urllib.error.URLError):
            pass
        attempt += 1
        if attempt in (25, 55):
            print("Waiting for Seerr 3.5.0 startup migrations...")
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            break
        time.sleep(min(2.0, remaining))
    raise IntegrationFailure(f"Seerr did not become ready within {timeout_seconds:.0f}s (container {container_id})")


def wait_for_auth(base_url: str, *, timeout_seconds: float = AUTH_READY_TIMEOUT_SECONDS) -> dict:
    deadline = time.monotonic() + timeout_seconds
    while True:
        try:
            status, user = api_request(base_url, "/api/v1/auth/me")
            if status == 200 and isinstance(user, dict) and str(user.get("id")) == "1":
                return user
        except (OSError, urllib.error.URLError, IntegrationFailure):
            pass
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            break
        time.sleep(min(2.0, remaining))
    raise IntegrationFailure("bootstrapped API key did not authenticate")


def create_provider(work_dir: pathlib.Path, base_url: str, target_id: int) -> None:
    tofu_dir = work_dir / "tofu"
    tofu_dir.mkdir()
    escaped_url = json.dumps(base_url)
    escaped_key = json.dumps(API_KEY)
    write(tofu_dir / "main.tf", f'''terraform {{
  required_providers {{
    seerr = {{
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }}
  }}
}}

provider "seerr" {{
  url     = {escaped_url}
  api_key = {escaped_key}
}}

data "seerr_watchlist" "authenticated" {{}}

data "seerr_watchlist" "explicit" {{
  user_id = 1
}}

resource "seerr_watchlist" "target" {{
  tmdb_id    = {target_id}
  media_type = "movie"
}}

resource "seerr_watchlist" "anchor" {{
  tmdb_id    = {FIRST_TMDB_ID}
  media_type = "movie"
}}

output "authenticated_total" {{ value = data.seerr_watchlist.authenticated.total }}
output "authenticated_ids" {{ value = [for item in data.seerr_watchlist.authenticated.watchlist : item.tmdb_id] }}
output "explicit_total" {{ value = data.seerr_watchlist.explicit.total }}
output "explicit_ids" {{ value = [for item in data.seerr_watchlist.explicit.watchlist : item.tmdb_id] }}
''')
    write(work_dir / "tofurc", "provider_installation { direct {} }\n")
    goos = run(["go", "env", "GOOS"], capture=True).stdout.strip()
    goarch = run(["go", "env", "GOARCH"], capture=True).stdout.strip()
    goexe = run(["go", "env", "GOEXE"], capture=True).stdout.strip()
    platform_dir = work_dir / "provider-mirror" / "registry.opentofu.org" / "josh-archer" / "seerr" / "99.99.99" / f"{goos}_{goarch}"
    platform_dir.mkdir(parents=True)
    run(["go", "build", "-o", str(platform_dir / f"terraform-provider-seerr{goexe}"), "."], cwd=ROOT)
    env = os.environ.copy()
    env["TF_CLI_CONFIG_FILE"] = str(work_dir / "tofurc")
    run(["tofu", "init", "-input=false", f"-plugin-dir={work_dir / 'provider-mirror'}"], cwd=tofu_dir, env=env)
    (work_dir / "tf_env.json").write_text(json.dumps({"TF_CLI_CONFIG_FILE": env["TF_CLI_CONFIG_FILE"]}))


def tofu_env(work_dir: pathlib.Path) -> dict[str, str]:
    env = os.environ.copy()
    env["TF_CLI_CONFIG_FILE"] = json.loads((work_dir / "tf_env.json").read_text())["TF_CLI_CONFIG_FILE"]
    return env


def output_json(work_dir: pathlib.Path, name: str) -> object:
    result = run(["tofu", "output", "-json", name], cwd=work_dir / "tofu",
                 env=tofu_env(work_dir), capture=True)
    return json.loads(result.stdout)


def assert_watchlist_outputs(work_dir: pathlib.Path, expected_ids: set[int]) -> None:
    for flavor in ("authenticated", "explicit"):
        total = output_json(work_dir, flavor + "_total")
        ids = output_json(work_dir, flavor + "_ids")
        require(total == len(expected_ids), f"{flavor} datasource total={total}, expected {len(expected_ids)}")
        require(len(ids) == len(expected_ids), f"{flavor} datasource list has {len(ids)} rows, expected {len(expected_ids)}")
        require(set(ids) == expected_ids, f"{flavor} datasource returned IDs different from the seeded rows")


def run_terraform_flow(work_dir: pathlib.Path, base_url: str, expected_ids: set[int],
                       last_page_id: int) -> None:
    tofu_dir = work_dir / "tofu"
    env = tofu_env(work_dir)

    print("Checking authenticated and explicit-user reads through OpenTofu...")
    run(["tofu", "apply", "-refresh-only", "-auto-approve", "-input=false"],
        cwd=tofu_dir, env=env)
    assert_watchlist_outputs(work_dir, expected_ids)

    print(f"Importing final-page item {last_page_id}...")
    run(["tofu", "import", "-input=false", "seerr_watchlist.target", f"{last_page_id}:movie"],
        cwd=tofu_dir, env=env)
    print(f"Importing unrelated first-page anchor {FIRST_TMDB_ID}...")
    run(["tofu", "import", "-input=false", "seerr_watchlist.anchor", f"{FIRST_TMDB_ID}:movie"],
        cwd=tofu_dir, env=env)
    plan = run(["tofu", "plan", "-detailed-exitcode", "-input=false"],
               cwd=tofu_dir, env=env, capture=True, allowed_codes=(0, 2))
    require(plan.returncode == 0, f"expected no-change plan after importing final-page item, got exit {plan.returncode}\n{plan.stdout}")
    print("Final-page import refresh produced a no-change plan.")

    status, page = api_request(base_url, "/api/v1/user/1/watchlist?page=3")
    require(status == 200 and isinstance(page, dict), "could not read final watchlist page before deletion")
    results = page.get("results")
    require(isinstance(results, list) and any(row.get("tmdbId") == last_page_id for row in results),
            "import target was not actually present on the final upstream page")

    query = urllib.parse.urlencode({"mediaType": "movie"})
    status, _ = api_request(base_url, f"/api/v1/watchlist/{last_page_id}?{query}", method="DELETE")
    require(status == 204, f"upstream delete returned {status}, expected 204")

    print("Checking refresh-only drift after external deletion...")
    plan_path = work_dir / "watchlist-refresh.tfplan"
    plan = run(["tofu", "plan", "-refresh-only", "-input=false", "-out", str(plan_path), "-detailed-exitcode"],
               cwd=tofu_dir, env=env, capture=True, allowed_codes=(0, 2))
    require(plan.returncode == 2, f"expected refresh-only plan to detect deleted resource, got exit {plan.returncode}\n{plan.stdout}")
    show = run(["tofu", "show", "-json", str(plan_path)], cwd=tofu_dir, env=env, capture=True)
    plan_json = json.loads(show.stdout)
    drift = [change for change in plan_json.get("resource_drift", []) if change.get("mode") == "managed"]
    require(len(drift) == 1, f"expected drift for only the imported watchlist resource, found {len(drift)} managed drift entries")
    change = drift[0]
    require(change.get("address") == "seerr_watchlist.target" and change.get("change", {}).get("actions") == ["delete"],
            f"refresh-only plan did not report only the missing target: {change}")
    normal_changes = [change for change in plan_json.get("resource_changes", []) if change.get("mode") == "managed"]
    require(all(change.get("change", {}).get("actions") == ["no-op"] for change in normal_changes),
            f"refresh-only plan proposed managed actions in addition to drift: {normal_changes}")

    run(["tofu", "apply", "-auto-approve", "-input=false", str(plan_path)],
        cwd=tofu_dir, env=env)
    state = run(["tofu", "state", "list"], cwd=tofu_dir, env=env, capture=True)
    require("seerr_watchlist.target" not in state.stdout, "deleted target remains in refreshed OpenTofu state")
    state_json = json.loads(run(["tofu", "show", "-json"], cwd=tofu_dir, env=env, capture=True).stdout)
    managed_state = [resource for resource in state_json.get("values", {}).get("root_module", {}).get("resources", [])
                     if resource.get("mode") == "managed"]
    addresses = {resource.get("address") for resource in managed_state}
    require(addresses == {"seerr_watchlist.anchor"},
            f"refresh-only should retain only the unrelated anchor resource, found {sorted(addresses)}")
    assert_watchlist_outputs(work_dir, expected_ids - {last_page_id})

    _, current = api_request(base_url, "/api/v1/user/1/watchlist?page=1")
    require(current.get("totalResults") == 41, "upstream did not retain the other 41 fixture rows")
    print("PASS: both read modes returned 42 items; final-page import stayed unchanged; refresh removed only the deleted item and retained 41 others.")


def tail_lines(text: str, limit: int = FAILURE_LOG_TAIL_LINES) -> str:
    if not text or limit <= 0:
        return ""
    return "\n".join(text.splitlines()[-limit:])


def _capture_command_output(command: list[str], *, timeout: int = LOG_CAPTURE_TIMEOUT_SECONDS) -> str:
    try:
        completed = subprocess.run(
            command,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            check=False,
            timeout=timeout,
        )
        return completed.stdout or ""
    except (subprocess.SubprocessError, OSError) as capture_error:
        return f"<failed to capture output from {' '.join(command)}: {capture_error}>\n"


def capture_failure_diagnostics(compose: list[str], container_id: str,
                                work_dir: pathlib.Path, *, stream=None) -> None:
    if stream is None:
        stream = sys.stderr
    container_logs = ""
    if container_id:
        container_logs = _capture_command_output(["docker", "logs", container_id])
        write(work_dir / "seerr-container.log", container_logs)
    compose_logs = _capture_command_output(compose + ["logs", "--no-color"])
    write(work_dir / "compose.log", compose_logs)

    tail = tail_lines(container_logs or compose_logs, FAILURE_LOG_TAIL_LINES)
    if tail:
        print(f"--- Last {FAILURE_LOG_TAIL_LINES} lines of Seerr container logs ---", file=stream)
        print(tail, file=stream)
        print("--- End of Seerr container logs ---", file=stream)
    print(f"Diagnostics saved under {work_dir}", file=stream)


def service_url(compose: list[str]) -> str:
    result = run(compose + ["port", "seerr", "5055"], capture=True, timeout=DOCKER_IO_TIMEOUT_SECONDS)
    host = result.stdout.strip().splitlines()[-1]
    host_port = int(host.rsplit(":", 1)[-1])
    require(0 < host_port < 65536, "Docker did not publish a valid Seerr port")
    base_url = f"http://127.0.0.1:{host_port}"
    print(f"Disposable Seerr endpoint: {base_url}", flush=True)
    return base_url


def main() -> int:
    if "SEERR_URL" in os.environ or "SEERR_API_KEY" in os.environ:
        print("Refusing to run with external Seerr target variables; this integration owns its disposable Docker stack.", file=sys.stderr)
        return 2
    for tool in ("docker", "go", "tofu"):
        if shutil.which(tool) is None:
            print(f"Required command not found: {tool}", file=sys.stderr)
            return 2

    artifacts = pathlib.Path(os.environ.get("SEERR_WATCHLIST_ARTIFACT_DIR", ROOT / "test-artifacts" / "watchlist-live"))
    artifacts.mkdir(parents=True, exist_ok=True)
    work_dir = artifacts / f"run-{uuid.uuid4().hex[:12]}"
    work_dir.mkdir()
    project = f"watchlist-live-{uuid.uuid4().hex[:10]}"
    compose = ["docker", "compose", "-p", project, "-f", str(COMPOSE_FILE)]
    container_id = ""
    base_url = ""
    try:
        print(f"Starting disposable {IMAGE} stack (project {project})...")
        run(compose + ["up", "-d"])
        result = run(compose + ["ps", "-q", "seerr"], capture=True, timeout=DOCKER_IO_TIMEOUT_SECONDS)
        container_id = result.stdout.strip()
        require(container_id != "", "Compose did not return the Seerr container ID")
        base_url = service_url(compose)
        wait_for_server(base_url, container_id)

        seed_container_database(compose, container_id, work_dir)
        run(compose + ["start", "seerr"], timeout=DOCKER_IO_TIMEOUT_SECONDS)
        # Docker can allocate a new random host port when the container starts again.
        base_url = service_url(compose)
        wait_for_server(base_url, container_id)
        wait_for_auth(base_url)

        expected_ids = set(range(FIRST_TMDB_ID, FIRST_TMDB_ID + FIXTURE_COUNT))
        status, first_page = api_request(base_url, "/api/v1/user/1/watchlist?page=1")
        require(status == 200 and isinstance(first_page, dict), "user watchlist endpoint did not return an object")
        require(first_page.get("totalResults") == FIXTURE_COUNT and first_page.get("totalPages") == 3,
                f"upstream fixture did not produce 42 results over 3 pages: {first_page}")
        last_page_number = first_page["totalPages"]
        status, final_page = api_request(base_url, f"/api/v1/user/1/watchlist?page={last_page_number}")
        require(status == 200 and isinstance(final_page, dict), "could not fetch the final upstream page")
        final_results = final_page.get("results")
        require(isinstance(final_results, list) and len(final_results) == 2,
                f"expected two items on the final page, got {final_results}")
        last_page_id = int(final_results[-1]["tmdbId"])
        require(last_page_id in expected_ids, f"final page item {last_page_id} is not a seeded fixture")

        create_provider(work_dir, base_url, last_page_id)
        run_terraform_flow(work_dir, base_url, expected_ids, last_page_id)
        return 0
    except Exception as error:
        print(f"WATCHLIST LIVE INTEGRATION FAILED: {error}", file=sys.stderr)
        capture_failure_diagnostics(compose, container_id, work_dir, stream=sys.stderr)
        return 1
    finally:
        primary_error_active = sys.exc_info()[0] is not None
        try:
            down = subprocess.run(compose + ["down", "-v", "--remove-orphans"], text=True,
                                  stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=False,
                                  timeout=COMMAND_TIMEOUT_SECONDS)
            write(work_dir / "cleanup.log", down.stdout)
            if down.returncode != 0:
                cleanup_message = f"Could not fully remove disposable Compose project {project}: {down.stdout}"
                write(work_dir / "cleanup-error.log", cleanup_message + "\n")
                print(cleanup_message, file=sys.stderr)
                if not primary_error_active:
                    raise IntegrationFailure("failed to clean up the isolated watchlist Docker stack")
        except Exception as cleanup_error:
            if not primary_error_active:
                raise IntegrationFailure(f"failed to clean up isolated Compose project {project}: {cleanup_error}") from cleanup_error
            print(f"Cleanup also failed for disposable Compose project {project}: {cleanup_error}", file=sys.stderr)


if __name__ == "__main__":
    sys.exit(main())
