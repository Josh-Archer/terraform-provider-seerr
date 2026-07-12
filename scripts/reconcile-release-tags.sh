#!/usr/bin/env bash
# Print the oldest tag that does not have a GitHub Release.
# Output format: <tag>|<true when a release run is already active>
set -euo pipefail

repository="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
release_workflow="${RELEASE_WORKFLOW:-Release}"
requested_tag="${RELEASE_TAG_OVERRIDE:-}"
latest_release_tag="$(gh release list \
  --repo "${repository}" \
  --limit 100 \
  --json tagName \
  --jq '.[].tagName' | grep '^v' | sort -V | tail -n 1 || true)"

is_newer_version() {
  local candidate="$1"

  [[ -n "${latest_release_tag}" ]] || return 0
  [[ "${candidate}" != "${latest_release_tag}" ]] || return 1
  [[ "$(printf '%s\n' "${latest_release_tag}" "${candidate}" | sort -V | tail -n 1)" == "${candidate}" ]]
}

if [[ -n "${requested_tag}" ]]; then
  git rev-parse --verify "refs/tags/${requested_tag}" >/dev/null
  if gh release view "${requested_tag}" --repo "${repository}" >/dev/null 2>&1; then
    printf '|false\n'
    exit 0
  fi

  active_runs="$(gh run list \
    --workflow "${release_workflow}" \
    --repo "${repository}" \
    --branch "${requested_tag}" \
    --json status \
    --jq '[.[] | select(.status != "completed")] | length')"
  if [[ "${active_runs}" -gt 0 ]]; then
    printf '%s|true\n' "${requested_tag}"
  else
    printf '%s|false\n' "${requested_tag}"
  fi
  exit 0
fi

while IFS= read -r tag; do
  [[ -n "${tag}" ]] || continue
  is_newer_version "${tag}" || continue

  if gh release view "${tag}" --repo "${repository}" >/dev/null 2>&1; then
    continue
  fi

  active_runs="$(gh run list \
    --workflow "${release_workflow}" \
    --repo "${repository}" \
    --branch "${tag}" \
    --json status \
    --jq '[.[] | select(.status != "completed")] | length')"

  if [[ "${active_runs}" -gt 0 ]]; then
    printf '%s|true\n' "${tag}"
  else
    printf '%s|false\n' "${tag}"
  fi
  exit 0
done < <(git tag --list 'v*' --sort=v:refname)

printf '|false\n'
