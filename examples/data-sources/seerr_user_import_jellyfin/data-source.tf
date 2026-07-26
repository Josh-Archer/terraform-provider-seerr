# Read existing Jellyfin users imported into Seerr
data "seerr_user_import_jellyfin" "all" {}

output "imported_jellyfin_users" {
  value = data.seerr_user_import_jellyfin.all.imported_users
}
