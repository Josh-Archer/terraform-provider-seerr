data "seerr_jobs" "example" {}

output "job_count" {
  value = length(data.seerr_jobs.example.jobs)
}
