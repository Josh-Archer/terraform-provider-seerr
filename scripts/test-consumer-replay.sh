#!/usr/bin/env bash
# Purpose: replay a sanitized consumer configuration (test/consumer/arr-config) against a
# disposable Seerr and require a clean re-plan, including after a provider upgrade.
#   1. fresh apply with the provider under test, then `tofu plan -detailed-exitcode` == 0
#   2. destroy, apply with the previously released provider from the registry, switch to the
#      provider under test, `tofu plan -detailed-exitcode` == 0, apply, plan == 0 again
# Usage:
#   bash ./scripts/test-consumer-replay.sh
# Environment:
#   SEERR_URL / SEERR_API_KEY            use an existing disposable Seerr instead of starting Docker
#   SEERR_REPLAY_PREVIOUS_VERSION        previous released version (X.Y.Z); default: latest stable tag
#   SEERR_REPLAY_SKIP_UPGRADE=true       only run phase 1
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"
fixture_src="${repo_root}/test/consumer/arr-config"
artifact_dir="${SEERR_ARTIFACT_DIR:-${repo_root}/test-artifacts/consumer-replay}"
mirror_root="${SEERR_PROVIDER_MIRROR_ROOT:-${repo_root}/provider-mirror-consumer}"
current_version="99.99.99"
provider_target_dir="${mirror_root}/registry.opentofu.org/josh-archer/seerr/${current_version}/$(go env GOOS)_$(go env GOARCH)"
work_dir="${artifact_dir}/work"
local_env_created=false

run() {
  echo "+ $*" >&2
  "$@"
}

write_versions() {
  local version="$1"
  cat >"${work_dir}/versions.tf" <<VERSIONS
terraform {
  required_providers {
    seerr = {
      source  = "registry.opentofu.org/josh-archer/seerr"
      version = "${version}"
    }
  }
}
VERSIONS
}

use_cli_config() {
  local mode="$1" file="${artifact_dir}/tofurc-${1}"
  if [[ "${mode}" == "local" ]]; then
    cat >"${file}" <<CFG
provider_installation {
  filesystem_mirror {
    path    = "${mirror_root}"
    include = ["registry.opentofu.org/josh-archer/seerr"]
  }
}
CFG
  else
    cat >"${file}" <<CFG
provider_installation {
  direct {}
}
CFG
  fi
  export TF_CLI_CONFIG_FILE="${file}"
}

init_with() {
  local mode="$1" version="$2"
  write_versions "${version}"
  use_cli_config "${mode}"
  rm -rf "${work_dir}/.terraform" "${work_dir}/.terraform.lock.hcl"
  (cd "${work_dir}" && run tofu init -input=false -no-color)
}

tf() {
  local sub="$1"
  shift
  (cd "${work_dir}" && run tofu "${sub}" -input=false -no-color "$@")
}

# Fails unless the plan is empty (exit 0). Exit 2 means drift or an inconsistent re-plan.
assert_clean_plan() {
  local label="$1" rc=0
  (cd "${work_dir}" && run tofu plan -detailed-exitcode -input=false -no-color) || rc=$?
  if [[ ${rc} -ne 0 ]]; then
    echo "consumer replay FAILED: '${label}' plan exited ${rc} (expected 0, no changes)" >&2
    return 1
  fi
  echo "consumer replay OK: ${label} plan is clean" >&2
}

# Newest stable tag that is not the commit under test and is published in the registry.
# On a release commit the newest tag points at HEAD and is not in the registry yet.
previous_version() {
  if [[ -n "${SEERR_REPLAY_PREVIOUS_VERSION:-}" ]]; then
    printf '%s\n' "${SEERR_REPLAY_PREVIOUS_VERSION#v}"
    return 0
  fi
  local head_sha published sha version
  head_sha="$(git rev-parse HEAD)"
  published="$(curl -fsSL --retry 3 "${SEERR_REPLAY_REGISTRY_VERSIONS_URL:-https://registry.opentofu.org/v1/providers/josh-archer/seerr/versions}" \
    | grep -oE '"version" *: *"[^"]+"' | sed -E 's/.*"([^"]+)"$/\1/')" || published=""
  # Without --refs, annotated tags also list a peeled "^{}" line carrying the commit SHA.
  git ls-remote --tags "${SEERR_REPLAY_REPO_URL:-https://github.com/Josh-Archer/terraform-provider-seerr.git}" 'v*' \
    | awk '{ sub("refs/tags/v", "", $2); sub("\\^\\{\\}$", "", $2); sha[$2] = $1 } END { for (v in sha) print sha[v], v }' \
    | grep -E ' [0-9]+\.[0-9]+\.[0-9]+$' \
    | sort -k2,2Vr \
    | while read -r sha version; do
        [[ "${sha}" == "${head_sha}" ]] && continue
        if [[ -n "${published}" ]] && ! grep -qxF "${version}" <<<"${published}"; then
          continue
        fi
        printf '%s\n' "${version}"
        break
      done
}

cleanup() {
  local exit_code=$?
  if [[ "${local_env_created}" == "true" ]]; then
    bash "${repo_root}/scripts/ci-docker-down.sh" || true
  fi
  exit $exit_code
}
trap cleanup EXIT

command -v tofu >/dev/null || { echo "tofu is required" >&2; exit 1; }

rm -rf "${work_dir}"
mkdir -p "${artifact_dir}" "${provider_target_dir}"
cp -R "${fixture_src}" "${work_dir}"
rm -f "${work_dir}/README.md"

run go build -o "${provider_target_dir}/terraform-provider-seerr" .

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

export TF_VAR_seerr_url="${SEERR_URL}"
export TF_VAR_seerr_api_key="${SEERR_API_KEY}"
export TF_IN_AUTOMATION=1

echo "== Phase 1: fresh apply with provider under test, then clean re-plan" >&2
init_with local "${current_version}"
tf apply -auto-approve
assert_clean_plan "fresh apply (current provider)"

if [[ "${SEERR_REPLAY_SKIP_UPGRADE:-false}" == "true" ]]; then
  exit 0
fi

echo "== Phase 2: previous release apply, provider upgrade, clean re-plan" >&2
prev="$(previous_version)"
if [[ -z "${prev}" ]]; then
  echo "Could not determine the previously released provider version" >&2
  exit 1
fi
echo "Previous released version: ${prev}" >&2
tf destroy -auto-approve
init_with direct "${prev}"
tf apply -auto-approve
init_with local "${current_version}"
assert_clean_plan "after provider upgrade ${prev} -> current (no apply)"
tf apply -auto-approve
assert_clean_plan "apply after provider upgrade"
