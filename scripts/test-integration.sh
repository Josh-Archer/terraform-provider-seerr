#!/usr/bin/env bash
# Purpose: run OpenTofu integration coverage against either the stable merge gate or the broader full suite.
# Usage:
#   bash ./scripts/test-integration.sh
#   SEERR_TEST_SUITE=stable bash ./scripts/test-integration.sh
#   SEERR_TEST_SUITE=all bash ./scripts/test-integration.sh
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"
tests_dir="${repo_root}/tests"
artifact_dir="${SEERR_ARTIFACT_DIR:-${repo_root}/test-artifacts/integration}"
mirror_root="${SEERR_PROVIDER_MIRROR_ROOT:-${repo_root}/provider-mirror}"
provider_namespace="${SEERR_PROVIDER_NAMESPACE:-registry.opentofu.org/josh-archer/seerr/99.99.99}"
test_suite="${SEERR_TEST_SUITE:-stable}"
tofu_test_log="${artifact_dir}/tofu-test.log"
compose_log="${artifact_dir}/docker-compose.log"
tofu_cli_config="${artifact_dir}/tofurc"
local_env_created=false
generated_provider_files=()

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

tofu_bin="$(resolve_tool tofu)" || {
  echo "Unable to find 'tofu' on PATH." >&2
  exit 1
}

goos="$("${go_bin}" env GOOS)"
goarch="$("${go_bin}" env GOARCH)"
goexe="$("${go_bin}" env GOEXE)"
provider_target_dir="${mirror_root}/${provider_namespace}/${goos}_${goarch}"
provider_binary="${provider_target_dir}/terraform-provider-seerr${goexe}"

mkdir -p "${artifact_dir}" "${provider_target_dir}"
cat >"${tofu_cli_config}" <<EOF
provider_installation {
  direct {}
}
EOF
export TF_CLI_CONFIG_FILE="${tofu_cli_config}"

capture_diagnostics() {
  if [[ "${local_env_created}" != "true" ]]; then
    return 0
  fi

  if command -v docker-compose >/dev/null 2>&1; then
    docker-compose -f "${tests_dir}/docker-compose.yml" logs --no-color >"${compose_log}" 2>&1 || true
  elif command -v docker >/dev/null 2>&1; then
    docker compose -f "${tests_dir}/docker-compose.yml" logs --no-color >"${compose_log}" 2>&1 || true
  fi
}

cleanup_generated_provider_files() {
  local generated_file
  for generated_file in "${generated_provider_files[@]}"; do
    rm -f "${generated_file}"
  done
}

generate_test_provider_configs() {
  local module_dir generated_file escaped_url escaped_api_key

  escaped_url="${SEERR_URL//\\/\\\\}"
  escaped_url="${escaped_url//\"/\\\"}"
  escaped_api_key="${SEERR_API_KEY//\\/\\\\}"
  escaped_api_key="${escaped_api_key//\"/\\\"}"

  while IFS= read -r module_dir; do
    generated_file="${module_dir}/zz_test_provider.auto.tf"
    cat >"${generated_file}" <<EOF
provider "seerr" {
  url     = "${escaped_url}"
  api_key = "${escaped_api_key}"
}
EOF
    generated_provider_files+=("${generated_file}")
  done < <(find "${tests_dir}/modules" -mindepth 1 -maxdepth 1 -type d | sort)
}

cleanup() {
  local exit_code=$?

  if [[ $exit_code -ne 0 ]]; then
    capture_diagnostics
  fi

  cleanup_generated_provider_files

  if [[ "${local_env_created}" == "true" ]]; then
    bash "${repo_root}/scripts/ci-docker-down.sh" || true
  fi

  exit $exit_code
}

trap cleanup EXIT

run "${go_bin}" build -v -o "${provider_binary}" .

if [[ -z "${SEERR_URL:-}" || -z "${SEERR_API_KEY:-}" ]]; then
  local_env_created=true
  docker_env="$(bash "${repo_root}/scripts/ci-docker-up.sh")" || {
    echo "Failed to start the local Seerr test target" >&2
    exit 1
  }

  while IFS= read -r line; do
    [[ -n "${line}" ]] || continue
    export "${line}"
  done <<< "${docker_env}"
fi

cd "${tests_dir}"
rm -rf .terraform
rm -f .terraform.lock.hcl

export TF_VAR_url="${SEERR_URL}"
export TF_VAR_api_key="${SEERR_API_KEY}"
generate_test_provider_configs

run "${tofu_bin}" init -plugin-dir="${mirror_root}"

tofu_test_args=()
if [[ -n "${SEERR_TOFU_FILTERS:-}" ]]; then
  for filter in ${SEERR_TOFU_FILTERS}; do
    tofu_test_args+=("-filter=${filter}")
  done
elif [[ "${test_suite}" == "stable" ]]; then
  stable_filters=(
    "api_objects.tftest.hcl"
    "current_user.tftest.hcl"
    "data_sources.tftest.hcl"
    "discover.tftest.hcl"
    "discover_slider_data_source.tftest.hcl"
    "features.tftest.hcl"
    "settings.tftest.hcl"
    "user.tftest.hcl"
    "user_permissions.tftest.hcl"
  )
  for filter in "${stable_filters[@]}"; do
    tofu_test_args+=("-filter=${filter}")
  done
elif [[ "${test_suite}" != "all" ]]; then
  echo "Unsupported SEERR_TEST_SUITE value: ${test_suite}" >&2
  exit 1
fi

run "${tofu_bin}" test \
  "${tofu_test_args[@]}" 2>&1 | tee "${tofu_test_log}"
