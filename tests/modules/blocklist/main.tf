terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "tmdb_id" {
  type    = number
  default = 438631
}

variable "media_type" {
  type    = string
  default = "movie"
}

variable "title" {
  type    = string
  default = "Dune"
}

variable "user_id" {
  type    = number
  default = 1
}

resource "seerr_blocklist" "test" {
  tmdb_id    = var.tmdb_id
  media_type = var.media_type
  title      = var.title
  user_id    = var.user_id
}

output "id" {
  value = seerr_blocklist.test.id
}

output "tmdb_id" {
  value = seerr_blocklist.test.tmdb_id
}

output "media_type" {
  value = seerr_blocklist.test.media_type
}

output "title" {
  value = seerr_blocklist.test.title
}

output "user_id" {
  value = seerr_blocklist.test.user_id
}
