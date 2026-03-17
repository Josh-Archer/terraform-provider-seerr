data "seerr_job_schedule" "plex_sync" {
  job_id = "plex-sync"
}

output "plex_sync_schedule" {
  value = data.seerr_job_schedule.plex_sync.schedule
}
