data "seerr_pushover_sounds" "example" {
  token = "your-pushover-token"
}

output "pushover_sounds" {
  value = data.seerr_pushover_sounds.example.sounds
}
