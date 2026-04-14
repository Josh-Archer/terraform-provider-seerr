resource "seerr_backup_settings" "example" {
  schedule     = "0 0 * * *"
  retention    = 7
  storage_path = "/app/config/backups"
}
