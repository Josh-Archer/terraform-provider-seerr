variable "radarr_servers" {
  description = "Radarr registrations keyed by routing name. Use multiple entries to route distinct quality profiles or 4K libraries."
  type = map(object({
    name                   = string
    hostname               = string
    api_key                = string
    quality_profile_id     = number
    root_folder            = string
    port                   = optional(number, 7878)
    use_ssl                = optional(bool, false)
    base_url               = optional(string, "")
    minimum_availability   = optional(string, "announced")
    tags                   = optional(list(number), [])
    is_4k                  = optional(bool, false)
    is_default             = optional(bool, false)
    enable_scan            = optional(bool, false)
    sync_enabled           = optional(bool, true)
    prevent_search         = optional(bool, false)
    tag_requests_with_user = optional(bool, true)
  }))
  default = {}
}

variable "sonarr_servers" {
  description = "Sonarr registrations keyed by routing name. Use multiple entries to route distinct quality profiles, anime, or 4K libraries."
  type = map(object({
    name                   = string
    hostname               = string
    api_key                = string
    quality_profile_id     = number
    root_folder            = string
    port                   = optional(number, 8989)
    use_ssl                = optional(bool, false)
    base_url               = optional(string, "")
    anime_root_folder      = optional(string)
    tags                   = optional(list(number), [])
    anime_tags             = optional(list(number), [])
    is_4k                  = optional(bool, false)
    is_default             = optional(bool, false)
    enable_scan            = optional(bool, false)
    enable_season_folders  = optional(bool, true)
    sync_enabled           = optional(bool, true)
    prevent_search         = optional(bool, false)
    tag_requests_with_user = optional(bool, true)
  }))
  default = {}
}

variable "override_rules" {
  description = "Content-routing overrides keyed by a stable name. Reference a Radarr or Sonarr input key."
  type = map(object({
    genre             = optional(string)
    language          = optional(string)
    root_folder       = string
    radarr_server_key = optional(string)
    sonarr_server_key = optional(string)
  }))
  default = {}

  validation {
    condition = alltrue([
      for rule in values(var.override_rules) : (rule.radarr_server_key == null) != (rule.sonarr_server_key == null)
    ])
    error_message = "Each override rule must reference exactly one Radarr or Sonarr server key."
  }
}

variable "notification_webhook" {
  description = "Optional global webhook notification route. Seerr exposes one webhook agent, so only one module should manage it."
  type = object({
    url                = string
    json_payload       = string
    auth_header        = optional(string)
    enabled            = optional(bool, true)
    embed_poster       = optional(bool, true)
    notification_types = optional(set(string), ["MEDIA_APPROVED", "MEDIA_AVAILABLE", "MEDIA_FAILED"])
  })
  default = null
}
