#!/usr/bin/env bash
# Install a pinned GitHub CLI build and prepend it to PATH.
# uecb-shared images do not reliably include gh (#387, #392).
# Pin is the immutable cli/cli v2.103.3 release. Update the version and both
# checksums together from gh_${version}_checksums.txt on that release.
set -euo pipefail

version="2.103.3"
sha_amd64="4c806978ffce1cec70853302a56d70ac6da4f19bae7adca4e1a8ab00633dd91d"
sha_arm64="ee32df2323c8847f5fac6cb6b4b69ca75828e9329a8f22f677021f7b3306ce63"

case "$(uname -m)" in
  x86_64|amd64)
    arch="amd64"
    want="${sha_amd64}"
    ;;
  aarch64|arm64)
    arch="arm64"
    want="${sha_arm64}"
    ;;
  *)
    echo "unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

if [[ ! "${want}" =~ ^[0-9a-f]{64}$ ]]; then
  echo "missing checksum for ${arch}" >&2
  exit 1
fi

tool_cache="${RUNNER_TOOL_CACHE:-${HOME:?HOME is unset}/.cache/runner-tools}"
bin_dir="${tool_cache}/gh/${version}/${arch}/bin"
binary="${bin_dir}/gh"

if ! "${binary}" version >/dev/null 2>&1; then
  tmp="$(mktemp -d)"
  trap 'rm -rf "${tmp}"' EXIT
  tarball="gh_${version}_linux_${arch}.tar.gz"
  url="https://github.com/cli/cli/releases/download/v${version}/${tarball}"
  curl -fsSL --retry 3 --retry-delay 2 -o "${tmp}/${tarball}" "${url}"
  echo "${want}  ${tmp}/${tarball}" | sha256sum -c -
  tar -xzf "${tmp}/${tarball}" -C "${tmp}"
  mkdir -p "${bin_dir}"
  cp "${tmp}/gh_${version}_linux_${arch}/bin/gh" "${binary}"
  chmod 0755 "${binary}"
fi

if [[ -n "${GITHUB_PATH:-}" ]]; then
  echo "${bin_dir}" >> "${GITHUB_PATH}"
fi
export PATH="${bin_dir}:${PATH}"
gh version
