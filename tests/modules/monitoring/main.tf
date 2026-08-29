terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

module "monitoring" {
  source = "../../../modules/monitoring"
}

output "metrics" {
  value = module.monitoring.metrics
}

output "prometheus_metrics" {
  value = module.monitoring.prometheus_metrics
}

