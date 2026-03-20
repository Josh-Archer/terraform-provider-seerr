# In Terraform 1.5.0 and later, use an import block to import seerr_notification_pushover. For example:
#
# import {
#   to = seerr_notification_pushover.example
#   id = "pushover"
# }

# Otherwise, use the terraform import command:
terraform import seerr_notification_pushover.example pushover
