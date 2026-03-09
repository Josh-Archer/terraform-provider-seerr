resource "seerr_user" "example" {
  username    = "jdoe"
  email       = "jdoe@example.com"
  permissions = 0
  locale      = "en"

  notification_settings {
    email_enabled = true
    
    notification_types {
      email = 1
    }
  }
}
