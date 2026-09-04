#!/usr/bin/env bash
# Fail when a release-please PR proposes a major bump with no new breaking
# commits after the latest stable tag. Historical feat! / BREAKING CHANGE
# footers already shipped in v1.0.0 must not mint v2/v3 from changelog replay.
set -euo pipefail

title="${PR_TITLE:-}"
if [[ ! "${title}" =~ release[[:space:]]+([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
  echo "Not a versioned release-please title; allowing."
  exit 0
fi

new_major="${BASH_REMATCH[1]}"
stable_pattern='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
latest_stable="$(git tag --list 'v*' --sort=-v:refname | grep -E "${stable_pattern}" | head -n 1 || true)"
if [[ -z "${latest_stable}" ]]; then
  echo "No stable tag; allowing first release."
  exit 0
fi

latest_major="${latest_stable#v}"
latest_major="${latest_major%%.*}"
if (( new_major <= latest_major )); then
  echo "Not a major bump versus ${latest_stable}; allowing."
  exit 0
fi

if ! git rev-parse --verify "${latest_stable}^{commit}" >/dev/null 2>&1; then
  echo "Missing tag object ${latest_stable}" >&2
  exit 1
fi

breaking=0
while IFS= read -r commit; do
  [[ -z "${commit}" ]] && continue
  subject="$(git log -1 --format=%s "${commit}")"
  if [[ "${subject}" =~ ^chore(\(main\))?:\ release[[:space:]] ]]; then
    continue
  fi
  if [[ "${subject}" =~ ^Merge\ pull\ request ]]; then
    continue
  fi
  body="$(git log -1 --format=%s%n%b "${commit}")"
  if echo "${body}" | grep -qE '(BREAKING CHANGE:|^[a-zA-Z]+(\([^)]+\))?!:)'; then
    breaking=1
    break
  fi
done < <(git rev-list "${latest_stable}..HEAD")

if (( breaking == 0 )); then
  echo "Refusing empty major v${new_major}.0.0: no feat!/BREAKING CHANGE commits after ${latest_stable}." >&2
  exit 1
fi

echo "Major bump has breaking commits after ${latest_stable}; allowing."
