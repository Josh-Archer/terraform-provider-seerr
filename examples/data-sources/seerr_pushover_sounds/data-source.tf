data "seerr_pushover_sounds" "available" {}

output "pushover_sounds" {
  value = data.seerr_pushover_sounds.available.sounds
}
