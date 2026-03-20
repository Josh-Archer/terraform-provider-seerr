# In Terraform 1.5.0 and later, use an import block to import seerr_job_schedule. For example:
#
# import {
#   to = seerr_job_schedule.example
#   id = "plex-sync"
# }

# The ID of the job, for example: `plex-sync`, `radarr-sync`, etc.
# Otherwise, use the terraform import command:
terraform import seerr_job_schedule.example plex-sync
