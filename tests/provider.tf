terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "username" {
  type = string
  default = "ci-admin"
}

variable "email" {
  type = string
  default = "ci-admin@example.invalid"
}

data "seerr_user" "admin" {
  email = "ci-admin@example.invalid"
}
