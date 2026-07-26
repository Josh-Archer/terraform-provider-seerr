# Import all users from connected Jellyfin media server during bootstrap
resource "seerr_user_import_jellyfin" "all" {
  triggers = {
    version = "1.0"
  }
}

# Inspect imported users
output "jellyfin_imported_count" {
  value = seerr_user_import_jellyfin.all.imported_count
}
