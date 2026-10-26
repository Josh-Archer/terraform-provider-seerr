#!/usr/bin/env bash
# Purpose: replay a sanitized consumer configuration (test/consumer/arr-config) against a
# disposable Seerr and require a clean re-plan, including after a provider upgrade.
#   1. fresh apply with the locally built provider (filesystem mirror), then
#      `tofu plan -detailed-exitcode` == 0
#   2. destroy, apply with the previously published provider from the registry,
#      switch back to the local build, `tofu plan -detailed-exitcode` == 0, apply, plan == 0
# The version under test is never installed from the registry. It is published only
# after this replay has passed and the Release workflow promotes the tag.
# Usage:
#   bash ./scripts/test-consumer-replay.sh
# Environment:
#   SEERR_URL / SEERR_API_KEY            use an existing disposable Seerr instead of starting Docker
#   SEERR_REPLAY_PREVIOUS_VERSION        previous released version (X.Y.Z); default: newest published stable tag older than the candidate
#   SEERR_REPLAY_RELEASE_VERSION         candidate release (X.Y.Z). Versions greater than or equal to it are not a valid upgrade baseline.
#   SEERR_REPLAY_SKIP_UPGRADE=true       only run phase 1
set -euo pipefail

# Sentinel version for the filesystem mirror. It must not exist in the public registry.
current_version="99.99.99"

version_lt() {
  local left right sorted oldest
  left="${1#v}"
  right="${2#v}"
  # Avoid `sort | head`: under pipefail, SIGPIPE makes that pipeline flaky.
  sorted="$(printf '%s\n%s\n' "${left}" "${right}" | LC_ALL=C sort -V)"
  oldest="${sorted%%$'\n'*}"
  [[ "${left}" != "${right}" && "${oldest}" == "${left}" ]]
}

# stdin: `git ls-remote --tags` lines. Stable X.Y.Z only.
# Prefer the peeled commit SHA (^{}) over an annotated tag object SHA so HEAD
# comparison still works when the tag object line is listed last.
parse_stable_tag_shas() {
  awk '
    {
      peeled = ($2 ~ /\^\{\}$/)
      name = $2
      sub("^refs/tags/v", "", name)
      sub("\\^\\{\\}$", "", name)
      if (name !~ /^[0-9]+\.[0-9]+\.[0-9]+$/) next
      if (peeled || !(name in sha)) sha[name] = $1
    }
    END {
      for (v in sha) print sha[v], v
    }
  '
}

# Newest published stable version strictly older than the release under test.
# Callers install that baseline from the registry. The candidate build is separate.
previous_version() {
  local explicit
  explicit="${SEERR_REPLAY_PREVIOUS_VERSION:-}"
  if [[ -n "${explicit}" ]]; then
    printf '%s\n' "${explicit#v}"
    return 0
  fi

  local head_sha published upper tags sha version best
  head_sha="$(git rev-parse HEAD)"
  if ! published="$(curl -fsSL --retry 3 "${SEERR_REPLAY_REGISTRY_VERSIONS_URL:-https://registry.opentofu.org/v1/providers/josh-archer/seerr/versions}" \
    | grep -oE '"version" *: *"[^"]+"' \
    | sed -E 's/.*"([^"]+)"$/\1/')"; then
    echo "consumer replay FAILED: could not list published provider versions from the registry" >&2
    return 1
  fi
  if [[ -z "${published}" ]]; then
    echo "consumer replay FAILED: registry returned no provider versions" >&2
    return 1
  fi

  if ! tags="$(git ls-remote --tags "${SEERR_REPLAY_REPO_URL:-https://github.com/Josh-Archer/terraform-provider-seerr.git}" 'v*' \
    | parse_stable_tag_shas)"; then
    echo "consumer replay FAILED: could not list release tags" >&2
    return 1
  fi

  upper="${SEERR_REPLAY_RELEASE_VERSION:-}"
  upper="${upper#v}"
  if [[ -z "${upper}" ]]; then
    while read -r sha version; do
      [[ -n "${version}" ]] || continue
      if [[ "${sha}" == "${head_sha}" ]] && { [[ -z "${upper}" ]] || version_lt "${upper}" "${version}"; }; then
        upper="${version}"
      fi
    done <<< "${tags}"
  fi

  best=""
  while read -r sha version; do
    [[ -n "${version}" ]] || continue
    if [[ -n "${upper}" ]] && ! version_lt "${version}" "${upper}"; then
      continue
    fi
    if [[ "${sha}" == "${head_sha}" ]]; then
      continue
    fi
    if ! grep -qxF "${version}" <<< "${published}"; then
      continue
    fi
    if [[ -z "${best}" ]] || version_lt "${best}" "${version}"; then
      best="${version}"
    fi
  done <<< "${tags}"

  if [[ -z "${best}" ]]; then
    echo "consumer replay FAILED: no published stable provider version is older than the candidate" >&2
    return 1
  fi
  printf '%s\n' "${best}"
}

