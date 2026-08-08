output "seerr_url" {
  value       = var.seerr_url
  description = "Configured URL of the Seerr instance"
}

output "admin_user_id" {
  value       = seerr_user.admin.id
  description = "Created admin user ID"
}

output "power_user_id" {
  value       = seerr_user.power_user.id
  description = "Created power user ID"
}

output "sonarr_server_id" {
  value       = seerr_sonarr_server.primary_sonarr.id
  description = "Configured Sonarr Server ID in Seerr"
}

output "radarr_server_id" {
  value       = seerr_radarr_server.primary_radarr.id
  description = "Configured Radarr Server ID in Seerr"
}
