# In Terraform 1.5.0 and later, use an import block to import seerr_backup_settings. For example:
#
# import {
#   to = seerr_backup_settings.example
#   id = "backup"
# }

# Otherwise, use the terraform import command:
terraform import seerr_backup_settings.example backup
