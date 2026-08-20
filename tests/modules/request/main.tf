terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "media_type" {
  type    = string
  default = "movie"
}

variable "media_id" {
  type    = number
  default = 550
}

variable "is_4k" {
  type    = bool
  default = false
}

variable "status" {
  type    = number
  default = 1
}

resource "seerr_request" "test" {
  media_type = var.media_type
  media_id   = var.media_id
  is_4k      = var.is_4k
  status     = var.status
}

output "id" {
  value = seerr_request.test.id
}

output "media_type" {
  value = seerr_request.test.media_type
}

output "media_id" {
  value = seerr_request.test.media_id
}

output "is_4k" {
  value = seerr_request.test.is_4k
}

output "status" {
  value = seerr_request.test.status
}
