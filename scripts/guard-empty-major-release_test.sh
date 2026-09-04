#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
guard="${repo_root}/scripts/guard-empty-major-release.sh"
test_root="$(mktemp -d)"
trap 'rm -rf "${test_root}"' EXIT

init_repo() {
  local dir="$1"
  git init -q "${dir}"
  git -C "${dir}" config user.email test@example.com
  git -C "${dir}" config user.name "Empty Major Guard Test"
}

assert_exit() {
  local name="$1"
  local expected="$2"
  local actual="$3"
  if [[ "${expected}" != "${actual}" ]]; then
    echo "FAIL ${name}: expected exit ${expected}, got ${actual}" >&2
    exit 1
  fi
  echo "PASS ${name}"
}

empty_major="${test_root}/empty-major"
init_repo "${empty_major}"
git -C "${empty_major}" commit -q --allow-empty -m "feat: initial"
git -C "${empty_major}" tag v2.0.0
git -C "${empty_major}" commit -q --allow-empty -m "chore(main): release 3.0.0"
set +e
(cd "${empty_major}" && PR_TITLE="chore(main): release 3.0.0" bash "${guard}" >/dev/null) 2>"${test_root}/empty.err"
status=$?
set -e
assert_exit "empty major is refused" 1 "${status}"
grep -Fq "Refusing empty major" "${test_root}/empty.err"

real_major="${test_root}/real-major"
init_repo "${real_major}"
git -C "${real_major}" commit -q --allow-empty -m "feat: initial"
git -C "${real_major}" tag v2.0.0
git -C "${real_major}" commit -q --allow-empty -m "$(printf 'feat!: drop the old schema\n\nBREAKING CHANGE: remove deprecated attribute\n')"
git -C "${real_major}" commit -q --allow-empty -m "chore(main): release 3.0.0"
set +e
(cd "${real_major}" && PR_TITLE="chore(main): release 3.0.0" bash "${guard}" >/dev/null)
status=$?
set -e
assert_exit "real major is allowed" 0 "${status}"

patch="${test_root}/patch"
init_repo "${patch}"
git -C "${patch}" commit -q --allow-empty -m "fix: typo"
git -C "${patch}" tag v2.0.0
git -C "${patch}" commit -q --allow-empty -m "chore(main): release 2.0.1"
set +e
(cd "${patch}" && PR_TITLE="chore(main): release 2.0.1" bash "${guard}" >/dev/null)
status=$?
set -e
assert_exit "patch title is allowed" 0 "${status}"
