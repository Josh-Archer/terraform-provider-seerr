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

resource "seerr_backup_settings" "test" {
  schedule     = var.schedule
  retention    = var.retention
  storage_path = var.storage_path
}

output "id" {
  value = seerr_backup_settings.test.id
}

