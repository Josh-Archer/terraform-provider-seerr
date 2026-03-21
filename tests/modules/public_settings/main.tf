terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
data "seerr_public_settings" "test" {}

output "initialized" {
  value = data.seerr_public_settings.test.initialized
}

output "app_title" {
  value = data.seerr_public_settings.test.app_title
}
