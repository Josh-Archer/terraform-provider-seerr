output "seerr_application_url" {
  value       = seerr_main_settings.main.application_url
  description = "Configured public application URL for Seerr."
}

output "radarr_server_id" {
  value       = try(seerr_radarr_server.radarr[0].id, null)
  description = "ID of the registered Radarr server."
}

output "sonarr_server_id" {
  value       = try(seerr_sonarr_server.sonarr[0].id, null)
  description = "ID of the registered Sonarr server."
}

output "discovered_radarr_root_folders" {
  value       = try(data.seerr_radarr_root_folders.movies[0].root_folders, [])
  description = "Discovered root folders on the Radarr instance."
}

output "discovered_sonarr_root_folders" {
  value       = try(data.seerr_sonarr_root_folders.shows[0].root_folders, [])
  description = "Discovered root folders on the Sonarr instance."
}
