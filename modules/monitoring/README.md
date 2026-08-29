# Monitoring module

Reads the Seerr `about` endpoint once and exposes structured metrics for Grafana plus a Prometheus text-exposition payload containing server build, request, and media counts.

```hcl
module "monitoring" {
  source = "./modules/monitoring"
}

output "prometheus_text" {
  value = module.monitoring.prometheus_metrics
}
```

Write `prometheus_metrics` to a node-exporter textfile directory or serve it from the exporter in `examples/monitoring/exporter`.

