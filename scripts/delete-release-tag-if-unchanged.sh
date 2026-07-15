#!/usr/bin/env bash
set -euo pipefail

tag="${1:?release tag is required}"
expected_object="${2:?expected tag object is required}"
remote="${3:-origin}"

if [[ ! "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "Invalid release tag: ${tag}" >&2
  exit 2
fi

current_object="$(git ls-remote --refs "${remote}" "refs/tags/${tag}" | awk 'NR == 1 { print $1 }')"
if [[ -z "${current_object}" ]]; then
  echo "Refusing to delete ${tag}: the remote tag no longer exists." >&2
  exit 1
fi
if [[ "${current_object}" != "${expected_object}" ]]; then
  echo "Refusing to delete ${tag}: the remote tag object changed from ${expected_object} to ${current_object}." >&2
  exit 1
fi

git_options=()
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  auth_header="$(printf 'x-access-token:%s' "${GITHUB_TOKEN}" | base64 | tr -d '\r\n')"
  git_options+=(
    -c
    "http.https://github.com/.extraheader=AUTHORIZATION: basic ${auth_header}"
  )
fi

git "${git_options[@]}" push \
  --force-with-lease="refs/tags/${tag}:${expected_object}" \
  "${remote}" ":refs/tags/${tag}"
