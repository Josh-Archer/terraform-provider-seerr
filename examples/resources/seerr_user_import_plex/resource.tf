# Import all users from connected Plex media server during bootstrap
resource "seerr_user_import_plex" "all" {
  triggers = {
    version = "1.0"
  }
}

# Inspect imported users
output "plex_imported_count" {
  value = seerr_user_import_plex.all.imported_count
}
