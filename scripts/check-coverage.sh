#!/usr/bin/env bash
# Purpose: enforce the deterministic unit-test statement coverage floor used by CI.
# Usage: bash ./scripts/check-coverage.sh
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

minimum="${SEERR_COVERAGE_MIN:-30.0}"
profile="$(mktemp "${TMPDIR:-/tmp}/seerr-coverage.XXXXXX")"
trap 'rm -f "${profile}"' EXIT

go test ./... -covermode=atomic -coverprofile="${profile}"

coverage="$(go tool cover -func="${profile}" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }')"
if [[ -z "${coverage}" ]]; then
  echo "Unable to determine total statement coverage." >&2
  exit 1
fi

echo "Total statement coverage: ${coverage}% (required: ${minimum}%)"
awk -v actual="${coverage}" -v minimum="${minimum}" 'BEGIN { exit !(actual + 0 >= minimum + 0) }' || {
  echo "Statement coverage ${coverage}% is below the required ${minimum}%." >&2
  exit 1
}
