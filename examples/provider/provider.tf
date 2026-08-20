# Standard OpenTofu / Terraform configuration
terraform {
  required_providers {
    seerr = {
      source  = "registry.opentofu.org/josh-archer/seerr"
      # For HashiCorp Terraform Registry, use:
      # source = "josh-archer/seerr"
      version = "~> 0.38.0"
    }
  }
}

provider "seerr" {
  url     = "https://seerr.example.com"
  api_key = var.seerr_api_key
}

variable "seerr_api_key" {
  type        = string
  description = "Seerr API Key"
  sensitive   = true
}

# First-run setup using a Plex token to bootstrap the initial API key:
# provider "seerr" {
#   url        = "https://seerr.example.com"
#   plex_token = var.plex_token
# }
