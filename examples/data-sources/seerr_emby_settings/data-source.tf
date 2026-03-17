data "seerr_emby_settings" "default" {}

output "emby_url_base" {
  value = data.seerr_emby_settings.default.url_base
}
