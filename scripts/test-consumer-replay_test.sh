#!/usr/bin/env bash
# Selects the upgrade baseline without Docker or a live registry.
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
# shellcheck source=scripts/test-consumer-replay.sh
source "${repo_root}/scripts/test-consumer-replay.sh"

test_root="$(mktemp -d)"
trap 'rm -rf "${test_root}"' EXIT

pass() {
  echo "PASS $1"
}

fail() {
  echo "FAIL $1" >&2
  exit 1
}

expect_eq() {
  local name="$1" expected="$2" actual="$3"
  if [[ "${expected}" != "${actual}" ]]; then
    fail "${name}: expected '${expected}', got '${actual}'"
  fi
  pass "${name}"
}

init_repo() {
  local dir="$1"
  git init -q -b main "${dir}"
  git -C "${dir}" config user.email test@example.com
  git -C "${dir}" config user.name "Consumer Replay Test"
  git -C "${dir}" commit -q --allow-empty -m "base"
  git -C "${dir}" tag -a v2.2.2 -m "2.2.2"
  git -C "${dir}" commit -q --allow-empty -m "release 2.3.0"
  git -C "${dir}" tag -a v2.3.0 -m "2.3.0"
  git -C "${dir}" commit -q --allow-empty -m "release 2.3.1"
  git -C "${dir}" tag -a v2.3.1 -m "2.3.1"
  git -C "${dir}" commit -q --allow-empty -m "dev"
}

write_registry() {
  local file="$1"
  shift
  {
    echo '{"versions":['
    local first=true version
    for version in "$@"; do
      if [[ "${first}" == true ]]; then
        first=false
      else
        echo ','
      fi
      printf '  {"version":"%s"}' "${version}"
    done
    echo '] }'
  } >"${file}"
}

checkout_tag() {
  git -C "$1" checkout -q "$2"
}

select_previous() {
  local dir="$1" registry="$2"
  shift 2
  (
    cd "${dir}"
    unset SEERR_REPLAY_PREVIOUS_VERSION SEERR_REPLAY_RELEASE_VERSION
    export SEERR_REPLAY_REGISTRY_VERSIONS_URL="file://${registry}"
    export SEERR_REPLAY_REPO_URL="${dir}"
    while [[ $# -gt 0 ]]; do
      # shellcheck disable=SC2163 # argument is NAME=value
      export "${1?}"
      shift
    done
    previous_version
  )
}

expect_previous() {
  local name="$1" expected="$2"
  shift 2
  local actual
  if ! actual="$(select_previous "$@")"; then
    fail "${name}: previous_version failed"
  fi
  expect_eq "${name}" "${expected}" "${actual}"
}

expect_previous_fails() {
  local name="$1"
  shift
  local err="${test_root}/err"
  if select_previous "$@" >"${test_root}/out" 2>"${err}"; then
    fail "${name}: expected previous_version to fail"
  fi
  if ! grep -q "consumer replay FAILED" "${err}"; then
    fail "${name}: missing failure message: $(cat "${err}")"
  fi
  pass "${name}"
}

repo="${test_root}/repo"
init_repo "${repo}"
dev_sha="$(git -C "${repo}" rev-parse HEAD)"
registry="${test_root}/registry.json"

parsed="$(git -C "${repo}" ls-remote --tags "${repo}" 'v*' | parse_stable_tag_shas | LC_ALL=C sort -k2,2)"
expect_eq "annotated tag uses peeled commit SHA" \
  "$(git -C "${repo}" rev-parse 'v2.3.0^{}') 2.3.0" \
  "$(printf '%s\n' "${parsed}" | grep ' 2.3.0$')"

both_orders="$(
  {
    printf '%s\n' "COMMIT refs/tags/v2.3.0^{}" "TAGOBJ refs/tags/v2.3.0"
    printf '%s\n' "TAGOBJ refs/tags/v2.2.2" "COMMIT refs/tags/v2.2.2^{}"
  } | parse_stable_tag_shas | LC_ALL=C sort -k2,2
)"
expect_eq "peeled SHA wins regardless of listing order" \
  "$(printf '%s\n' "COMMIT 2.2.2" "COMMIT 2.3.0")" \
  "${both_orders}"

write_registry "${registry}" 2.2.2 2.3.1 2.2.3-rc.1
checkout_tag "${repo}" v2.3.0
expect_previous "release commit skips unpublished HEAD and newer published tags" 2.2.2 "${repo}" "${registry}"

write_registry "${registry}" 2.2.2 2.3.0 2.3.1
checkout_tag "${repo}" v2.3.0
expect_previous "published tag on HEAD is not the upgrade baseline" 2.2.2 "${repo}" "${registry}"

write_registry "${registry}" 2.2.2 2.3.1
checkout_tag "${repo}" "${dev_sha}"
expect_previous "untagged HEAD uses the newest published stable release" 2.3.1 "${repo}" "${registry}"

checkout_tag "${repo}" "${dev_sha}"
expect_previous "explicit release bound ignores newer published versions" 2.2.2 "${repo}" "${registry}" \
  SEERR_REPLAY_RELEASE_VERSION=v2.3.0

checkout_tag "${repo}" "${dev_sha}"
actual="$(
  cd "${repo}"
  SEERR_REPLAY_PREVIOUS_VERSION=v2.1.2 \
    SEERR_REPLAY_REGISTRY_VERSIONS_URL="file://${test_root}/missing-registry.json" \
    SEERR_REPLAY_REPO_URL="${repo}" \
    previous_version
)"
expect_eq "explicit previous version skips the registry" 2.1.2 "${actual}"

