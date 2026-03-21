# In Terraform 1.5.0 and later, use an import block to import seerr_notification_webhook. For example:
#
# import {
#   to = seerr_notification_webhook.example
#   id = "webhook"
# }

# Otherwise, use the terraform import command:
terraform import seerr_notification_webhook.example webhook
