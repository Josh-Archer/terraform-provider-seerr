# In Terraform 1.5.0 and later, use an import block to import seerr_notification_agent. For example:
#
# import {
#   to = seerr_notification_agent.example
#   id = "discord"
# }
#
# Otherwise, use the terraform import command:
terraform import seerr_notification_agent.example discord
