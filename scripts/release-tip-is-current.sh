#!/usr/bin/env bash
set -euo pipefail

expected_sha="${1:?expected commit SHA is required}"
default_branch="${2:?default branch is required}"
remote="${RELEASE_REMOTE:-origin}"
remote_ref="refs/remotes/${remote}/${default_branch}"

git fetch --no-tags "${remote}" \
  "+refs/heads/${default_branch}:${remote_ref}" >&2
default_tip="$(git rev-parse "${remote_ref}^{commit}")"
expected_commit="$(git rev-parse "${expected_sha}^{commit}")"

if [[ "${expected_commit}" == "${default_tip}" ]]; then
  printf 'true\n'
else
  printf 'false\n'
fi
