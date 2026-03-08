terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "application_url" { type = string }
variable "application_title" { type = string }

resource "seerr_main_settings" "test" {
  application_url   = var.application_url
  application_title = var.application_title
}

output "application_url" {
  value = seerr_main_settings.test.application_url
}

