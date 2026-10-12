#!/usr/bin/env python3
"""Exercise watchlist pagination and drift against an isolated real Seerr 3.5.0."""

from __future__ import annotations

import json
import os
import pathlib
import shutil
import subprocess
import sys
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


class IntegrationFailure(RuntimeError):
    pass


def require(condition: bool, message: str) -> None:
    if not condition:
        raise IntegrationFailure(message)


def run(command: list[str], *, cwd: pathlib.Path | None = None, env=None,
        input_text: str | None = None, capture: bool = False,
        allowed_codes: tuple[int, ...] = (0,)) -> subprocess.CompletedProcess[str]:
    print("+", " ".join(command), flush=True)
    result = subprocess.run(
        command,
        cwd=cwd,
        env=env,
        input=input_text,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.STDOUT if capture else None,
        check=False,
        timeout=COMMAND_TIMEOUT_SECONDS,
    )
    if result.returncode not in allowed_codes:
        output = result.stdout or "" if capture else ""
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


def seed_database(container_id: str) -> None:
    bootstrap = r'''import json
import sqlite3
from pathlib import Path

config_dir = Path("/app/config")
settings_path = config_dir / "settings.json"
settings = json.loads(settings_path.read_text())
settings.setdefault("main", {})["apiKey"] = "seerr-watchlist-live-ci-key"
settings.setdefault("public", {})["initialized"] = True
settings_path.write_text(json.dumps(settings, indent=1) + "\n")

db = sqlite3.connect(config_dir / "db" / "db.sqlite3")
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
    raise RuntimeError("Seerr 3.5.0 database has no watchlist table")
columns = {row[1] for row in db.execute(f'PRAGMA table_info("{table}")')}
required = {"tmdbId", "mediaType", "title", "ratingKey", "requestedById"}
missing = required - columns
if missing:
    raise RuntimeError(f"watchlist schema is missing columns: {sorted(missing)}")
user_row = db.execute('SELECT id FROM "user" WHERE id=1').fetchone()
if user_row is None:
    raise RuntimeError("expected the existing Seerr bootstrap to create admin user 1")
start_id = 9220000
end_id = start_id + 42
collision = db.execute(
    'SELECT tmdbId, mediaType FROM "' + table + '" WHERE tmdbId >= ? AND tmdbId < ?',
    (start_id, end_id),
).fetchall()
if collision:
    raise RuntimeError(f"fixture TMDB IDs already exist: {[tuple(row) for row in collision]}")

for offset in range(42):
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
count = db.execute('SELECT COUNT(*) FROM "' + table + '" WHERE requestedById=1 AND tmdbId >= ? AND tmdbId < ?', (start_id, end_id)).fetchone()[0]
if count != 42:
    raise RuntimeError(f"expected 42 seeded watchlist rows, found {count}")
db.close()
print("Seeded 42 native watchlist rows for admin user 1")
'''
    run(["docker", "exec", "-i", "-u", "root", container_id, "python3", "-"],
        input_text=bootstrap)
    run(["docker", "exec", "-u", "root", container_id, "sh", "-c",
         "chown -R 1000:1000 /app/config"])


def wait_for_server(base_url: str, container_id: str) -> None:
    for attempt in range(90):
        try:
            with urllib.request.urlopen(base_url + "/api/v1/status", timeout=3) as response:
                if response.status == 200:
                    return
        except (OSError, urllib.error.URLError):
            pass
        if attempt in (25, 55):
            print("Waiting for Seerr 3.5.0 startup migrations...")
        time.sleep(2)
    logs = subprocess.run(["docker", "logs", container_id], text=True,
                          stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=False)
    raise IntegrationFailure(f"Seerr did not become ready. Container log:\n{logs.stdout}")


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


def service_url(compose: list[str]) -> str:
    result = run(compose + ["port", "seerr", "5055"], capture=True)
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
        result = run(compose + ["ps", "-q", "seerr"], capture=True)
        container_id = result.stdout.strip()
        require(container_id != "", "Compose did not return the Seerr container ID")
        base_url = service_url(compose)
        wait_for_server(base_url, container_id)

        print("Installing the SQLite bootstrap tools in the disposable container...")
        run(["docker", "exec", "-u", "root", container_id, "sh", "-c",
             "attempt=1; until apk add --no-cache python3 sqlite; do "
             "if [ \"$attempt\" -ge 10 ]; then exit 1; fi; "
             "attempt=$((attempt + 1)); sleep 2; done"])
        seed_database(container_id)
        run(compose + ["restart", "seerr"])
        # Docker can allocate a new random host port when the container restarts.
        base_url = service_url(compose)
        wait_for_server(base_url, container_id)
        authenticated = False
        for _ in range(30):
            try:
                api_request(base_url, "/api/v1/auth/me")
                authenticated = True
                break
            except (OSError, urllib.error.URLError, IntegrationFailure):
                time.sleep(2)
        require(authenticated, "bootstrapped API key did not authenticate")
        status, user = api_request(base_url, "/api/v1/auth/me")
        require(status == 200 and isinstance(user, dict) and str(user.get("id")) == "1",
                f"expected authenticated admin user 1, got {user}")

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
        if container_id:
            logs = subprocess.run(["docker", "logs", container_id], text=True,
                                  stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=False)
            write(work_dir / "seerr-container.log", logs.stdout)
        diagnostics = subprocess.run(compose + ["logs", "--no-color"], text=True,
                                     stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=False)
        write(work_dir / "compose.log", diagnostics.stdout)
        print(f"Diagnostics saved under {work_dir}", file=sys.stderr)
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
