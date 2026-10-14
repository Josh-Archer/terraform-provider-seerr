resource "seerr_radarr_server_test" "example" {
  hostname = "127.0.0.1"
  port     = 7878
  api_key  = "your-radarr-api-key"
  use_ssl  = false
}
