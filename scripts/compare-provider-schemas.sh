#!/usr/bin/env bash
# Purpose: compare candidate provider schema against the previous stable release.
# Detects removed resources/data sources/attributes, type changes, newly required
# attributes, and lost sensitive markings. Produces a readable PR artifact.
#
# Usage:
#   bash ./scripts/compare-provider-schemas.sh [flags]
#
# Environment variables:
#   SEERR_SCHEMA_PREVIOUS_VERSION       Previous stable version (e.g. 2.2.2 or v2.2.2)
#   SEERR_SCHEMA_BASELINE_FILE          Explicit baseline schema JSON file
#   SEERR_SCHEMA_REPORT_PATH            File path to write the markdown PR artifact
#   SEERR_SCHEMA_OUT_CANDIDATE          File path to export normalized candidate schema
#   SEERR_SCHEMA_OUT_PREVIOUS           File path to export normalized baseline schema
#   SEERR_SCHEMA_TARGET_VERSION         Candidate version (e.g. v2.3.0 or v3.0.0)
#   SEERR_SCHEMA_MIGRATION_NOTE         File path to reviewed migration note
#   SEERR_SCHEMA_ALLOW_BREAKING=true    Allow breaking changes if policy/migration note pass
#   SEERR_SCHEMA_ALLOW_OFFLINE=true     Allow running without registry connection
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

resolve_tool() {
  local tool="$1"
  if command -v "$tool" >/dev/null 2>&1; then
    command -v "$tool"
    return 0
  fi
  if command -v "${tool}.exe" >/dev/null 2>&1; then
    command -v "${tool}.exe"
    return 0
  fi
  return 1
}

go_bin="$(resolve_tool go)" || {
  echo "Unable to find 'go' on PATH." >&2
  exit 1
}

cmd_args=()

if [[ -n "${SEERR_SCHEMA_PREVIOUS_VERSION:-}" ]]; then
  cmd_args+=("-previous" "${SEERR_SCHEMA_PREVIOUS_VERSION#v}")
fi

if [[ -n "${SEERR_SCHEMA_BASELINE_FILE:-}" ]]; then
  cmd_args+=("-baseline-file" "${SEERR_SCHEMA_BASELINE_FILE}")
elif [[ -f "${repo_root}/testdata/provider_schema_baseline.json" && -z "${SEERR_SCHEMA_PREVIOUS_VERSION:-}" ]]; then
  cmd_args+=("-baseline-file" "${repo_root}/testdata/provider_schema_baseline.json")
fi

report_file="${SEERR_SCHEMA_REPORT_PATH:-${repo_root}/test-artifacts/schema-comparison/schema-comparison-report.md}"
cmd_args+=("-report" "${report_file}")

cand_out="${SEERR_SCHEMA_OUT_CANDIDATE:-${repo_root}/test-artifacts/schema-comparison/candidate-schema.json}"
cmd_args+=("-output-candidate-schema" "${cand_out}")

prev_out="${SEERR_SCHEMA_OUT_PREVIOUS:-${repo_root}/test-artifacts/schema-comparison/previous-schema.json}"
cmd_args+=("-output-previous-schema" "${prev_out}")

if [[ -n "${SEERR_SCHEMA_TARGET_VERSION:-}" ]]; then
  cmd_args+=("-target-version" "${SEERR_SCHEMA_TARGET_VERSION}")
fi

if [[ -n "${SEERR_SCHEMA_MIGRATION_NOTE:-}" ]]; then
  cmd_args+=("-migration-note" "${SEERR_SCHEMA_MIGRATION_NOTE}")
fi

if [[ "${SEERR_SCHEMA_ALLOW_BREAKING:-false}" == "true" ]]; then
  cmd_args+=("-allow-breaking")
fi

if [[ "${SEERR_SCHEMA_ALLOW_OFFLINE:-true}" == "true" ]]; then
  cmd_args+=("-allow-offline")
fi

mkdir -p "$(dirname "${report_file}")"

"${go_bin}" run ./tools/schemacmp "${cmd_args[@]}" "$@"
