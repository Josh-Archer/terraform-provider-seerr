terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = ">= 0.37.0"
    }
  }
}

# --------------------------------------------------------------------------
# 1. Main & System Settings
# --------------------------------------------------------------------------

resource "seerr_main_settings" "main" {
  app_title         = var.app_title
  application_url   = var.seerr_url
  locale            = var.locale
  region            = var.region
  original_language = var.original_language
  hide_available    = false
}

resource "seerr_metadata_settings" "metadata" {
  tv    = var.tv_metadata_provider
  anime = var.anime_metadata_provider
}

# --------------------------------------------------------------------------
# 2. Media Server Connection (Plex or Jellyfin)
# --------------------------------------------------------------------------

resource "seerr_plex_settings" "plex" {
  count   = var.enable_plex ? 1 : 0
  ip      = var.plex_host
  port    = var.plex_port
  use_ssl = var.plex_use_ssl
}

resource "seerr_jellyfin_settings" "jellyfin" {
  count   = var.enable_jellyfin ? 1 : 0
  ip      = var.jellyfin_host
  port    = var.jellyfin_port
  use_ssl = var.jellyfin_use_ssl
  api_key = var.jellyfin_api_key
}

# --------------------------------------------------------------------------
# 3. Dynamic *arr Entity Resolution
# --------------------------------------------------------------------------

data "seerr_radarr_quality_profile" "movies" {
  count   = var.enable_radarr ? 1 : 0
  url     = var.radarr_url
  api_key = var.radarr_api_key
  name    = var.radarr_quality_profile_name
}

data "seerr_radarr_root_folders" "movies" {
  count   = var.enable_radarr ? 1 : 0
  url     = var.radarr_url
  api_key = var.radarr_api_key
}

data "seerr_sonarr_quality_profile" "shows" {
  count   = var.enable_sonarr ? 1 : 0
  url     = var.sonarr_url
  api_key = var.sonarr_api_key
  name    = var.sonarr_quality_profile_name
}

data "seerr_sonarr_root_folders" "shows" {
  count   = var.enable_sonarr ? 1 : 0
  url     = var.sonarr_url
  api_key = var.sonarr_api_key
}

# --------------------------------------------------------------------------
# 4. Servarr Service Registrations
# --------------------------------------------------------------------------

resource "seerr_radarr_server" "radarr" {
  count              = var.enable_radarr ? 1 : 0
  name               = var.radarr_server_name
  hostname           = var.radarr_host
  port               = var.radarr_port
  api_key            = var.radarr_api_key
  use_ssl            = var.radarr_use_ssl
  is_default         = true
  is_4k              = false
  quality_profile_id = data.seerr_radarr_quality_profile.movies[0].quality_profile_id
  active_directory   = var.radarr_root_folder != "" ? var.radarr_root_folder : data.seerr_radarr_root_folders.movies[0].root_folders[0].path
}

resource "seerr_sonarr_server" "sonarr" {
  count              = var.enable_sonarr ? 1 : 0
  name               = var.sonarr_server_name
  hostname           = var.sonarr_host
  port               = var.sonarr_port
  api_key            = var.sonarr_api_key
  use_ssl            = var.sonarr_use_ssl
  is_default         = true
  is_4k              = false
  quality_profile_id = data.seerr_sonarr_quality_profile.shows[0].quality_profile_id
  active_directory   = var.sonarr_root_folder != "" ? var.sonarr_root_folder : data.seerr_sonarr_root_folders.shows[0].root_folders[0].path
}

# --------------------------------------------------------------------------
# 5. Content Routing & Override Rules
# --------------------------------------------------------------------------

resource "seerr_override_rule" "anime" {
  count             = var.enable_sonarr && var.anime_root_folder != "" ? 1 : 0
  genre             = "16" # Animation
  language          = "ja"
  root_folder       = var.anime_root_folder
  sonarr_service_id = seerr_sonarr_server.sonarr[0].id
}

# --------------------------------------------------------------------------
# 6. Discover Sliders Customization
# --------------------------------------------------------------------------

resource "seerr_discover_slider" "home" {
  sliders {
    type    = 1 # Recently Added TV
    enabled = true
  }
  sliders {
    type    = 2 # Recently Added Movies
    enabled = true
  }
  sliders {
    type    = 4 # Trending
    enabled = true
  }
  sliders {
    type    = 5 # Popular Movies
    enabled = true
  }
  sliders {
    type    = 6 # Popular TV
    enabled = true
  }
}

# --------------------------------------------------------------------------
# 7. Notification Agent Configuration
# --------------------------------------------------------------------------

resource "seerr_notification_discord" "alerts" {
  count   = var.discord_webhook_url != "" ? 1 : 0
  enabled = true

  discord = {
    webhook_url  = var.discord_webhook_url
    bot_username = var.discord_bot_name
  }

  notification_types = [
    "MEDIA_PENDING",
    "MEDIA_APPROVED",
    "MEDIA_AUTO_APPROVED",
    "MEDIA_AVAILABLE",
    "MEDIA_FAILED",
    "ISSUE_CREATED",
    "ISSUE_RESOLVED"
  ]
}

resource "seerr_notification_ntfy" "alerts" {
  count   = var.ntfy_url != "" ? 1 : 0
  enabled = true

  ntfy = {
    url               = var.ntfy_url
    topic             = var.ntfy_topic
    auth_method_token = var.ntfy_token != ""
    token             = var.ntfy_token
    priority          = 3
  }

  notification_types = [
    "MEDIA_APPROVED",
    "MEDIA_AVAILABLE",
    "MEDIA_FAILED"
  ]
}
