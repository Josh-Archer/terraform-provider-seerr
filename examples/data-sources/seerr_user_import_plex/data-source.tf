# Read existing Plex users imported into Seerr
data "seerr_user_import_plex" "all" {}

output "imported_plex_users" {
  value = data.seerr_user_import_plex.all.imported_users
}
