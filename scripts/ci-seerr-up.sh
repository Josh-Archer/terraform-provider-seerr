#!/usr/bin/env bash
set -euo pipefail

namespace="${1:?namespace is required}"
image="${SEERR_TEST_IMAGE:-ghcr.io/seerr-team/seerr:latest}"
config_path="${SEERR_TEST_CONFIG_PATH:-/app/config}"
port="${SEERR_TEST_PORT:-5055}"
log_level="${SEERR_TEST_LOG_LEVEL:-debug}"
pvc_name="${SEERR_TEST_PVC_NAME:-seerr-config}"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required" >&2
  exit 1
fi

kubectl get namespace "${namespace}" >/dev/null 2>&1 || kubectl create namespace "${namespace}"

cat <<EOF | kubectl apply -n "${namespace}" -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ${pvc_name}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: seerr
spec:
  replicas: 1
  selector:
    matchLabels:
      app: seerr
  template:
    metadata:
      labels:
        app: seerr
    spec:
      initContainers:
        - name: init-config
          image: busybox:1.36
          command:
            - sh
            - -c
            - mkdir -p "${config_path}" && chown -R 1000:1000 "${config_path}"
          volumeMounts:
            - name: config
              mountPath: "${config_path}"
      containers:
        - name: seerr
          image: "${image}"
          env:
            - name: LOG_LEVEL
              value: "${log_level}"
            - name: PORT
              value: "${port}"
          ports:
            - name: http
              containerPort: ${port}
          readinessProbe:
            httpGet:
              path: /api/v1/status
              port: http
            initialDelaySeconds: 10
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /api/v1/status
              port: http
            initialDelaySeconds: 20
            periodSeconds: 10
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
          volumeMounts:
            - name: config
              mountPath: "${config_path}"
      volumes:
        - name: config
          persistentVolumeClaim:
            claimName: ${pvc_name}
---
apiVersion: v1
kind: Service
metadata:
  name: seerr
spec:
  selector:
    app: seerr
  ports:
    - name: http
      port: ${port}
      targetPort: http
EOF

kubectl rollout status deployment/seerr -n "${namespace}" --timeout=180s
