#!/usr/bin/env bash
# Run the existing unit suite once, optionally recording coverage for CI evidence.
set -euo pipefail
go_bin="${GO_BIN:-go}"
if [[ -z "${SEERR_COVERAGE_DIR:-}" ]]; then
  exec "$go_bin" test ./...
fi
mkdir -p "$SEERR_COVERAGE_DIR"
"$go_bin" test ./... -coverprofile="$SEERR_COVERAGE_DIR/coverage.out"
"$go_bin" tool cover -func="$SEERR_COVERAGE_DIR/coverage.out" > "$SEERR_COVERAGE_DIR/functions.txt"
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    printf '\n### Offline Go unit-test coverage\n\n```text\n'
    tail -n 1 "$SEERR_COVERAGE_DIR/functions.txt"
    printf '```\n\nFunction-level coverage and the profile are attached as unit-test-coverage.\n'
  } >> "$GITHUB_STEP_SUMMARY"
fi
