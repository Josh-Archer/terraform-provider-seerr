variable "seerr_url" {
  type        = string
  description = "The public or base URL where Seerr is accessible."
}

variable "app_title" {
  type        = string
  default     = "Overseerr"
  description = "The title displayed across the Seerr web UI."
}

variable "locale" {
  type        = string
  default     = "en"
  description = "Default UI locale code (e.g. en, fr, de)."
}

variable "region" {
  type        = string
  default     = "US"
  description = "Default TMDB region code for metadata searches."
}

variable "original_language" {
  type        = string
  default     = "en"
  description = "Default TMDB language code for title and overview metadata."
}

variable "tv_metadata_provider" {
  type        = string
  default     = "tmdb"
  description = "Metadata provider for TV series (tmdb or tvdb)."
}

variable "anime_metadata_provider" {
  type        = string
  default     = "tmdb"
  description = "Metadata provider for anime series (tmdb or tvdb)."
}

variable "enable_plex" {
  type        = bool
  default     = true
  description = "Whether to configure Plex Media Server connection."
}

variable "plex_host" {
  type        = string
  default     = "plex.media.svc.cluster.local"
  description = "Hostname or IP address of the Plex server."
}

variable "plex_port" {
  type        = number
  default     = 32400
  description = "Port for the Plex server."
}

variable "plex_use_ssl" {
  type        = bool
  default     = false
  description = "Whether to connect to Plex over HTTPS."
}

variable "enable_jellyfin" {
  type        = bool
  default     = false
  description = "Whether to configure Jellyfin/Emby connection."
}

variable "jellyfin_host" {
  type        = string
  default     = "jellyfin.media.svc.cluster.local"
  description = "Hostname or IP address of the Jellyfin server."
}

variable "jellyfin_port" {
  type        = number
  default     = 8096
  description = "Port for the Jellyfin server."
}

variable "jellyfin_use_ssl" {
  type        = bool
  default     = false
  description = "Whether to connect to Jellyfin over HTTPS."
}

variable "jellyfin_api_key" {
  type        = string
  default     = ""
  sensitive   = true
  description = "API key for the Jellyfin server."
}

variable "enable_radarr" {
  type        = bool
  default     = true
  description = "Whether to register a Radarr movie server."
}

variable "radarr_server_name" {
  type        = string
  default     = "Radarr (HD)"
  description = "Display name for the Radarr server instance."
}

variable "radarr_url" {
  type        = string
  default     = "http://radarr-service.media.svc.cluster.local:7878"
  description = "Base URL of the Radarr instance used by Seerr provider data sources."
}

variable "radarr_host" {
  type        = string
  default     = "radarr-service.media.svc.cluster.local"
  description = "Hostname/IP of Radarr as reached by Seerr backend."
}

variable "radarr_port" {
  type        = number
  default     = 7878
  description = "Port of Radarr."
}

variable "radarr_api_key" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Radarr API key."
}

variable "radarr_use_ssl" {
  type        = bool
  default     = false
  description = "Whether to connect to Radarr via SSL."
}

variable "radarr_quality_profile_name" {
  type        = string
  default     = "HD-1080p"
  description = "Name of the quality profile to look up in Radarr."
}

variable "radarr_root_folder" {
  type        = string
  default     = ""
  description = "Root folder path in Radarr. If left empty, automatically uses the first discovered root folder."
}

variable "enable_sonarr" {
  type        = bool
  default     = true
  description = "Whether to register a Sonarr TV server."
}

variable "sonarr_server_name" {
  type        = string
  default     = "Sonarr (HD)"
  description = "Display name for the Sonarr server instance."
}

variable "sonarr_url" {
  type        = string
  default     = "http://sonarr-service.media.svc.cluster.local:8989"
  description = "Base URL of the Sonarr instance used by Seerr provider data sources."
}

variable "sonarr_host" {
  type        = string
  default     = "sonarr-service.media.svc.cluster.local"
  description = "Hostname/IP of Sonarr as reached by Seerr backend."
}

variable "sonarr_port" {
  type        = number
  default     = 8989
  description = "Port of Sonarr."
}

variable "sonarr_api_key" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Sonarr API key."
}

variable "sonarr_use_ssl" {
  type        = bool
  default     = false
  description = "Whether to connect to Sonarr via SSL."
}

variable "sonarr_quality_profile_name" {
  type        = string
  default     = "HD-1080p"
  description = "Name of the quality profile to look up in Sonarr."
}

variable "sonarr_root_folder" {
  type        = string
  default     = ""
  description = "Root folder path in Sonarr. If left empty, automatically uses the first discovered root folder."
}

variable "anime_root_folder" {
  type        = string
  default     = ""
  description = "Optional root folder path for routing Anime content in Sonarr."
}

variable "discord_webhook_url" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Discord incoming webhook URL for notifications. Omit to disable Discord notifications."
}

variable "discord_bot_name" {
  type        = string
  default     = "Overseerr Bot"
  description = "Custom username for Discord notification bot."
}

variable "ntfy_url" {
  type        = string
  default     = ""
  description = "Ntfy server URL (e.g. https://ntfy.sh or internal URL). Omit to disable Ntfy notifications."
}

variable "ntfy_topic" {
  type        = string
  default     = ""
  description = "Ntfy topic name."
}

variable "ntfy_token" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Optional authentication token for Ntfy."
}
