terraform {
  required_version = ">= 1.5.0"

  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

data "seerr_about" "this" {}

locals {
  metrics = {
    server_version    = data.seerr_about.this.version
    total_requests    = data.seerr_about.this.total_requests
    total_media_items = data.seerr_about.this.total_media_items
    timezone          = data.seerr_about.this.tz
  }

  prometheus_metrics = join("\n", [
    "# HELP seerr_requests_total Total requests known to Seerr.",
    "# TYPE seerr_requests_total gauge",
    "seerr_requests_total ${data.seerr_about.this.total_requests}",
    "# HELP seerr_media_items_total Total media items known to Seerr.",
    "# TYPE seerr_media_items_total gauge",
    "seerr_media_items_total ${data.seerr_about.this.total_media_items}",
    "# HELP seerr_build_info Seerr build information.",
    "# TYPE seerr_build_info gauge",
    "seerr_build_info{version=\"${replace(data.seerr_about.this.version, "\"", "\\\"")}\"} 1",
    "",
  ])
}

