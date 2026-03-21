terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
variable "email" { type = string }

data "seerr_user" "test" {
  email = var.email
}

output "id" {
  value = data.seerr_user.test.id
}

