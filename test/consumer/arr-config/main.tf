# Sanitized replay of the seerr_* resources used by the consumer repo's
# terraform/arr-config root module (Josh-Archer/home). Hostnames, URLs and
# secrets are test values; the provider source/version are injected by
# scripts/test-consumer-replay.sh. Resources that need a live Plex server
# (seerr_plex_library_*) and the Radarr/Sonarr quality-profile data sources
# (they need a reachable Arr API from the runner) are intentionally omitted;
# the profile id/name pair below matches the mock Arr fixtures.

provider "seerr" {
  url                     = var.seerr_url
  api_key                 = var.seerr_api_key
  request_timeout_seconds = 120
}

locals {
  seerr_managed_job_schedules = {
    for job in data.seerr_jobs.all.jobs : job.id => job.schedule
    if job.enabled && trimspace(job.schedule) != ""
  }
}

data "seerr_current_user" "admin" {}

data "seerr_about" "system" {}

data "seerr_metadata_settings" "current" {}

data "seerr_watchlist" "current" {}

data "seerr_jobs" "all" {}

data "seerr_service_status" "status" {}

resource "seerr_network_settings" "network" {
  csrf_protection = false
  trust_proxy     = false
}

resource "seerr_main_settings" "main" {
  app_title          = "Seerr"
  application_url    = "https://seerr.example.invalid"
  hide_available     = true
  cache_images       = true
  local_login        = true
  locale             = "en"
  original_language  = ""
  partial_requests   = true
  media_server_login = true
  new_plex_login     = true
  streaming_region   = "US"
  discover_region    = "US"
}

data "seerr_permission_set" "admin" {
  admin           = true
  manage_users    = true
  manage_requests = true
}

data "seerr_permission_set" "standard_user" {
  request       = true
  request_movie = true
  request_tv    = true
}

data "seerr_permission_set" "power_user" {
  request       = true
  request_movie = true
  request_tv    = true
  request_4k    = true
  auto_approve  = true
}

resource "seerr_user_permissions" "admin" {
  user_id     = tonumber(data.seerr_current_user.admin.id)
  permissions = data.seerr_permission_set.admin.permissions
}

resource "seerr_user_watchlist_settings" "admin" {
  user_id               = tonumber(data.seerr_current_user.admin.id)
  watchlist_sync_movies = true
  watchlist_sync_tv     = true
}

# Admin quota: Seerr ignores quota writes for users holding MANAGE_USERS, so
# the admin always reads back null (global default). All four quota attributes
# stay unset, exactly as the consumer does (see provider #344 / #349).
resource "seerr_user_quota" "admin" {
  user_id = tonumber(data.seerr_current_user.admin.id)
}

resource "seerr_discover_slider" "discover" {
  dynamic "sliders" {
    for_each = range(1, 13)
    content {
      type    = sliders.value
      enabled = true
    }
  }
}

resource "seerr_job_schedule" "managed" {
  for_each = local.seerr_managed_job_schedules

  job_id   = each.key
  schedule = each.value
}

resource "seerr_radarr_server" "radarr" {
  name                   = "radarr"
  hostname               = var.radarr_hostname
  port                   = 7878
  api_key                = var.radarr_api_key
  use_ssl                = false
  quality_profile_id     = 1
  quality_profile_name   = var.radarr_quality_profile_name
  active_directory       = var.movie_root_folder
  minimum_availability   = "announced"
  is_4k                  = false
  is_default             = true
  enable_scan            = true
  sync_enabled           = true
  prevent_search         = false
  tag_requests_with_user = true
  tags                   = []
}

resource "seerr_sonarr_server" "sonarr" {
  name                   = "sonarr"
  hostname               = var.sonarr_hostname
  port                   = 8989
  api_key                = var.sonarr_api_key
  use_ssl                = false
  quality_profile_id     = 1
  quality_profile_name   = var.sonarr_quality_profile_name
  active_directory       = var.tv_root_folder
  active_anime_directory = var.tv_root_folder
  is_4k                  = false
  is_default             = true
  enable_scan            = true
  enable_season_folders  = true
  sync_enabled           = true
  prevent_search         = false
  tag_requests_with_user = true
  tags                   = []
  anime_tags             = []
}

resource "seerr_notification_ntfy" "ntfy" {
  enabled      = true
  embed_poster = true
  notification_types = toset([
    "MEDIA_PENDING",
    "MEDIA_APPROVED",
    "MEDIA_AVAILABLE",
    "MEDIA_FAILED",
    "MEDIA_DECLINED",
    "MEDIA_AUTO_APPROVED",
  ])
  ntfy = {
    url               = var.ntfy_server_url
    topic             = var.ntfy_topic_media
    auth_method_token = true
    token             = var.ntfy_token
  }
}

output "seerr_service_status" {
  value = data.seerr_service_status.status.version
}

output "seerr_admin_permissions" {
  description = "Bitmask composed by seerr_permission_set for the admin user (expected 26)."
  value       = data.seerr_permission_set.admin.permissions
}

output "seerr_standard_user_permissions" {
  value = data.seerr_permission_set.standard_user.permissions
}

output "seerr_power_user_permissions" {
  value = data.seerr_permission_set.power_user.permissions
}
