variable "seerr_url" {
  description = "Base URL of the Seerr / Jellyseerr / Overseerr instance (e.g. http://localhost:5055)."
  type        = string
  default     = "http://localhost:5055"
}

variable "seerr_api_key" {
  description = "API Key for authenticating against the Seerr instance."
  type        = string
  sensitive   = true
}
