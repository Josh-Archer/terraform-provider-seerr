terraform {
  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

resource "seerr_main_settings" "this" {
  app_title               = var.app_title
  application_url         = var.application_url
  locale                  = var.locale
  region                  = var.region
  original_language       = var.original_language
  movie_requests_enabled  = var.movie_requests_enabled
  series_requests_enabled = var.series_requests_enabled
}

output "resource_id" {
  value = seerr_main_settings.this.id
}
