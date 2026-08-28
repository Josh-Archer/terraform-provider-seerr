#!/usr/bin/env bash
# Copyright (c) Josh Archer
# SPDX-License-Identifier: MPL-2.0
#
# Automated Drift Detection Script for OpenTofu / Terraform
# Executes non-destructive plan with -detailed-exitcode, parses changes,
# exports drift metrics, and sends notifications on detected drift.

set -uo pipefail

DISCORD_WEBHOOK_URL="${DISCORD_WEBHOOK_URL:-}"
SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"
GENERIC_WEBHOOK_URL="${GENERIC_WEBHOOK_URL:-}"
METRICS_FILE="${SEERR_DRIFT_METRICS_FILE:-./backups/state/.drift_metrics}"

# Select binary
IAC_CLI=""
if command -v tofu &>/dev/null; then
    IAC_CLI="tofu"
elif command -v terraform &>/dev/null; then
    IAC_CLI="terraform"
else
    echo "ERROR: Neither 'tofu' nor 'terraform' found in PATH." >&2
    exit 1
fi

echo "==> [Drift Check] Running ${IAC_CLI} plan -detailed-exitcode..."

PLAN_LOG="$(mktemp /tmp/seerr_drift_plan_XXXXXX.log)"
trap 'rm -f "${PLAN_LOG}"' EXIT

# Run plan
set +e
${IAC_CLI} plan -detailed-exitcode -no-color > "${PLAN_LOG}" 2>&1
PLAN_EXIT_CODE=$?
set -e

mkdir -p "$(dirname "${METRICS_FILE}")"

if [[ ${PLAN_EXIT_CODE} -eq 0 ]]; then
    echo "==> [Drift Check] SUCCESS: No drift detected. Infrastructure matches code."
    echo "SEERR_TF_DRIFT_STATUS=0" > "${METRICS_FILE}"
    exit 0
elif [[ ${PLAN_EXIT_CODE} -eq 2 ]]; then
    echo "==> [Drift Check] WARNING: Drift detected! Infrastructure has diverged from state."
    echo "SEERR_TF_DRIFT_STATUS=1" > "${METRICS_FILE}"

    # Extract summary lines
    SUMMARY="$(grep -E "(Plan:|No changes\.|Changes to Outputs:)" "${PLAN_LOG}" || echo "Changes detected in live instance.")"
    DIFF_PREVIEW="$(tail -n 30 "${PLAN_LOG}")"

    # Send Discord notification
    if [[ -n "${DISCORD_WEBHOOK_URL}" ]]; then
        echo "==> [Drift Check] Sending alert to Discord..."
        JSON_PAYLOAD=$(jq -n \
            --arg title "⚠️ Seerr Configuration Drift Detected" \
            --arg summary "${SUMMARY}" \
            --arg diff "\`\`\`\n${DIFF_PREVIEW}\n\`\`\`" \
            '{
                embeds: [{
                    title: $title,
                    color: 16753920,
                    description: ($summary + "\n\n" + $diff),
                    footer: { text: "OpenTofu / Terraform Seerr Drift Monitor" }
                }]
            }')
        curl -s -X POST -H "Content-Type: application/json" -d "${JSON_PAYLOAD}" "${DISCORD_WEBHOOK_URL}" >/dev/null || true
    fi

    # Send Slack notification
    if [[ -n "${SLACK_WEBHOOK_URL}" ]]; then
        echo "==> [Drift Check] Sending alert to Slack..."
        JSON_PAYLOAD=$(jq -n \
            --arg text "⚠️ *Seerr Configuration Drift Detected*\n*Summary:* ${SUMMARY}\n\`\`\`\n${DIFF_PREVIEW}\n\`\`\`" \
            '{text: $text}')
        curl -s -X POST -H "Content-Type: application/json" -d "${JSON_PAYLOAD}" "${SLACK_WEBHOOK_URL}" >/dev/null || true
    fi

    # Send Generic webhook
    if [[ -n "${GENERIC_WEBHOOK_URL}" ]]; then
        echo "==> [Drift Check] Sending webhook payload..."
        JSON_PAYLOAD=$(jq -n \
            --arg event "drift_detected" \
            --arg summary "${SUMMARY}" \
            --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
            '{event: $event, summary: $summary, timestamp: $timestamp}')
        curl -s -X POST -H "Content-Type: application/json" -d "${JSON_PAYLOAD}" "${GENERIC_WEBHOOK_URL}" >/dev/null || true
    fi

    exit 2
else
    echo "==> [Drift Check] ERROR: ${IAC_CLI} plan failed with exit code ${PLAN_EXIT_CODE}."
    echo "SEERR_TF_DRIFT_STATUS=2" > "${METRICS_FILE}"
    cat "${PLAN_LOG}" >&2
    exit "${PLAN_EXIT_CODE}"
fi
