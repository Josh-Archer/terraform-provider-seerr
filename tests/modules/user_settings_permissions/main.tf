terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

resource "seerr_user_settings_permissions" "test" {
  user_id                = var.user_id
  auto_approve_movies    = var.auto_approve_movies
  auto_approve_tv        = var.auto_approve_tv
  auto_approve_4k_movies = var.auto_approve_4k_movies
  auto_approve_4k_tv     = var.auto_approve_4k_tv
}

output "id" {
  value = seerr_user_settings_permissions.test.id
}

output "user_id" {
  value = seerr_user_settings_permissions.test.user_id
}
