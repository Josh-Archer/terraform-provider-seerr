data "seerr_jellyfin_settings" "example" {}

output "jellyfin_url_base" {
  value = data.seerr_jellyfin_settings.example.url_base
}
