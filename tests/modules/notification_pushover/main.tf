terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "access_token" {
  type      = string
  sensitive = true
}

variable "user_token" {
  type      = string
  sensitive = true
}

variable "sound" {
  type = string
}

resource "seerr_notification_pushover" "test" {
  enabled = true

  pushover = {
    access_token = var.access_token
    user_token   = var.user_token
    sound        = var.sound
  }

  on_request_pending = true
  on_issue_created   = true
}
