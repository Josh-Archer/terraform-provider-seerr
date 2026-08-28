# -----------------------------------------------------------------------------
# Server Health & Version Outputs
# -----------------------------------------------------------------------------

output "server_version" {
  description = "Current running version of the Seerr server."
  value       = data.seerr_about.server.version
}

output "server_timezone" {
  description = "Configured timezone of the Seerr server."
  value       = data.seerr_about.server.tz
}

output "update_available" {
  description = "Whether a newer upstream release is available."
  value       = data.seerr_service_status.status.update_available
}

output "commits_behind" {
  description = "Number of commits the running build is behind upstream."
  value       = data.seerr_service_status.status.commits_behind
}

output "restart_required" {
  description = "Whether the server requires a restart to apply updates/settings."
  value       = data.seerr_service_status.status.restart_required
}

# -----------------------------------------------------------------------------
# Workload & Capacity Metrics
# -----------------------------------------------------------------------------

output "total_requests" {
  description = "Cumulative total requests tracked by Seerr."
  value       = data.seerr_about.server.total_requests
}

output "total_media_items" {
  description = "Cumulative total media items indexed by Seerr."
  value       = data.seerr_about.server.total_media_items
}

output "total_users" {
  description = "Total number of registered users across all auth providers."
  value       = length(data.seerr_users.all.users)
}

output "pending_requests_count" {
  description = "Number of media requests currently awaiting admin approval."
  value       = data.seerr_requests.pending.total
}

# -----------------------------------------------------------------------------
# Scheduled Jobs & Disaster Recovery State
# -----------------------------------------------------------------------------

output "backup_storage_path" {
  description = "Filesystem path where Seerr automated database backups are written."
  value       = data.seerr_backup_settings.backup.storage_path
}
output "backup_schedule" {
  description = "Cron schedule expression for automated database backups."
  value       = data.seerr_backup_settings.backup.schedule
}

output "backup_retention_days" {
  description = "Number of days database backups are retained."
  value       = data.seerr_backup_settings.backup.retention
}

output "active_jobs_count" {
  description = "Number of configured background scheduler jobs."
  value       = length(data.seerr_jobs.background_jobs.jobs)
}

# -----------------------------------------------------------------------------
# Structured Metrics Envelope (for Exporters & Dashboards)
# -----------------------------------------------------------------------------

output "metrics_summary" {
  description = "Consolidated JSON metrics summary for ingestion into monitoring stacks."
  value = {
    timestamp         = plantimestamp()
    version           = data.seerr_about.server.version
    update_available  = data.seerr_service_status.status.update_available
    commits_behind    = data.seerr_service_status.status.commits_behind
    restart_required  = data.seerr_service_status.status.restart_required
    total_requests    = data.seerr_about.server.total_requests
    total_media_items = data.seerr_about.server.total_media_items
    total_users       = length(data.seerr_users.all.users)
    pending_requests  = data.seerr_requests.pending.total
    active_jobs       = length(data.seerr_jobs.background_jobs.jobs)
    backup_retention  = data.seerr_backup_settings.backup.retention
  }
}
