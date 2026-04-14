terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
variable "ip" { type = string }
variable "port" { type = number }
variable "use_ssl" { type = bool }
variable "enabled" {
  type    = bool
  default = false
}

resource "seerr_plex_settings" "test" {
  count   = var.enabled ? 1 : 0
  ip      = var.ip
  port    = var.port
  use_ssl = var.use_ssl
}
