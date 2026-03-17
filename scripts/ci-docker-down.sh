#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tests_dir="${script_dir}/../tests"

if ! command -v docker-compose >/dev/null 2>&1 && ! command -v docker >/dev/null 2>&1; then
  echo "docker or docker-compose is required" >&2
  exit 1
fi

compose_cmd="docker compose"
if command -v docker-compose >/dev/null 2>&1; then
  compose_cmd="docker-compose"
fi

cd "${tests_dir}"
echo "Tearing down Seerr test environment..."
${compose_cmd} down -v
