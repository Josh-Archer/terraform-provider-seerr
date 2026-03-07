terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

provider "seerr" {
  # Configuration is usually pulled from environment variables:
  # SEERR_URL
  # SEERR_API_KEY
}
