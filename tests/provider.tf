terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "url" {
  type        = string
  description = "The URL of the Overseerr instance"
  default     = "http://localhost:5055"
}

variable "api_key" {
  type        = string
  description = "The API key for the Overseerr instance"
  sensitive   = true
  default     = "dummy"
}

data "seerr_user" "admin" {
  email = "admin@example.com"
}

data "seerr_users" "all" {}

data "seerr_jobs" "all" {}

data "seerr_notification_agents" "all" {}

provider "seerr" {
  url     = var.url
  api_key = var.api_key
}
