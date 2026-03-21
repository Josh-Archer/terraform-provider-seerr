terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
variable "application_url" { type = string }
variable "application_title" { type = string }

resource "seerr_main_settings" "test" {
  application_url = var.application_url
  app_title       = var.application_title
}

output "application_url" {
  value = seerr_main_settings.test.application_url
}

output "app_title" {
  value = seerr_main_settings.test.app_title
}

