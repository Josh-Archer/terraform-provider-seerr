variable "name" {
  type    = string
  default = "Radarr"
}

variable "url" {
  type    = string
  default = null
}

variable "hostname" {
  type    = string
  default = "radarr-service"
}

variable "port" {
  type    = number
  default = 7878
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

variable "is_4k" {
  type    = bool
  default = false
}

variable "minimum_availability" {
  type    = string
  default = "announced"
}

variable "tags" {
  type    = list(number)
  default = []
}

variable "is_default" {
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

variable "tag_requests" {
  type    = bool
  default = true
}

variable "extra_payload" {
  type    = map(any)
  default = {}
}
