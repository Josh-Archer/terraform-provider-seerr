resource "seerr_jellyfin_settings" "default" {
  name    = "Jellyfin"
  ip      = "192.168.1.100"
  port    = 8096
  use_ssl = false
  api_key = "YOUR_JELLYFIN_API_KEY"
}
