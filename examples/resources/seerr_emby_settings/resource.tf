resource "seerr_emby_settings" "default" {
  name    = "Emby"
  ip      = "192.168.1.100"
  port    = 8096
  use_ssl = false
  api_key = "YOUR_EMBY_API_KEY"
}
