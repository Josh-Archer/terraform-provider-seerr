# Standard setup using an API key
provider "seerr" {
  url     = "https://seerr.example.com"
  api_key = var.seerr_api_key
}

variable "seerr_api_key" {
  type      = string
  sensitive = true
}

# First-run setup using a Plex token to bootstrap the API key
provider "seerr" {
  url        = "https://seerr.example.com"
  plex_token = var.plex_token
}

variable "plex_token" {
  type      = string
  sensitive = true
}
