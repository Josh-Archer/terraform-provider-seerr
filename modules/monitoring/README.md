# Monitoring module

Reads the Seerr `about` endpoint once and exposes structured metrics for Grafana plus a Prometheus text-exposition payload containing server build, request, and media counts.

```hcl
module "monitoring" {
  source = "github.com/Josh-Archer/terraform-provider-seerr//modules/monitoring?ref=v0.41.0"
}

output "prometheus_text" {
  value = module.monitoring.prometheus_metrics
}
```

When calling the module from a checkout of this repository, use `source = "./modules/monitoring"` instead.

Write `prometheus_metrics` to a node-exporter textfile directory or serve it from the exporter in `examples/monitoring/exporter`.
