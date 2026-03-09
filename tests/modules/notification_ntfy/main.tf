terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "ntfy_url" {
  type = string
}

variable "topic" {
  type = string
}

variable "token" {
  type      = string
  sensitive = true
}

variable "priority" {
  type = number
}

resource "seerr_notification_ntfy" "test" {
  enabled      = true
  embed_poster = true

  ntfy = {
    url               = var.ntfy_url
    topic             = var.topic
    auth_method_token = true
    token             = var.token
    priority          = var.priority
  }

  on_request_pending = true
  on_issue_created   = true
}
