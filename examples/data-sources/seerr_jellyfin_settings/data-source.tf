data "seerr_jellyfin_settings" "default" {}

output "jellyfin_url_base" {
  value = data.seerr_jellyfin_settings.default.url_base
}
