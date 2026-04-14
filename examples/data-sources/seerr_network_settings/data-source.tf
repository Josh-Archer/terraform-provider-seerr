data "seerr_network_settings" "example" {}

output "network_timeout_ms" {
  value = data.seerr_network_settings.example.api_request_timeout_ms
}
