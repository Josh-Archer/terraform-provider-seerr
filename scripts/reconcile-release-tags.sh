#!/usr/bin/env bash
# Print the oldest stable tag at or after the configured floor that does not
# have a published GitHub Release.
# Output format: <tag>|<true when a release run is already active>
set -euo pipefail

repository="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
release_workflow="${RELEASE_WORKFLOW:-release.yml}"
requested_tag="${RELEASE_TAG_OVERRIDE:-}"
reconcile_from_tag="${RELEASE_RECONCILE_FROM_TAG:-}"
semver_pattern='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$'
stable_pattern='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'

validate_tag() {
  local tag="$1"

  if [[ ! "${tag}" =~ ${semver_pattern} ]]; then
    echo "Invalid release tag: ${tag}" >&2
    return 1
  fi
  if ! git show-ref --verify --quiet "refs/tags/${tag}"; then
    echo "Release tag does not exist: ${tag}" >&2
    return 1
  fi
}

# Return 0 for a published release, 1 when the tag has no release or only a
# draft, and 2 for API/auth/network failures. GoReleaser can resume drafts.
has_published_release() {
  local tag="$1"
  local response

  if response="$(gh api "repos/${repository}/releases/tags/${tag}" --jq '.draft' 2>&1)"; then
    case "${response}" in
      false)
        return 0
        ;;
      true)
        return 1
        ;;
      *)
        echo "Unexpected GitHub Release response for ${tag}: ${response}" >&2
        return 2
        ;;
    esac
  fi

  if [[ "${response}" == *"(HTTP 404)"* ]]; then
    return 1
  fi

  echo "Unable to inspect GitHub Release for ${tag}: ${response}" >&2
  return 2
}

has_active_release_run() {
  local tag="$1"
  local active_runs

  if ! active_runs="$(gh run list \
    --workflow "${release_workflow}" \
    --repo "${repository}" \
    --limit 100 \
    --json status,displayTitle \
    --jq "[.[] | select(.displayTitle == \"Release ${tag}\" and .status != \"completed\")] | length")"; then
    echo "Unable to inspect active Release runs for ${tag}" >&2
    return 2
  fi
  if [[ ! "${active_runs}" =~ ^[0-9]+$ ]]; then
    echo "Unexpected active Release run count for ${tag}: ${active_runs}" >&2
    return 2
  fi
  [[ "${active_runs}" -gt 0 ]]
}

print_candidate() {
  local tag="$1"

  if has_active_release_run "${tag}"; then
    printf '%s|true\n' "${tag}"
  else
    run_status=$?
    if [[ "${run_status}" -eq 1 ]]; then
      printf '%s|false\n' "${tag}"
    else
      return "${run_status}"
    fi
  fi
}

if [[ -n "${requested_tag}" ]]; then
  validate_tag "${requested_tag}"
  if has_published_release "${requested_tag}"; then
    printf '|false\n'
    exit 0
  else
    release_status=$?
    [[ "${release_status}" -eq 1 ]] || exit "${release_status}"
  fi
  print_candidate "${requested_tag}"
  exit 0
fi

if [[ -z "${reconcile_from_tag}" ]]; then
  echo "RELEASE_RECONCILE_FROM_TAG must identify the oldest automatically managed stable tag" >&2
  exit 1
fi
if [[ ! "${reconcile_from_tag}" =~ ${stable_pattern} ]]; then
  echo "Automatic reconciliation floor must be a stable version tag: ${reconcile_from_tag}" >&2
  exit 1
fi
validate_tag "${reconcile_from_tag}"

reached_floor=false
while IFS= read -r tag; do
  [[ -n "${tag}" ]] || continue
  if [[ "${tag}" == "${reconcile_from_tag}" ]]; then
    reached_floor=true
  fi
  [[ "${reached_floor}" == "true" ]] || continue
  [[ "${tag}" =~ ${stable_pattern} ]] || continue

  if has_published_release "${tag}"; then
    continue
  else
    release_status=$?
    [[ "${release_status}" -eq 1 ]] || exit "${release_status}"
  fi

  print_candidate "${tag}"
  exit 0
done < <(git tag --list 'v*' --sort=v:refname)

printf '|false\n'
