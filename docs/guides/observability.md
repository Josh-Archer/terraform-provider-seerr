# Observability & Metrics Guide

This guide describes how to monitor your Seerr / Jellyseerr / Overseerr instances and OpenTofu / Terraform state using native provider telemetry data sources, Prometheus metrics exposition, and Grafana dashboards.

---

## ⚡ Overview

Monitoring media request infrastructure involves two complementary pillars:
1. **Application Health & Capacity**: Tracking server version, upstream update availability, background job statuses, pending media requests, indexed media items, and user activity.
2. **Infrastructure-as-Code State & Drift**: Verifying that the running instance matches your declared Terraform/OpenTofu configuration and tracking state backup freshness.

---

## 📊 Telemetry Data Sources

The `seerr` provider includes dedicated data sources for querying system state:

| Data Source | Key Telemetry Attributes | Use Case |
|:---|:---|:---|
| `seerr_about` | `version`, `total_requests`, `total_media_items`, `tz`, `app_data_path` | Server metadata, total catalog volume, request throughput |
| `seerr_service_status` | `version`, `commit_tag`, `update_available`, `commits_behind`, `restart_required` | Upstream release tracking, update alerts, restart reminders |
| `seerr_requests` | `filter_status` (1=Pending, 2=Approved, 3=Declined), `total` | Request backlog and fulfillment tracking |
| `seerr_users` | `users` list, user types (Plex, Local, Jellyfin, Emby) | User base and active account counts |
| `seerr_jobs` | `jobs` list (`id`, `name`, `type`, `running`, `next_run`) | Background scheduler execution health |
| `seerr_backup_settings` | `storage_path`, `schedule`, `retention` | Automated database backup verification |

### Terraform Telemetry Configuration Example

```hcl
# main.tf
data "seerr_about" "server" {}
data "seerr_service_status" "status" {}
data "seerr_requests" "pending" {
  filter_status = 1 # Pending approval
}
data "seerr_users" "all" {}
data "seerr_jobs" "scheduler" {}
data "seerr_backup_settings" "backup" {}

output "metrics_summary" {
  value = {
    version           = data.seerr_about.server.version
    update_available  = data.seerr_service_status.status.update_available
    commits_behind    = data.seerr_service_status.status.commits_behind
    restart_required  = data.seerr_service_status.status.restart_required
    total_requests    = data.seerr_about.server.total_requests
    total_media_items = data.seerr_about.server.total_media_items
    total_users       = length(data.seerr_users.all.users)
    pending_requests  = data.seerr_requests.pending.total
    backup_retention  = data.seerr_backup_settings.backup.retention
  }
}
```

---

## 🚀 Prometheus Exporter Setup

The repository includes a standalone, zero-dependency Prometheus exporter in `examples/monitoring/exporter/exporter.py`.

### Running with Docker

```bash
docker run -d \
  --name seerr-prometheus-exporter \
  --restart unless-stopped \
  -p 9850:9850 \
  -e SEERR_URL="http://192.168.1.100:5055" \
  -e SEERR_API_KEY="YOUR_SEERR_API_KEY" \
  ghcr.io/josh-archer/seerr-prometheus-exporter:latest
```

### Running with Docker Compose

```yaml
version: "3.8"
services:
  seerr-exporter:
    image: python:3.11-alpine
    container_name: seerr-exporter
    restart: unless-stopped
    volumes:
      - ./examples/monitoring/exporter/exporter.py:/app/exporter.py:ro
    environment:
      - EXPORTER_PORT=9850
      - SEERR_URL=http://seerr:5055
      - SEERR_API_KEY=${SEERR_API_KEY}
    ports:
      - "9850:9850"
    command: ["python", "-u", "/app/exporter.py"]
```

### Prometheus Scrape Configuration

Add the scrape target to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'seerr'
    static_configs:
      - targets: ['seerr-exporter:9850']
    scrape_interval: 30s
    scrape_timeout: 10s
```

---

## 📈 Pre-Built Grafana Dashboard

A production-ready Grafana dashboard is provided in [`examples/monitoring/grafana-dashboard.json`](https://github.com/Josh-Archer/terraform-provider-seerr/blob/main/examples/monitoring/grafana-dashboard.json).

### Features Included:
- **Server Release & Health**: Version badge, update availability indicator, commit lag, restart alerts.
- **Request Backlog Stat**: Color-coded pending request indicator (>0 turns amber, >10 turns red).
- **Catalog & Request Time Series**: Live tracking of cumulative requests and media catalog volume.
- **Background Scheduler Table**: Execution status of automated scans, library syncs, and cache cleanups.
- **IaC Drift & State Freshness**: Real-time status of Terraform/OpenTofu state convergence.

---

## 🔔 Prometheus Alerting Rules

Add these standard alerting rules to your Prometheus alert manager configuration:

```yaml
groups:
  - name: seerr_alerts
    rules:
      - alert: SeerrUpdateAvailable
        expr: seerr_update_available == 1
        for: 24h
        labels:
          severity: info
        annotations:
          summary: "New Seerr upstream release available"
          description: "Seerr instance is {{ $value }} commits behind upstream."

      - alert: SeerrRestartRequired
        expr: seerr_restart_required == 1
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Seerr restart required"
          description: "Seerr requires a restart to apply configuration or database updates."

      - alert: SeerrPendingRequestsHigh
        expr: seerr_requests_pending_total > 15
        for: 2h
        labels:
          severity: warning
        annotations:
          summary: "High volume of pending media requests"
          description: "There are currently {{ $value }} media requests awaiting approval."

      - alert: SeerrConfigurationDriftDetected
        expr: seerr_terraform_drift_status == 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Seerr configuration drift detected"
          description: "Live Seerr instance has diverged from declared OpenTofu / Terraform code."
```
