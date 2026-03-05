terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

resource "seerr_sonarr_server" "this" {
  name                   = var.name
  hostname               = var.hostname
  port                   = var.port
  api_key                = var.api_key
  use_ssl                = var.use_ssl
  base_url               = var.base_url
  active_profile_id      = var.active_profile_id
  active_directory       = var.active_directory
  active_anime_directory = var.active_anime_directory
  tags                   = var.tags
  anime_tags             = var.anime_tags
  is_4k                  = var.is_4k
  is_default             = var.is_default
  enable_season_folders  = var.enable_season_folders
  sync_enabled           = var.sync_enabled
  prevent_search         = var.prevent_search
  tag_requests           = var.tag_requests
  extra_payload_json     = jsonencode(var.extra_payload)
  url                    = var.url
}

output "resource_id" {
  value = seerr_sonarr_server.this.id
}

output "server_id" {
  value = seerr_sonarr_server.this.server_id
}
