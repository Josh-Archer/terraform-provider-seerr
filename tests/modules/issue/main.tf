terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "issue_type" {
  type    = number
  default = 4
}

variable "message" {
  type    = string
  default = "Test issue message"
}

variable "media_id" {
  type    = number
  default = null
}

variable "status" {
  type    = number
  default = 1
}

resource "seerr_request" "media_item" {
  count      = var.media_id == null ? 1 : 0
  media_type = "movie"
  media_id   = 550
  status     = 2
}

resource "seerr_issue" "test" {
  issue_type = var.issue_type
  message    = var.message
  media_id   = coalesce(var.media_id, try(seerr_request.media_item[0].media_id, null))
  status     = var.status
}

output "id" {
  value = seerr_issue.test.id
}

output "issue_type" {
  value = seerr_issue.test.issue_type
}

output "message" {
  value = seerr_issue.test.message
}

output "media_id" {
  value = seerr_issue.test.media_id
}

output "status" {
  value = seerr_issue.test.status
}
