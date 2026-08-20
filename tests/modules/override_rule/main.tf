terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "genre" {
  type    = string
  default = "18"
}

variable "language" {
  type    = string
  default = "en"
}

variable "keywords" {
  type    = string
  default = "heist"
}

resource "seerr_override_rule" "test" {
  genre    = var.genre
  language = var.language
  keywords = var.keywords
}

output "id" {
  value = seerr_override_rule.test.id
}

output "genre" {
  value = seerr_override_rule.test.genre
}

output "language" {
  value = seerr_override_rule.test.language
}

output "keywords" {
  value = seerr_override_rule.test.keywords
}
