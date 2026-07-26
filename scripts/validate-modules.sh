#!/usr/bin/env bash
set -eo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export TF_CLI_CONFIG_FILE="${repo_root}/.tofurc"

echo "Building provider locally..."
mkdir -p "${repo_root}/.bin"
cd "${repo_root}"
"${GO_BIN:-go}" build -o "${repo_root}/.bin/terraform-provider-seerr" .
"${GO_BIN:-go}" build -o "${repo_root}/.bin/terraform-provider-seerr.exe" . || true

echo "Generating .tofurc..."
bin_dir="${repo_root}/.bin"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
  bin_dir=$(cygpath -m "$bin_dir")
fi

cat <<EOF > "${TF_CLI_CONFIG_FILE}"
provider_installation {
  dev_overrides {
    "registry.opentofu.org/josh-archer/seerr" = "${bin_dir}"
    "registry.terraform.io/josh-archer/seerr" = "${bin_dir}"
    "josh-archer/seerr" = "${bin_dir}"
  }
  direct {}
}
EOF

echo "Running regression test..."
cd "${repo_root}/scripts/testdata/invalid_module"
"${TOFU_BIN:-tofu}" init -backend=false || true
if "${TOFU_BIN:-tofu}" validate; then
  echo "Regression test failed: validate should have caught the invalid argument"
  exit 1
else
  echo "Regression test passed: invalid argument caught."
fi

echo "Validating modules..."
shopt -s nullglob
for d in "${repo_root}/modules"/* "${repo_root}/examples/modules"/*; do
  if [ -d "$d" ]; then
    echo "Validating $d..."
    cd "$d"
    # Provider resolution still requires init for nested modules, even with
    # development overrides. All modules must declare josh-archer/seerr so
    # OpenTofu does not invent a hashicorp/seerr dependency.
    "${TOFU_BIN:-tofu}" init -backend=false -input=false || true
    "${TOFU_BIN:-tofu}" validate
  fi
done
shopt -u nullglob

echo "Module validation complete."
