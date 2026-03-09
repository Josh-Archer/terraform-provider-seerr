#!/usr/bin/env bash
set -euo pipefail

namespace="${1:?namespace is required}"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required" >&2
  exit 1
fi

kubectl delete namespace "${namespace}" --ignore-not-found=true --wait=false
