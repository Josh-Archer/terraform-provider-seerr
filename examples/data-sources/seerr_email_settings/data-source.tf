data "seerr_email_settings" "example" {}

output "email_enabled" {
  value = data.seerr_email_settings.example.enabled
}

output "smtp_host" {
  value = data.seerr_email_settings.example.email.smtp_host
}
