
terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "ip" {
  type = string
}

variable "port" {
  type = number
}

variable "api_key" {
  type      = string
  sensitive = true
}

resource "seerr_jellyfin_settings" "test" {
  ip      = var.ip
  port    = var.port
  api_key = var.api_key
}
