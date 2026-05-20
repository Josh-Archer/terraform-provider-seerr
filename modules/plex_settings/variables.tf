variable "ip" {
  description = "Plex server IP address or hostname."
  type        = string
}

variable "port" {
  description = "Plex server port."
  type        = number
}

variable "use_ssl" {
  description = "Whether to connect to Plex with HTTPS."
  type        = bool
  default     = false
}
