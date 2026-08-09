# Complete Media Stack IaC with OpenTofu / Terraform Provider for Seerr
# Connects Seerr to Radarr, Sonarr, Notification Agents, and sets up User Tiers

terraform {
  required_version = ">= 1.6.0"
  required_providers {
    seerr = {
      source  = "Josh-Archer/seerr"
      version = "~> 0.31.1"
    }
  }
}

provider "seerr" {
  url     = var.seerr_url
  api_key = var.seerr_api_key
}

# 1. Main Seerr Application Configuration
resource "seerr_main_settings" "config" {
  api_key          = var.seerr_api_key
  application_title = var.app_title
  application_url   = var.seerr_url
  trust_proxy      = true
  hide_available   = false
  csrf_protection  = true
}

# 2. Sonarr Server Connection (TV Shows)
resource "seerr_sonarr_server" "primary_sonarr" {
  name                = "Sonarr Primary (HD/4K)"
  hostname            = var.sonarr_hostname
  port                = var.sonarr_port
  api_key             = var.sonarr_api_key
  use_ssl             = var.sonarr_use_ssl
  active_profile_id   = var.sonarr_quality_profile_id
  active_directory    = var.sonarr_root_folder
  is_default          = true
  is_4k               = false
  enable_season_folders = true
}

# 3. Radarr Server Connection (Movies)
resource "seerr_radarr_server" "primary_radarr" {
  name              = "Radarr Primary (HD/4K)"
  hostname          = var.radarr_hostname
  port              = var.radarr_port
  api_key           = var.radarr_api_key
  use_ssl           = var.radarr_use_ssl
  active_profile_id = var.radarr_quality_profile_id
  active_directory  = var.radarr_root_folder
  is_default        = true
  is_4k             = false
  minimum_availability = "announced"
}

# 4. Homelab Admin User
resource "seerr_user" "admin" {
  email       = var.admin_email
  username    = var.admin_username
  permissions = 2097150 # Full Admin Privileges
}

# 5. Standard Power User Tier (Automatic Approvals & Higher Quotas)
resource "seerr_user" "power_user" {
  email    = "poweruser@example.com"
  username = "poweruser"
  permissions = (
    2    # MANAGE_REQUESTS
    + 32   # AUTO_APPROVE
    + 64   # AUTO_APPROVE_MOVIE
    + 128  # AUTO_APPROVE_TV
    + 256  # AUTO_APPROVE_4K
    + 512  # AUTO_APPROVE_4K_MOVIE
    + 1024 # AUTO_APPROVE_4K_TV
  )
}

resource "seerr_user_quota" "power_user_quota" {
  user_id           = seerr_user.power_user.id
  movie_quota_limit = 20
  movie_quota_days  = 7
  tv_quota_limit    = 10
  tv_quota_days     = 7
}

# 6. Discord Webhook Notification Agent
resource "seerr_notification_discord" "alerts" {
  enabled     = var.enable_discord_notifications
  webhook_url = var.discord_webhook_url
  types       = 1048575 # Notify on all events
}

# 7. Pushover Notification Agent for Critical Alerts
resource "seerr_notification_pushover" "admin_push" {
  count       = var.enable_pushover_notifications ? 1 : 0
  enabled     = true
  user_token  = var.pushover_user_key
  token       = var.pushover_app_token
  types       = 2050 # Media Auto-Approved + Issue Created
}
