variable "seerr_url" {
  type        = string
  description = "Base URL of your Seerr / Jellyseerr / Overseerr instance"
  default     = "http://localhost:5055"
}

variable "seerr_api_key" {
  type        = string
  description = "API Key from Seerr Settings -> General"
  sensitive   = true
}

variable "app_title" {
  type        = string
  description = "Custom title for Seerr UI"
  default     = "Homelab Media Requests"
}

variable "admin_email" {
  type        = string
  description = "Email address for primary admin user"
}

variable "admin_username" {
  type        = string
  description = "Username for primary admin user"
  default     = "admin"
}

# Sonarr Settings
variable "sonarr_hostname" {
  type        = string
  description = "Hostname or IP of Sonarr"
  default     = "sonarr.local"
}

variable "sonarr_port" {
  type        = number
  description = "Port for Sonarr"
  default     = 8989
}

variable "sonarr_api_key" {
  type        = string
  description = "Sonarr API Key"
  sensitive   = true
}

variable "sonarr_use_ssl" {
  type    = bool
  default = false
}

variable "sonarr_quality_profile_id" {
  type        = number
  description = "Quality profile ID in Sonarr (e.g. 1 for HD - 1080p)"
  default     = 1
}

variable "sonarr_root_folder" {
  type        = string
  description = "Root storage directory for TV series in Sonarr"
  default     = "/data/media/tv"
}

# Radarr Settings
variable "radarr_hostname" {
  type        = string
  description = "Hostname or IP of Radarr"
  default     = "radarr.local"
}

variable "radarr_port" {
  type        = number
  description = "Port for Radarr"
  default     = 7878
}

variable "radarr_api_key" {
  type        = string
  description = "Radarr API Key"
  sensitive   = true
}

variable "radarr_use_ssl" {
  type    = bool
  default = false
}

variable "radarr_quality_profile_id" {
  type        = number
  description = "Quality profile ID in Radarr (e.g. 1 for HD - 1080p)"
  default     = 1
}

variable "radarr_root_folder" {
  type        = string
  description = "Root storage directory for movies in Radarr"
  default     = "/data/media/movies"
}

# Notification Settings
variable "enable_discord_notifications" {
  type    = bool
  default = true
}

variable "discord_webhook_url" {
  type        = string
  description = "Discord Webhook URL for media request notifications"
  default     = ""
  sensitive   = true
}

variable "enable_pushover_notifications" {
  type    = bool
  default = false
}

variable "pushover_user_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "pushover_app_token" {
  type      = string
  default   = ""
  sensitive = true
}
