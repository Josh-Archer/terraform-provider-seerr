#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
selector="${repo_root}/scripts/reconcile-release-tags.sh"
tip_guard="${repo_root}/scripts/release-tip-is-current.sh"
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

tip_remote="${test_root}/tip-remote.git"
git init -q --bare "${tip_remote}"
git -C "${fixture_repo}" remote add tip-test "${tip_remote}"
git -C "${fixture_repo}" branch -M main
git -C "${fixture_repo}" push -q tip-test main
original_tip="$(git -C "${fixture_repo}" rev-parse HEAD)"
output="$(cd "${fixture_repo}" && RELEASE_REMOTE=tip-test bash "${tip_guard}" "${original_tip}" main)"
assert_output "current default-branch tip is accepted" "true" "${output}"

git -C "${fixture_repo}" commit -q --allow-empty -m advance
git -C "${fixture_repo}" push -q tip-test main
output="$(cd "${fixture_repo}" && RELEASE_REMOTE=tip-test bash "${tip_guard}" "${original_tip}" main)"
assert_output "stale default-branch tip is rejected" "false" "${output}"

python3 - "${release_workflow}" "${test_workflow}" <<'PY'
from pathlib import Path
import sys


def step_block(path: Path, name: str) -> str:
    lines = path.read_text(encoding="utf-8").splitlines()
    marker = f"      - name: {name}"
    try:
        start = lines.index(marker)
    except ValueError as exc:
        raise SystemExit(f"missing workflow step: {name}") from exc
    end = len(lines)
    for index in range(start + 1, len(lines)):
        if lines[index].startswith("      - name:"):
            end = index
            break
    return "\n".join(lines[start:end])


release_path = Path(sys.argv[1])
test_path = Path(sys.argv[2])
release_text = release_path.read_text(encoding="utf-8")

qualified_head_guard = (
    "if: github.ref == format('refs/heads/{0}', "
    "github.event.repository.default_branch)"
)
if qualified_head_guard not in release_text:
    raise SystemExit("release job is not restricted to the fully qualified default-branch ref")

checkout = step_block(release_path, "Checkout tagged source")
if "ref: refs/tags/${{ env.RELEASE_TAG }}" not in checkout:
    raise SystemExit("tagged source checkout is not fully qualified")

precheck = step_block(test_path, "Verify current default-branch tip")
if "release-tip-is-current.sh" not in precheck:
    raise SystemExit("pre-tag default-tip behavior guard is missing")

bump = step_block(test_path, "Bump version and push tag")
if "if: steps.default_branch_tip.outputs.current == 'true'" not in bump:
    raise SystemExit("version allocation is not gated by the pre-tag tip check")

postcheck = step_block(test_path, "Verify tagged commit is still the default-branch tip")
for required in (
    "release-tip-is-current.sh",
    'git push origin ":refs/tags/${RELEASE_TAG}"',
    "tagged_commit=",
):
    if required not in postcheck:
        raise SystemExit(f"post-tag stale rollback is missing: {required}")

dispatch = step_block(test_path, "Trigger Release workflow for new tag")
if "if: steps.tagged_branch_tip.outputs.current == 'true'" not in dispatch:
    raise SystemExit("release dispatch is not gated by the post-tag tip check")

print("PASS release workflow structure")
PY
