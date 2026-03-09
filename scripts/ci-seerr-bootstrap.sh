#!/usr/bin/env bash
set -euo pipefail

namespace="${1:?namespace is required}"
port="${SEERR_TEST_PORT:-5055}"
service_name="${SEERR_TEST_SERVICE_NAME:-seerr}"
pvc_name="${SEERR_TEST_PVC_NAME:-seerr-config}"
seed_pod_name="${SEERR_TEST_SEED_POD_NAME:-seerr-bootstrap}"
local_port="${SEERR_TEST_LOCAL_PORT:-15055}"
use_port_forward="${SEERR_TEST_USE_PORT_FORWARD:-true}"
service_url="http://${service_name}.${namespace}.svc.cluster.local:${port}"
api_key="${SEERR_BOOTSTRAP_API_KEY:-seerr-ci-api-key}"
admin_email="${SEERR_TEST_ADMIN_EMAIL:-ci-admin@example.invalid}"
admin_username="${SEERR_TEST_ADMIN_USERNAME:-ci-admin}"
admin_avatar="${SEERR_TEST_ADMIN_AVATAR:-/logo_full.svg}"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required" >&2
  exit 1
fi

if [[ "${use_port_forward}" == "true" ]] && ! command -v curl >/dev/null 2>&1; then
  echo "curl is required when SEERR_TEST_USE_PORT_FORWARD=true" >&2
  exit 1
fi

service_url="http://${service_name}.${namespace}.svc.cluster.local:${port}"

for _ in $(seq 1 60); do
  if kubectl exec -n "${namespace}" deploy/seerr -- node -e "fetch('http://127.0.0.1:${port}/api/v1/status').then((r) => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if ! kubectl exec -n "${namespace}" deploy/seerr -- node -e "fetch('http://127.0.0.1:${port}/api/v1/status').then((r) => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))" >/dev/null 2>&1; then
  echo "Seerr did not become reachable in namespace ${namespace}" >&2
  exit 1
fi

export SEERR_NAMESPACE="${namespace}"
export SEERR_URL="${service_url}"

cat <<EOF | kubectl apply -n "${namespace}" -f -
apiVersion: v1
kind: Pod
metadata:
  name: ${seed_pod_name}
spec:
  restartPolicy: Never
  containers:
    - name: bootstrap
      image: python:3.12-alpine
      command:
        - sh
        - -c
        - sleep 3600
      volumeMounts:
        - name: config
          mountPath: /config
  volumes:
    - name: config
      persistentVolumeClaim:
        claimName: ${pvc_name}
EOF

kubectl wait -n "${namespace}" --for=condition=Ready "pod/${seed_pod_name}" --timeout=180s >/dev/null

cat <<EOF | kubectl exec -i -n "${namespace}" "${seed_pod_name}" -- python3
import json
import sqlite3
from pathlib import Path

settings_path = Path("/config/settings.json")
settings = json.loads(settings_path.read_text())
settings.setdefault("main", {})["apiKey"] = "${api_key}"
settings.setdefault("public", {})["initialized"] = True
settings_path.write_text(json.dumps(settings, indent=1) + "\n")

db = sqlite3.connect("/config/db/db.sqlite3")
user_result = db.execute(
    "UPDATE user SET email = ?, username = ?, permissions = 2, avatar = ?, userType = 1 WHERE id = 1",
    ("${admin_email}", "${admin_username}", "${admin_avatar}"),
)
if user_result.rowcount == 0:
    db.execute(
        "INSERT INTO user (id, email, username, permissions, avatar, userType) VALUES (1, ?, ?, 2, ?, 1)",
        ("${admin_email}", "${admin_username}", "${admin_avatar}"),
    )

settings_result = db.execute(
    "UPDATE user_settings SET locale = 'en', watchlistSyncMovies = 0, watchlistSyncTv = 0 WHERE userId = 1"
)
if settings_result.rowcount == 0:
    db.execute(
        "INSERT INTO user_settings (userId, locale, watchlistSyncMovies, watchlistSyncTv) VALUES (1, 'en', 0, 0)"
    )
db.commit()
db.close()
EOF

kubectl rollout restart deployment/seerr -n "${namespace}" >/dev/null
kubectl rollout status deployment/seerr -n "${namespace}" --timeout=180s >/dev/null

if ! kubectl exec -n "${namespace}" deploy/seerr -- node -e "fetch('http://127.0.0.1:${port}/api/v1/settings/main', { headers: { 'X-Api-Key': '${api_key}' } }).then((r) => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))" >/dev/null 2>&1; then
  echo "Bootstrapped API key did not authenticate against ${service_url}" >&2
  exit 1
fi

output_url="${service_url}"
if [[ "${use_port_forward}" == "true" ]]; then
  log_file="${TMPDIR:-/tmp}/seerr-port-forward-${namespace}.log"
  kubectl port-forward -n "${namespace}" "service/${service_name}" "${local_port}:${port}" >"${log_file}" 2>&1 &
  port_forward_pid=$!

  for _ in $(seq 1 30); do
    if curl -fsS "http://127.0.0.1:${local_port}/api/v1/status" >/dev/null 2>&1; then
      output_url="http://127.0.0.1:${local_port}"
      break
    fi
    sleep 1
  done

  if [[ "${output_url}" != "http://127.0.0.1:${local_port}" ]]; then
    echo "Port-forward to ${namespace}/${service_name} did not become ready" >&2
    exit 1
  fi
fi

api_key="${api_key//$'\r'/}"
api_key="${api_key//$'\n'/}"

printf 'SEERR_URL=%s\n' "${output_url}"
printf 'SEERR_API_KEY=%s\n' "${api_key}"
if [[ -n "${port_forward_pid:-}" ]]; then
  printf 'SEERR_PORT_FORWARD_PID=%s\n' "${port_forward_pid}"
fi
