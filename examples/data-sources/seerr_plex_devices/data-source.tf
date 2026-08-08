data "seerr_plex_devices" "all" {}

output "plex_servers" {
  value = data.seerr_plex_devices.all.devices
}
