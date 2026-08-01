data "seerr_pushbullet_settings" "example" {}

output "pushbullet_enabled" {
  value = data.seerr_pushbullet_settings.example.enabled
}

output "pushbullet_channel_tag" {
  value = data.seerr_pushbullet_settings.example.pushbullet.channel_tag
}
