provider "seerr" {
  url     = "https://seerr.example.com"
  api_key = var.seerr_api_key
}

variable "seerr_api_key" {
  type      = string
  sensitive = true
}

