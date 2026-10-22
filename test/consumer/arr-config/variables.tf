variable "seerr_url" {
  type        = string
  description = "Base URL of the disposable Seerr instance under test."
}

variable "seerr_api_key" {
  type        = string
  description = "API key of the disposable Seerr instance under test."
  sensitive   = true
}

variable "radarr_hostname" {
  type    = string
  default = "radarr-mock"
}

variable "radarr_api_key" {
  type      = string
  sensitive = true
  default   = "radarr_api_key" # test value accepted by tests/fixtures/mock_media_service.py
}

variable "sonarr_hostname" {
  type    = string
  default = "sonarr-mock"
}

variable "sonarr_api_key" {
  type      = string
  sensitive = true
  default   = "sonarr_api_key" # test value accepted by tests/fixtures/mock_media_service.py
}

variable "movie_root_folder" {
  type    = string
  default = "/media/radarr"
}

variable "tv_root_folder" {
  type    = string
  default = "/media/sonarr"
}

variable "radarr_quality_profile_name" {
  type    = string
  default = "HD-1080p"
}

variable "sonarr_quality_profile_name" {
  type    = string
  default = "HD-1080p"
}

variable "ntfy_server_url" {
  type    = string
  default = "https://ntfy.example.invalid"
}

variable "ntfy_topic_media" {
  type    = string
  default = "media-test"
}

variable "ntfy_token" {
  type      = string
  sensitive = true
  default   = "tk_testtoken000000000000000000000"
}
