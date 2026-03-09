data "seerr_public_settings" "example" {}

output "seerr_is_initialized" {
  value = data.seerr_public_settings.example.initialized
}
