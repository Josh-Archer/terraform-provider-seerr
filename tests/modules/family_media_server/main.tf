terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

module "family" {
  source = "../../../modules/family_media_server"

  users = {
    integration = {
      username       = var.username
      email          = var.email
      permission_set = "standard"
      quota = {
        movie_limit = 2
        movie_days  = 7
        tv_limit    = 1
        tv_days     = 7
      }
    }
  }
}

variable "username" { type = string }
variable "email" { type = string }

output "user_id" {
  value = module.family.user_ids.integration
}

output "permission_bitmask" {
  value = module.family.permission_bitmasks.integration
}

