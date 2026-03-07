variable "name" {
  type    = string
  default = "Sonarr"
}

variable "url" {
  type    = string
  default = null
}

variable "hostname" {
  type    = string
  default = "sonarr-service"
}

variable "port" {
  type    = number
  default = 8989
}

variable "api_key" {
  type      = string
  sensitive = true
}

variable "use_ssl" {
  type    = bool
  default = false
}

variable "base_url" {
  type    = string
  default = ""
}

variable "quality_profile_id" {
  type = number
}

variable "active_directory" {
  type = string
}

variable "active_anime_directory" {
  type    = string
  default = null
}

variable "tags" {
  type    = list(number)
  default = []
}

variable "anime_tags" {
  type    = list(number)
  default = []
}

variable "is_4k" {
  type    = bool
  default = false
}

variable "is_default" {
  type    = bool
  default = true
}

variable "enable_scan" {
  type    = bool
  default = false
}

variable "enable_season_folders" {
  type    = bool
  default = true
}

variable "sync_enabled" {
  type    = bool
  default = true
}

variable "prevent_search" {
  type    = bool
  default = false
}

variable "tag_requests_with_user" {
  type    = bool
  default = true
}

variable "extra_payload" {
  type    = map(any)
  default = {}
}
