provider "seerr" {
  url     = var.seerr_url
  api_key = var.seerr_api_key
}

# These locals simulate values coming from existing Radarr/Sonarr providers.
# In real usage, replace with direct references to those provider resources/outputs.
locals {
  radarr_url     = var.radarr_url
  radarr_api_key = var.radarr_api_key
  sonarr_url     = var.sonarr_url
  sonarr_api_key = var.sonarr_api_key
}

data "seerr_radarr_quality_profile" "movies" {
  url     = local.radarr_url
  api_key = local.radarr_api_key
  name    = var.radarr_quality_profile_name
}

data "seerr_sonarr_quality_profile" "shows" {
  url     = local.sonarr_url
  api_key = local.sonarr_api_key
  name    = var.sonarr_quality_profile_name
}

module "radarr_server" {
  source = "../../../modules/radarr_server"

  url                = local.radarr_url
  api_key            = local.radarr_api_key
  quality_profile_id = data.seerr_radarr_quality_profile.movies.quality_profile_id
  active_directory   = var.radarr_root
}

module "sonarr_server" {
  source = "../../../modules/sonarr_server"

  url                = local.sonarr_url
  api_key            = local.sonarr_api_key
  quality_profile_id = data.seerr_sonarr_quality_profile.shows.quality_profile_id
  active_directory   = var.sonarr_root
}

module "ntfy_notification" {
  source = "../../../modules/notification_agent"

  agent = "ntfy"
  payload = {
    enabled = true
    types = {
      MEDIA_APPROVED  = true
      MEDIA_AVAILABLE = true
      MEDIA_FAILED    = true
    }
    options = {
      serverUrl   = var.ntfy_server_url
      topic       = var.ntfy_topic
      accessToken = var.ntfy_access_token
      priority    = 3
    }
  }
}

module "main_settings" {
  source = "../../../modules/main_settings"

  payload = {
    applicationTitle = "Seerr"
    locale           = "en"
  }
}

module "plex_settings" {
  source = "../../../modules/plex_settings"

  payload = {
    hostname = "plex.media.svc.cluster.local"
    port     = 32400
    useSsl   = false
  }
}

variable "seerr_url" {
  type = string
}

variable "seerr_api_key" {
  type      = string
  sensitive = true
}

variable "radarr_url" {
  type = string
}

variable "radarr_api_key" {
  type      = string
  sensitive = true
}

variable "radarr_quality_profile_name" {
  type = string
}

variable "radarr_root" {
  type = string
}

variable "sonarr_url" {
  type = string
}

variable "sonarr_api_key" {
  type      = string
  sensitive = true
}

variable "sonarr_quality_profile_name" {
  type = string
}

variable "sonarr_root" {
  type = string
}

variable "ntfy_server_url" {
  type = string
}

variable "ntfy_topic" {
  type = string
}

variable "ntfy_access_token" {
  type      = string
  sensitive = true
}
