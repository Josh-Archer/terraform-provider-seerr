data "seerr_user_notification_settings" "example" {
  user_id = 1
}

output "user_email_notifications_enabled" {
  value = data.seerr_user_notification_settings.example.email_enabled
}
