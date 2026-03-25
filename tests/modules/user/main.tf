terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
variable "username" { type = string }
variable "email" { type = string }
variable "permissions" { type = number }

resource "seerr_user" "test" {
  username    = var.username
  email       = var.email
  permissions = var.permissions
}

output "id" {
  value = seerr_user.test.id
}

output "username" {
  value = seerr_user.test.username
}

