terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "url" {
  type = string
}

variable "api_key" {
  type      = string
  sensitive = true
}

variable "username" {
  type = string
  default = "ci-admin"
}

variable "email" {
  type = string
  default = "ci-admin@example.invalid"
}

provider "seerr" {
  url     = var.url
  api_key = var.api_key
}

data "seerr_user" "admin" {
  email = "ci-admin@example.invalid"
}
