#!/usr/bin/env bash

set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
tmp_root="$repo_root/.git/tmp"
mkdir -p "$tmp_root"
tmp_dir="$(mktemp -d "$tmp_root/seerr-generate-check.XXXXXX")"

resolve_tool() {
  local tool="$1"

  if command -v "$tool" >/dev/null 2>&1; then
    command -v "$tool"
    return 0
  fi

  if command -v "${tool}.exe" >/dev/null 2>&1; then
    command -v "${tool}.exe"
    return 0
  fi

  if command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -Command "(Get-Command $tool).Source" 2>/dev/null | tr -d '\r'
    return 0
  fi

  return 1
}

go_bin="$(resolve_tool go)" || {
  echo "Unable to find 'go' on PATH." >&2
  exit 1
}

gofmt_bin="$(resolve_tool gofmt)" || {
  echo "Unable to find 'gofmt' on PATH." >&2
  exit 1
}

cleanup() {
  git -C "$repo_root" worktree remove --force "$tmp_dir" >/dev/null 2>&1 || rm -rf "$tmp_dir"
}

trap cleanup EXIT INT TERM

git -C "$repo_root" worktree add --detach "$tmp_dir" HEAD >/dev/null

cd "$tmp_dir"

echo "Running go generate ./..."
"$go_bin" generate ./...

echo "Running gofmt -w ."
"$gofmt_bin" -w .

if ! git diff --quiet -- .; then
  echo "Generated files or formatting are out of date." >&2
  echo "Run the following from the repo root and commit the changes:" >&2
  echo "  go generate ./..." >&2
  echo "  gofmt -w ." >&2
  echo >&2
  git --no-pager diff --stat -- . >&2
  exit 1
fi
