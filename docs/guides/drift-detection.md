# Automated Drift Detection Guide

This guide explains how to detect, alert on, and remediate **configuration drift** between your declared OpenTofu / Terraform code and live Seerr, Jellyseerr, or Overseerr instances.

---

## ⚡ What is Configuration Drift?

Configuration drift occurs when settings on a live Seerr instance are modified outside of Infrastructure as Code—for example, when an administrator adjusts Sonarr/Radarr quality profile mappings in the Seerr web UI, updates notification tokens, or changes user quota rules without updating the corresponding HCL files.

Regular drift detection guarantees:
- **Reproducibility**: Your code remains the single source of truth for disaster recovery.
- **Auditability**: Unauthorized or accidental UI modifications are surfaced immediately.
- **State Integrity**: Provider schema differences or API side-effects are identified before deployment failures occur.

---

## 🔍 How Drift Detection Works

OpenTofu and Terraform support the `-detailed-exitcode` flag on `plan` commands:

| Exit Code | Meaning | Action |
|:---|:---|:---|
| `0` | Succeeded, diff is empty (0 changes) | No action; system is fully converged. |
| `1` | Error encountered during execution | Alert on infrastructure/API failure. |
| `2` | Succeeded, diff is non-empty (**Drift detected**) | Alert team; review and reconcile changes. |

```bash
tofu plan -detailed-exitcode -no-color
```

---

## 🛠️ Automated Drift Check Script

A production-ready script is provided at [`examples/monitoring/scripts/drift-check.sh`](https://github.com/Josh-Archer/terraform-provider-seerr/blob/main/examples/monitoring/scripts/drift-check.sh).

### Supported Environment Variables:
- `DISCORD_WEBHOOK_URL`: Sends formatted embed alerts with plan diff summaries to a Discord channel.
- `SLACK_WEBHOOK_URL`: Sends formatted markdown notifications to a Slack channel.
- `GENERIC_WEBHOOK_URL`: Sends structured JSON payloads to custom webhooks or incident management systems.
- `SEERR_DRIFT_METRICS_FILE`: Exports drift metrics for Prometheus scraping.

### Example Run:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
./examples/monitoring/scripts/drift-check.sh
```

---

## ⏰ Scheduling Methods

### Option 1: Scheduled Cron Job (Self-Hosted / Homelab)

Add a cron entry under the user running Terraform:

```cron
# Run drift detection every 2 hours on the hour
0 */2 * * * cd /opt/homelab/terraform-seerr && DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..." /opt/homelab/terraform-seerr/examples/monitoring/scripts/drift-check.sh >> /var/log/seerr-drift.log 2>&1
```

### Option 2: Systemd Timer (Linux)

Create `/etc/systemd/system/seerr-drift.service`:

```ini
[Unit]
Description=Seerr OpenTofu Drift Detection
After=network-online.target

[Service]
Type=oneshot
User=homelab
WorkingDirectory=/opt/homelab/terraform-seerr
Environment=SEERR_URL=http://localhost:5055
EnvironmentFile=/etc/seerr/secrets.env
ExecStart=/opt/homelab/terraform-seerr/examples/monitoring/scripts/drift-check.sh
```

Create `/etc/systemd/system/seerr-drift.timer`:

```ini
[Unit]
Description=Run Seerr Drift Detection every 2 hours

[Timer]
OnCalendar=*-*-* 00/2:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

Enable and start the timer:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now seerr-drift.timer
```

### Option 3: GitHub Actions Scheduled Workflow

Create `.github/workflows/drift-detection.yml`:

```yaml
name: "Scheduled Drift Detection"

on:
  schedule:
    - cron: "0 6 * * *" # Daily at 06:00 UTC
  workflow_dispatch:

jobs:
  drift-check:
    name: "Check Seerr Configuration Drift"
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Setup OpenTofu
        uses: opentofu/setup-opentofu@v1
        with:
          tofu_version: 1.8.0

      - name: OpenTofu Init
        run: tofu init
        env:
          TF_VAR_seerr_url: ${{ secrets.SEERR_URL }}
          TF_VAR_seerr_api_key: ${{ secrets.SEERR_API_KEY }}

      - name: Execute Drift Detection
        run: ./examples/monitoring/scripts/drift-check.sh
        env:
          TF_VAR_seerr_url: ${{ secrets.SEERR_URL }}
          TF_VAR_seerr_api_key: ${{ secrets.SEERR_API_KEY }}
          DISCORD_WEBHOOK_URL: ${{ secrets.DISCORD_WEBHOOK_URL }}
```

---

## 🔄 Remediation Strategies

When drift is detected, you have two choices depending on intent:

### 1. Enforce Code (Overwrite Manual UI Changes)
If the UI changes were accidental or unauthorized, re-apply your configuration:

```bash
tofu apply
```
The provider will update the live Seerr instance to match your Git repository.

### 2. Adopt Changes (Update Code to Match UI)
If changes made in the UI should be kept permanently:
1. Inspect the diff output from `tofu plan`.
2. Update the corresponding HCL resource definitions in your `.tf` files.
3. Run `tofu plan -detailed-exitcode` to verify that the diff returns `0`.
4. Commit and push the updated HCL to Git.