expect_previous_fails "registry outage fails closed" "${repo}" "${test_root}/missing-registry.json"

if version_lt 2.9.0 2.10.0 && ! version_lt 2.10.0 2.9.0 && ! version_lt 2.3.0 2.3.0; then
  pass "version order"
else
  fail "version order"
fi

(
  unset SEERR_REPLAY_RELEASE_VERSION
  reject_registry_version 2.2.2
)
pass "published baseline may come from the registry"

if reject_registry_version 99.99.99 2>"${test_root}/sentinel.err"; then
  fail "sentinel version must not be installed from the registry"
fi
grep -q "candidate sentinel" "${test_root}/sentinel.err"
pass "candidate sentinel is refused from the registry"

if SEERR_REPLAY_RELEASE_VERSION=2.3.0 reject_registry_version 2.3.0 2>"${test_root}/release.err"; then
  fail "release version must not be installed from the registry"
fi
grep -q "published only after promotion" "${test_root}/release.err"
pass "unreleased version is refused from the registry"

artifact_dir="${test_root}/artifacts"
mirror_root="${test_root}/mirror"
mkdir -p "${artifact_dir}"
use_cli_config local
if grep -q 'direct' "${TF_CLI_CONFIG_FILE}"; then
  fail "local install config fell through to the registry"
fi
grep -q 'filesystem_mirror' "${TF_CLI_CONFIG_FILE}"
pass "candidate install is filesystem mirror only"

use_cli_config direct
if grep -q 'filesystem_mirror' "${TF_CLI_CONFIG_FILE}"; then
  fail "registry install config used the local mirror"
fi
grep -q 'direct' "${TF_CLI_CONFIG_FILE}"
pass "upgrade baseline install is the public registry"

provider_target_dir="${test_root}/built"
work_dir="${test_root}/tfwork"
mkdir -p "${provider_target_dir}" "${work_dir}/.terraform/providers/local"
echo 'local-build' >"${provider_target_dir}/terraform-provider-seerr"
echo 'local-build' >"${work_dir}/.terraform/providers/local/terraform-provider-seerr"
assert_local_candidate >/dev/null
pass "local candidate bytes are accepted"

echo 'registry-build' >"${work_dir}/.terraform/providers/local/terraform-provider-seerr"
if assert_local_candidate 2>"${test_root}/bytes.err"; then
  fail "mismatched candidate bytes were accepted"
fi
grep -q "not installed from the local build" "${test_root}/bytes.err"
pass "registry bytes are not a local candidate"

assert_registry_baseline >/dev/null
pass "registry baseline bytes differ from the candidate"

echo 'local-build' >"${work_dir}/.terraform/providers/local/terraform-provider-seerr"
if assert_registry_baseline 2>"${test_root}/same.err"; then
  fail "candidate bytes were accepted as the registry baseline"
fi
grep -q "not installed from the registry" "${test_root}/same.err"
pass "candidate bytes are not a registry baseline"
