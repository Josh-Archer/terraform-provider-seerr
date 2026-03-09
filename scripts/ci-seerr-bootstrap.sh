#!/usr/bin/env bash
set -euo pipefail

namespace="${1:?namespace is required}"
port="${SEERR_TEST_PORT:-5055}"
service_name="${SEERR_TEST_SERVICE_NAME:-seerr}"
service_url="http://${service_name}.${namespace}.svc.cluster.local:${port}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

for _ in $(seq 1 60); do
  if curl -fsS "${service_url}/api/v1/status" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! curl -fsS "${service_url}/api/v1/status" >/dev/null 2>&1; then
  echo "Seerr did not become reachable at ${service_url}" >&2
  exit 1
fi

export SEERR_NAMESPACE="${namespace}"
export SEERR_URL="${service_url}"

api_key="${SEERR_BOOTSTRAP_API_KEY:-}"

if [[ -z "${api_key}" && -n "${SEERR_BOOTSTRAP_COMMAND:-}" ]]; then
  bash -lc "${SEERR_BOOTSTRAP_COMMAND}"
fi

if [[ -z "${api_key}" && -n "${SEERR_BOOTSTRAP_API_KEY_COMMAND:-}" ]]; then
  api_key="$(bash -lc "${SEERR_BOOTSTRAP_API_KEY_COMMAND}")"
fi

if [[ -z "${api_key}" ]]; then
  cat >&2 <<'EOF'
No API key bootstrap method configured.
Set one of:
  - SEERR_BOOTSTRAP_API_KEY
  - SEERR_BOOTSTRAP_COMMAND and SEERR_BOOTSTRAP_API_KEY_COMMAND
EOF
  exit 1
fi

api_key="${api_key//$'\r'/}"
api_key="${api_key//$'\n'/}"

printf 'SEERR_URL=%s\n' "${service_url}"
printf 'SEERR_API_KEY=%s\n' "${api_key}"
