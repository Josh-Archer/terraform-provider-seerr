#!/usr/bin/env bash
# Purpose: repo fast gate for generated files, build, unit tests, and optional lint/integration.
# Usage:
#   bash ./scripts/test-all-locally.sh
#   SEERR_RUN_INTEGRATION=true bash ./scripts/test-all-locally.sh
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

run() {
  echo "+ $*" >&2
  "$@"
}

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

  if command -v powershell.exe >/dev/null 2>&1; then
    local resolved
    resolved="$(powershell.exe -NoProfile -Command "(Get-Command $tool).Source" 2>/dev/null | tr -d '\r')"
    if [[ -n "${resolved}" ]]; then
      printf '%s\n' "${resolved}"
      return 0
    fi
  fi

  return 1
}

go_bin="$(resolve_tool go)" || {
  echo "Unable to find 'go' on PATH." >&2
  exit 1
}

run bash "${repo_root}/tools/check-generated.sh"
run "${go_bin}" build ./...
run "${go_bin}" test ./...

case "${SEERR_RUN_LINT:-auto}" in
  true)
    golangci_lint_bin="$(resolve_tool golangci-lint)" || {
      echo "golangci-lint is required when SEERR_RUN_LINT=true" >&2
      exit 1
    }
    run "${golangci_lint_bin}" run
    ;;
  auto)
    if golangci_lint_bin="$(resolve_tool golangci-lint)"; then
      run "${golangci_lint_bin}" run
    else
      echo "Skipping golangci-lint; command not found on PATH" >&2
    fi
    ;;
  false)
    echo "Skipping golangci-lint because SEERR_RUN_LINT=false" >&2
    ;;
  *)
    echo "Invalid SEERR_RUN_LINT value: ${SEERR_RUN_LINT}" >&2
    exit 1
    ;;
esac

if [[ "${SEERR_RUN_INTEGRATION:-false}" == "true" ]]; then
  run bash "${repo_root}/scripts/test-integration.sh"
fi
