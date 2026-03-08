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
}

variable "api_key" {
  type        = string
  description = "The API key for the Overseerr instance"
  sensitive   = true
}

provider "seerr" {
  url     = var.url
  api_key = var.api_key
}
