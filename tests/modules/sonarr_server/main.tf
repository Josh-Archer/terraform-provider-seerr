terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "name" { type = string }
variable "hostname" { type = string }
variable "port" { type = number }
variable "api_key" { type = string }
variable "quality_profile_id" {
  type    = number
  default = 1
}
variable "active_directory" {
  type    = string
  default = "/tv"
}

resource "seerr_sonarr_server" "test" {
  name               = var.name
  hostname           = var.hostname
  port               = var.port
  api_key            = var.api_key
  quality_profile_id = var.quality_profile_id
  active_directory   = var.active_directory
}

output "id" {
  value = seerr_sonarr_server.test.id
}

output "hostname" { value = seerr_sonarr_server.test.hostname }
output "port"     { value = seerr_sonarr_server.test.port }
output "quality_profile_id" { value = seerr_sonarr_server.test.quality_profile_id }
output "is_default" { value = seerr_sonarr_server.test.is_default }
output "sync_enabled" { value = seerr_sonarr_server.test.sync_enabled }


