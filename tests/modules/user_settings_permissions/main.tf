terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
data "seerr_user" "test" {
  count = var.user_id == null ? 1 : 0
  email = var.email
}

locals {
  resolved_user_id = var.user_id != null ? var.user_id : tonumber(data.seerr_user.test[0].id)
}

resource "seerr_user_settings_permissions" "test" {
  user_id                = local.resolved_user_id
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
