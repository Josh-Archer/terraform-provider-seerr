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
  default = 1
}

variable "status" {
  type    = number
  default = 1
}

resource "seerr_issue" "test" {
  issue_type = var.issue_type
  message    = var.message
  media_id   = var.media_id
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
