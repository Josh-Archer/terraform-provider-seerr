#!/usr/bin/env bash
# Copyright (c) Josh Archer
# SPDX-License-Identifier: MPL-2.0
#
# State Backup Automation Script for OpenTofu / Terraform
# Captures cryptographic state snapshots before/after apply operations
# with retention pruning and optional S3/MinIO offsite replication.

set -euo pipefail

BACKUP_DIR="${SEERR_STATE_BACKUP_DIR:-./backups/state}"
RETENTION_DAYS="${SEERR_STATE_RETENTION_DAYS:-30}"
S3_BUCKET="${SEERR_STATE_S3_BUCKET:-}"
TIMESTAMP="$(date -u +"%Y%m%d_%H%M%SZ")"
STATE_FILE="${BACKUP_DIR}/terraform_${TIMESTAMP}.tfstate"
CHECKSUM_FILE="${STATE_FILE}.sha256"

mkdir -p "${BACKUP_DIR}"

echo "==> [State Backup] Capturing state snapshot..."

# Determine whether tofu or terraform is available
IAC_CLI=""
if command -v tofu &>/dev/null; then
    IAC_CLI="tofu"
elif command -v terraform &>/dev/null; then
    IAC_CLI="terraform"
fi

if [[ -n "${IAC_CLI}" ]]; then
    ${IAC_CLI} state pull > "${STATE_FILE}"
elif [[ -f "terraform.tfstate" ]]; then
    cp "terraform.tfstate" "${STATE_FILE}"
else
    echo "ERROR: Neither 'tofu', 'terraform', nor local 'terraform.tfstate' found." >&2
    exit 1
fi
# Verify state file is valid JSON and non-empty
if ! jq -e . "${STATE_FILE}" >/dev/null 2>&1; then
    echo "ERROR: Captured state snapshot is not valid JSON." >&2
    rm -f "${STATE_FILE}"
    exit 1
fi

# Compute cryptographic checksum using base filename
(cd "${BACKUP_DIR}" && sha256sum "$(basename "${STATE_FILE}")" > "$(basename "${CHECKSUM_FILE}")")
echo "==> [State Backup] Created snapshot: ${STATE_FILE}"
echo "==> [State Backup] SHA256: $(cat "${CHECKSUM_FILE}")"

# Export metrics timestamp for Prometheus exporter
echo "$(date +%s)" > "${BACKUP_DIR}/.last_backup_timestamp"

# Offsite sync if S3 bucket is configured
if [[ -n "${S3_BUCKET}" ]]; then
    echo "==> [State Backup] Syncing snapshot to ${S3_BUCKET}..."
    if command -v aws &>/dev/null; then
        aws s3 cp "${STATE_FILE}" "${S3_BUCKET}/"
        aws s3 cp "${CHECKSUM_FILE}" "${S3_BUCKET}/"
    elif command -v rclone &>/dev/null; then
        rclone copy "${STATE_FILE}" "${S3_BUCKET}/"
        rclone copy "${CHECKSUM_FILE}" "${S3_BUCKET}/"
    else
        echo "WARN: S3_BUCKET specified but neither 'aws' nor 'rclone' found in PATH." >&2
    fi
fi

# Prune snapshots older than retention window
echo "==> [State Backup] Pruning snapshots older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -type f -name "terraform_*.tfstate*" -mtime "+${RETENTION_DAYS}" -delete

echo "==> [State Backup] Backup completed successfully."
