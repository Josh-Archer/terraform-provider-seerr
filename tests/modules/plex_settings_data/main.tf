data "seerr_plex_settings" "test" {}

output "ip" {
  value = data.seerr_plex_settings.test.ip
}
