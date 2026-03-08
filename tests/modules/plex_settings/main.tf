terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "ip" { type = string }
variable "port" { type = number }
variable "use_ssl" { type = bool }

resource "seerr_plex_settings" "test" {
  ip      = var.ip
  port    = var.port
  use_ssl = var.use_ssl
}

