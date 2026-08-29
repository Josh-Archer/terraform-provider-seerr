terraform {
  required_version = ">= 1.5.0"

  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

locals {
  notification_webhook = var.notification_webhook == null ? null : {
    webhook_url  = var.notification_webhook.url
    auth_header  = var.notification_webhook.auth_header
    json_payload = var.notification_webhook.json_payload
  }
}

resource "seerr_radarr_server" "this" {
  for_each = var.radarr_servers

  name                   = each.value.name
  hostname               = each.value.hostname
  port                   = each.value.port
  api_key                = each.value.api_key
  use_ssl                = each.value.use_ssl
  base_url               = each.value.base_url
  quality_profile_id     = each.value.quality_profile_id
  active_directory       = each.value.root_folder
  minimum_availability   = each.value.minimum_availability
  tags                   = each.value.tags
  is_4k                  = each.value.is_4k
  is_default             = each.value.is_default
  enable_scan            = each.value.enable_scan
  sync_enabled           = each.value.sync_enabled
  prevent_search         = each.value.prevent_search
  tag_requests_with_user = each.value.tag_requests_with_user
}

resource "seerr_sonarr_server" "this" {
  for_each = var.sonarr_servers

  name                   = each.value.name
  hostname               = each.value.hostname
  port                   = each.value.port
  api_key                = each.value.api_key
  use_ssl                = each.value.use_ssl
  base_url               = each.value.base_url
  quality_profile_id     = each.value.quality_profile_id
  active_directory       = each.value.root_folder
  active_anime_directory = each.value.anime_root_folder
  tags                   = each.value.tags
  anime_tags             = each.value.anime_tags
  is_4k                  = each.value.is_4k
  is_default             = each.value.is_default
  enable_scan            = each.value.enable_scan
  enable_season_folders  = each.value.enable_season_folders
  sync_enabled           = each.value.sync_enabled
  prevent_search         = each.value.prevent_search
  tag_requests_with_user = each.value.tag_requests_with_user
}

resource "seerr_override_rule" "this" {
  for_each = var.override_rules

  genre             = each.value.genre
  language          = each.value.language
  root_folder       = each.value.root_folder
  radarr_service_id = each.value.radarr_server_key == null ? null : seerr_radarr_server.this[each.value.radarr_server_key].id
  sonarr_service_id = each.value.sonarr_server_key == null ? null : seerr_sonarr_server.this[each.value.sonarr_server_key].id
}

resource "seerr_notification_webhook" "this" {
  count = var.notification_webhook == null ? 0 : 1

  enabled            = var.notification_webhook.enabled
  embed_poster       = var.notification_webhook.embed_poster
  notification_types = var.notification_webhook.notification_types
  webhook            = local.notification_webhook
}
