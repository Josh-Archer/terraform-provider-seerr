terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "schedule" {
  type = string
}

variable "retention" {
  type = number
}

variable "storage_path" {
  type = string
}

variable "backup_settings_supported" {
  type    = bool
  default = true
}

resource "seerr_backup_settings" "test" {
  count = var.backup_settings_supported ? 1 : 0

  schedule     = var.schedule
  retention    = var.retention
  storage_path = var.storage_path
}

output "id" {
  value = var.backup_settings_supported ? seerr_backup_settings.test[0].id : null
}

output "supported" {
  value = var.backup_settings_supported
}

output "schedule" {
  value = var.backup_settings_supported ? seerr_backup_settings.test[0].schedule : null
}
