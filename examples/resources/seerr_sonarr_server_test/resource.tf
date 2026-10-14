resource "seerr_sonarr_server_test" "example" {
  hostname = "127.0.0.1"
  port     = 8989
  api_key  = "your-sonarr-api-key"
  use_ssl  = false
}
