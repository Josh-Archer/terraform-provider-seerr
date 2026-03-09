terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "hostname" { type = string }
variable "port" { type = number }
variable "api_key" { type = string }

resource "seerr_tautulli_settings" "test" {
  hostname = var.hostname
  port     = var.port
  api_key  = var.api_key
}

output "id" {
  value = seerr_tautulli_settings.test.id
}
