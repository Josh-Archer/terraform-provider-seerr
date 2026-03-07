terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

resource "seerr_radarr_server" "this" {
  name                 = var.name
  hostname             = var.hostname
  port                 = var.port
  api_key              = var.api_key
  use_ssl              = var.use_ssl
  base_url             = var.base_url
  quality_profile_id   = var.quality_profile_id
  active_directory     = var.active_directory
  is_4k                = var.is_4k
  minimum_availability = var.minimum_availability
  tags                 = var.tags
  is_default           = var.is_default
  enable_scan          = var.enable_scan
  sync_enabled         = var.sync_enabled
  prevent_search       = var.prevent_search
  tag_requests_with_user = var.tag_requests_with_user
  extra_payload_json   = jsonencode(var.extra_payload)
  url                  = var.url
}

output "resource_id" {
  value = seerr_radarr_server.this.id
}

output "server_id" {
  value = seerr_radarr_server.this.server_id
}
