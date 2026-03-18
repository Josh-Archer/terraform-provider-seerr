#!/usr/bin/env bash
set -euo pipefail

exec 3>&1
exec 1>&2

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tests_dir="${script_dir}/../tests"
port="${SEERR_TEST_PORT:-5055}"
local_port="${SEERR_TEST_LOCAL_PORT:-15055}"
api_key="${SEERR_BOOTSTRAP_API_KEY:-seerr-ci-api-key}"
admin_email="${SEERR_TEST_ADMIN_EMAIL:-ci-admin@example.invalid}"
admin_username="${SEERR_TEST_ADMIN_USERNAME:-ci-admin}"
admin_avatar="${SEERR_TEST_ADMIN_AVATAR:-/logo_full.svg}"

if ! command -v docker-compose >/dev/null 2>&1 && ! command -v docker >/dev/null 2>&1; then
  echo "docker or docker-compose is required" >&2
  exit 1
fi

compose_cmd="docker compose"
if command -v docker-compose >/dev/null 2>&1; then
  compose_cmd="docker-compose"
fi

cd "${tests_dir}"

echo "Starting Seerr via Docker Compose..."
${compose_cmd} down -v
${compose_cmd} up -d

echo "Waiting for Seerr to become ready..."
for _ in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${local_port}/api/v1/status" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! curl -fsS "http://127.0.0.1:${local_port}/api/v1/status" >/dev/null 2>&1; then
  echo "Seerr did not become reachable in Docker at http://127.0.0.1:${local_port}" >&2
  ${compose_cmd} logs
  exit 1
fi

echo "Bootstrapping Seerr database and settings..."
cat <<EOF | docker exec -i -u root seerr-test sh
apk add --no-cache python3 sqlite
EOF

cat <<EOF | docker exec -i -u root seerr-test python3
import json
import sqlite3
from pathlib import Path
import os

config_dir = Path("/app/config")
config_dir.mkdir(parents=True, exist_ok=True)
db_dir = config_dir / "db"
db_dir.mkdir(parents=True, exist_ok=True)

settings_path = config_dir / "settings.json"
try:
    settings = json.loads(settings_path.read_text())
except FileNotFoundError:
    settings = {}

settings.setdefault("main", {})["apiKey"] = "${api_key}"
settings.setdefault("public", {})["initialized"] = True

email_opts = {
    "emailFrom": "test@example.com",
    "emailUrlBase": "http://127.0.0.1:15055",
    "senderName": "Seerr Test",
    "smtpHost": "127.0.0.1",
    "smtpPort": 25,
    "secure": False,
    "ignoreTls": True,
    "requireTls": False,
    "allowSelfSigned": True
}
settings.setdefault("notifications", {}).setdefault("email", {})["enabled"] = True
settings["notifications"]["email"]["options"] = email_opts

settings_path.write_text(json.dumps(settings, indent=1) + "\n")
db_path = db_dir / "db.sqlite3"
db = sqlite3.connect(str(db_path))
user_result = db.execute(
    "UPDATE user SET email = ?, username = ?, permissions = 2, avatar = ?, userType = 1 WHERE id = 1",
    ("${admin_email}", "${admin_username}", "${admin_avatar}"),
)
if user_result.rowcount == 0:
    db.execute(
        "INSERT INTO user (id, email, username, permissions, avatar, userType) VALUES (1, ?, ?, 2, ?, 1)",
        ("${admin_email}", "${admin_username}", "${admin_avatar}"),
    )

settings_result = db.execute(
    "UPDATE user_settings SET locale = 'en', watchlistSyncMovies = 0, watchlistSyncTv = 0 WHERE userId = 1"
)
if settings_result.rowcount == 0:
    db.execute(
        "INSERT INTO user_settings (userId, locale, watchlistSyncMovies, watchlistSyncTv) VALUES (1, 'en', 0, 0)"
    )
db.commit()
db.close()

# Ensure the Seerr user can own the files
os.system("chown -R 1000:1000 /app/config")
EOF

echo "Restarting Seerr to apply bootstrapped settings..."
${compose_cmd} restart seerr

echo "Waiting for Seerr to become ready after restart..."
for _ in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${local_port}/api/v1/status" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! curl -fsS -H "X-Api-Key: ${api_key}" "http://127.0.0.1:${local_port}/api/v1/settings/main" >/dev/null 2>&1; then
  echo "Bootstrapped API key did not authenticate against http://127.0.0.1:${local_port}" >&2
  exit 1
fi

echo "Enabling email notifications via API..."
curl -fsS -H "X-Api-Key: ${api_key}" -H "Content-Type: application/json" \
  -X POST "http://127.0.0.1:${local_port}/api/v1/settings/notifications/email" \
  -d '{"enabled":true,"types":0,"options":{"emailFrom":"noreply@example.com","senderName":"Seerr CI","smtpHost":"127.0.0.1","smtpPort":25}}'

api_key="${api_key//$'\r'/}"
api_key="${api_key//$'\n'/}"

printf 'SEERR_URL=%s\n' "http://127.0.0.1:${local_port}" >&3
printf 'SEERR_API_KEY=%s\n' "${api_key}" >&3
