resource "seerr_user_notification_settings" "example" {
  user_id         = 1
  email_enabled   = true
  discord_enabled = true
  discord_id      = "123456789"

  notification_types = {
    discord = 2
    email   = 4
  }
}
