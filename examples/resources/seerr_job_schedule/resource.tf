resource "seerr_job_schedule" "example" {
  job_id   = "plex-sync"
  schedule = "0 0 * * *"
}
