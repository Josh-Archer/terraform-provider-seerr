#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
selector="${repo_root}/scripts/reconcile-release-tags.sh"
fake_bin="${repo_root}/scripts/testdata"
test_root="$(mktemp -d)"
fixture_repo="${test_root}/repo"
trap 'rm -rf "${test_root}"' EXIT

git init -q "${fixture_repo}"
git -C "${fixture_repo}" config user.email test@example.com
git -C "${fixture_repo}" config user.name "Release Reconciler Test"
git -C "${fixture_repo}" commit -q --allow-empty -m fixture
for tag in \
  v0.19.2 \
  v0.20.0-test.1 \
  v0.20.5 \
  v0.20.6 \
  v0.20.7 \
  v0.20.9 \
  v0.20.10; do
  git -C "${fixture_repo}" tag "${tag}"
done

select_tag() {
  (
    cd "${fixture_repo}"
    env \
      PATH="${fake_bin}:${PATH}" \
      GITHUB_REPOSITORY="example/repository" \
      RELEASE_RECONCILE_FROM_TAG="${RELEASE_RECONCILE_FROM_TAG:-v0.20.5}" \
      RELEASE_TAG_OVERRIDE="${RELEASE_TAG_OVERRIDE:-}" \
      MOCK_PUBLISHED_TAGS="${MOCK_PUBLISHED_TAGS:-}" \
      MOCK_DRAFT_TAGS="${MOCK_DRAFT_TAGS:-}" \
      MOCK_ACTIVE_TAGS="${MOCK_ACTIVE_TAGS:-}" \
      MOCK_API_ERROR_TAG="${MOCK_API_ERROR_TAG:-}" \
      MOCK_RUN_API_ERROR="${MOCK_RUN_API_ERROR:-false}" \
      bash "${selector}"
  )
}

assert_output() {
  local name="$1"
  local expected="$2"
  local actual="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    echo "FAIL ${name}: expected '${expected}', got '${actual}'" >&2
    exit 1
  fi
  echo "PASS ${name}"
}

assert_file_contains() {
  local name="$1"
  local file="$2"
  local expected="$3"

  if ! grep -Fq -- "${expected}" "${file}"; then
    echo "FAIL ${name}: '${expected}' not found in ${file}" >&2
    exit 1
  fi
  echo "PASS ${name}"
}

output="$(MOCK_PUBLISHED_TAGS=v0.20.7 select_tag)"
assert_output "finds hole before later release" "v0.20.5|false" "${output}"

output="$(MOCK_PUBLISHED_TAGS=v0.20.5,v0.20.7 select_tag)"
assert_output "selects next hole" "v0.20.6|false" "${output}"

output="$(MOCK_PUBLISHED_TAGS=v0.20.7 MOCK_ACTIVE_TAGS=v0.20.5 select_tag)"
assert_output "reports active run" "v0.20.5|true" "${output}"

output="$(MOCK_PUBLISHED_TAGS=v0.20.5,v0.20.6,v0.20.7,v0.20.9,v0.20.10 select_tag)"
assert_output "no orphan" "|false" "${output}"

output="$(MOCK_PUBLISHED_TAGS=v0.20.5,v0.20.6,v0.20.7 select_tag)"
assert_output "version ordering" "v0.20.9|false" "${output}"

output="$(RELEASE_TAG_OVERRIDE=v0.19.2 MOCK_PUBLISHED_TAGS=v0.20.7 select_tag)"
assert_output "historical override bypasses floor" "v0.19.2|false" "${output}"

output="$(RELEASE_TAG_OVERRIDE=v0.20.0-test.1 select_tag)"
assert_output "explicit prerelease override" "v0.20.0-test.1|false" "${output}"

output="$(RELEASE_TAG_OVERRIDE=v0.19.2 MOCK_PUBLISHED_TAGS=v0.19.2 select_tag)"
assert_output "published override is a no-op" "|false" "${output}"

output="$(MOCK_DRAFT_TAGS=v0.20.5 MOCK_PUBLISHED_TAGS=v0.20.7 select_tag)"
assert_output "draft release is retryable" "v0.20.5|false" "${output}"

if MOCK_API_ERROR_TAG=v0.20.5 select_tag >"${test_root}/unexpected.out" 2>"${test_root}/api.err"; then
  echo "FAIL release API failure fails closed: command succeeded unexpectedly" >&2
  exit 1
fi
grep -Fq "simulated API failure" "${test_root}/api.err"
echo "PASS release API failure fails closed"

if MOCK_RUN_API_ERROR=true select_tag >"${test_root}/unexpected.out" 2>"${test_root}/run-api.err"; then
  echo "FAIL run API failure fails closed: command succeeded unexpectedly" >&2
  exit 1
fi
grep -Fq "simulated run API failure" "${test_root}/run-api.err"
echo "PASS run API failure fails closed"

if RELEASE_TAG_OVERRIDE='v0.20.5^{commit}' select_tag >"${test_root}/unexpected.out" 2>"${test_root}/invalid.err"; then
  echo "FAIL invalid override: command succeeded unexpectedly" >&2
  exit 1
fi
grep -Fq "Invalid release tag" "${test_root}/invalid.err"
echo "PASS invalid override"

if RELEASE_RECONCILE_FROM_TAG=v9.9.9 select_tag >"${test_root}/unexpected.out" 2>"${test_root}/floor.err"; then
  echo "FAIL missing floor: command succeeded unexpectedly" >&2
  exit 1
fi
grep -Fq "Release tag does not exist" "${test_root}/floor.err"
echo "PASS missing floor"

release_workflow="${repo_root}/.github/workflows/release.yml"
test_workflow="${repo_root}/.github/workflows/test.yml"
# shellcheck disable=SC2016 # Literal GitHub expression under test.
assert_file_contains \
  "tag checkout cannot resolve a same-named branch" \
  "${release_workflow}" \
  'ref: refs/tags/${{ env.RELEASE_TAG }}'
assert_file_contains \
  "release runs only from the default branch" \
  "${release_workflow}" \
  "if: github.ref_name == github.event.repository.default_branch"
# shellcheck disable=SC2016 # Literal shell expression embedded in the workflow.
assert_file_contains \
  "stale push guard fetches the current default-branch tip" \
  "${test_workflow}" \
  'refs/heads/${DEFAULT_BRANCH}:refs/remotes/origin/${DEFAULT_BRANCH}'
assert_file_contains \
  "stale push guard controls version allocation" \
  "${test_workflow}" \
  "if: steps.default_branch_tip.outputs.current == 'true'"
