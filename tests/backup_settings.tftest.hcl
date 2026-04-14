variables {
  backup_settings_supported = true
}

run "backup_settings_lifecycle" {
  command = apply

  variables {
    backup_settings_supported = var.backup_settings_supported
    schedule                  = "0 2 * * *"
    retention                 = 14
    storage_path              = "/opt/seerr/backups"
  }

  module {
    source = "./modules/backup_settings"
  }

  assert {
    condition     = output.supported == false || output.schedule == var.schedule
    error_message = "Backup settings endpoint is supported but the schedule did not match expected value"
  }
}

run "backup_settings_no_drift" {
  command = plan

  variables {
    backup_settings_supported = var.backup_settings_supported
    schedule                  = "0 2 * * *"
    retention                 = 14
    storage_path              = "/opt/seerr/backups"
  }

  module {
    source = "./modules/backup_settings"
  }
}
