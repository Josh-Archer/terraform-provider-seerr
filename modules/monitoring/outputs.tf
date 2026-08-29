output "metrics" {
  description = "Structured values suitable for Grafana provisioning, dashboards, or external exporters."
  value       = local.metrics
}

output "prometheus_metrics" {
  description = "Prometheus text exposition payload for a textfile collector or HTTP exporter."
  value       = local.prometheus_metrics
}

output "grafana_annotations" {
  description = "Stable labels suitable for annotating Grafana dashboards."
  value = {
    service  = "seerr"
    version  = data.seerr_about.this.version
    timezone = data.seerr_about.this.tz
  }
}

