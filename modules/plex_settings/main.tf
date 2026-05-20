terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

resource "seerr_plex_settings" "this" {
  ip      = var.ip
  port    = var.port
  use_ssl = var.use_ssl
}

output "resource_id" {
  value = seerr_plex_settings.this.id
}