# The registry copy of the release under test does not exist until promotion.
reject_registry_version() {
  local version release
  version="$1"
  if [[ "${version}" == "${current_version}" ]]; then
    echo "consumer replay FAILED: refusing to install candidate sentinel ${version} from the registry" >&2
    return 1
  fi
  release="${SEERR_REPLAY_RELEASE_VERSION:-}"
  release="${release#v}"
  if [[ -n "${release}" && "${version}" == "${release}" ]]; then
    echo "consumer replay FAILED: refusing to install release ${version} from the registry; it is published only after promotion" >&2
    return 1
  fi
}

use_cli_config() {
  local mode="$1" file="${artifact_dir}/tofurc-${1}"
  if [[ "${mode}" == "local" ]]; then
    # filesystem_mirror replaces the implicit direct method. The candidate
    # cannot fall through to the public registry.
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

installed_provider_bins() {
  if [[ ! -d "${work_dir}/.terraform/providers" ]]; then
    return 0
  fi
  find "${work_dir}/.terraform/providers" -type f -name 'terraform-provider-seerr*' ! -name '*.json' -print
}

assert_local_candidate() {
  local built installed matched
  built="${provider_target_dir}/terraform-provider-seerr"
  matched=false
  while IFS= read -r installed; do
    [[ -n "${installed}" ]] || continue
    if cmp -s "${built}" "${installed}"; then
      matched=true
    fi
  done < <(installed_provider_bins)
  if [[ "${matched}" != true ]]; then
    echo "consumer replay FAILED: candidate provider was not installed from the local build (${built})" >&2
    return 1
  fi
  echo "consumer replay OK: candidate provider is the local build" >&2
}

assert_registry_baseline() {
  local built installed found
  built="${provider_target_dir}/terraform-provider-seerr"
  found=false
  while IFS= read -r installed; do
    [[ -n "${installed}" ]] || continue
    found=true
    if cmp -s "${built}" "${installed}"; then
      echo "consumer replay FAILED: previous provider matches the candidate build; it was not installed from the registry" >&2
      return 1
    fi
  done < <(installed_provider_bins)
  if [[ "${found}" != true ]]; then
    echo "consumer replay FAILED: previous provider binary was not installed" >&2
    return 1
  fi
  echo "consumer replay OK: previous provider is the published registry build" >&2
}

cleanup() {
  local exit_code=$?
  if [[ "${local_env_created}" == "true" ]]; then
    bash "${repo_root}/scripts/ci-docker-down.sh" || true
  fi
  exit "${exit_code}"
}

main() {
  repo_root="$(git rev-parse --show-toplevel)"
  cd "${repo_root}"
  local fixture_src docker_env line prev
  fixture_src="${repo_root}/test/consumer/arr-config"
  artifact_dir="${SEERR_ARTIFACT_DIR:-${repo_root}/test-artifacts/consumer-replay}"
  mirror_root="${SEERR_PROVIDER_MIRROR_ROOT:-${repo_root}/provider-mirror-consumer}"
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

  init_with() {
    local mode="$1" version="$2"
    if [[ "${mode}" == "direct" ]]; then
      reject_registry_version "${version}"
    fi
    write_versions "${version}"
    use_cli_config "${mode}"
    rm -rf "${work_dir}/.terraform" "${work_dir}/.terraform.lock.hcl"
    (cd "${work_dir}" && run tofu init -input=false -no-color)
    if [[ "${mode}" == "local" ]]; then
      assert_local_candidate
    else
      assert_registry_baseline
    fi
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
      # shellcheck disable=SC2163 # line is KEY=value from ci-docker-up.sh, not a variable name
      export "${line?}"
    done <<< "${docker_env}"
  fi

  export TF_VAR_seerr_url="${SEERR_URL}"
  export TF_VAR_seerr_api_key="${SEERR_API_KEY}"
  export TF_IN_AUTOMATION=1

  echo "== Phase 1: fresh apply with local candidate build, then clean re-plan" >&2
  init_with local "${current_version}"
  tf apply -auto-approve
  assert_clean_plan "fresh apply (current provider)"

  if [[ "${SEERR_REPLAY_SKIP_UPGRADE:-false}" == "true" ]]; then
    exit 0
  fi

  echo "== Phase 2: previous published release, upgrade to local candidate, clean re-plan" >&2
  if ! prev="$(previous_version)"; then
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
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
