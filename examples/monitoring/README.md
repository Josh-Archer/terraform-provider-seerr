# Observability & Disaster Recovery Reference Architecture

This directory provides turnkey monitoring, Prometheus metrics exposition, Grafana visualization, automated state backup, and configuration drift detection for `terraform-provider-seerr`.

---

## 📁 Directory Contents

| Path | Purpose |
|:---|:---|
| `main.tf`, `variables.tf`, `outputs.tf` | OpenTofu / Terraform configuration querying server health, users, requests, jobs, and backup status. |
| `grafana-dashboard.json` | Pre-built, import-ready Grafana dashboard (v10+ compatible) visualizing health, requests, catalog growth, and DR status. |
| `exporter/` | Standalone Prometheus exporter serving `/metrics` on port `9850`. |
| `scripts/backup-state.sh` | Automated state snapshot script with SHA256 verification, retention rotation, and S3 offsite sync. |
| `scripts/drift-check.sh` | Automated drift detection script using `tofu plan -detailed-exitcode` with Discord/Slack webhooks. |

---

## 🚀 Quick Start

### 1. Run the Monitoring Terraform Configuration

```bash
cd examples/monitoring
export TF_VAR_seerr_url="http://localhost:5055"
export TF_VAR_seerr_api_key="your-api-key"

tofu init
tofu apply
tofu output -json metrics_summary
```

### 2. Launch the Prometheus Exporter

Using Docker Compose:

```bash
cd examples/monitoring/exporter
export SEERR_API_KEY="your-api-key"
docker compose up -d
```

Or running standalone:

```bash
export SEERR_URL="http://localhost:5055"
export SEERR_API_KEY="your-api-key"
python3 examples/monitoring/exporter/exporter.py
```

Scrape metrics at `http://localhost:9850/metrics`.

### 3. Import the Grafana Dashboard

1. In Grafana, navigate to **Dashboards** ➔ **New** ➔ **Import**.
2. Upload or paste the contents of `grafana-dashboard.json`.
3. Select your Prometheus datasource and click **Import**.

### 4. Configure Automated Drift Detection

Add a cron job to check for configuration drift every hour:

```cron
0 * * * * cd /path/to/your/tf-seerr && DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..." ./examples/monitoring/scripts/drift-check.sh
```

### 5. Automated State Backup Hook

Run state backup prior to any scheduled workflow or CI/CD apply:

```bash
./examples/monitoring/scripts/backup-state.sh
```
