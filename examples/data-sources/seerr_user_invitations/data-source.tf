# Deprecated and unsupported: Seerr does not expose an invitation API.
# Use seerr_user to create local users instead.
resource "seerr_user" "example" {
  username = "jdoe"
  email    = "jdoe@example.com"
}
