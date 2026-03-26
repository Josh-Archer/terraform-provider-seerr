run "backup_settings_lifecycle" {
  command = plan

  variables {
    schedule     = "0 2 * * *"
    retention    = 14
    storage_path = "/opt/seerr/backups"
  }

  module {
    source = "./modules/backup_settings"
  }

  assert {
    condition     = seerr_backup_settings.test.schedule == var.schedule
    error_message = "Backup settings schedule did not match expected value"
  }
}

