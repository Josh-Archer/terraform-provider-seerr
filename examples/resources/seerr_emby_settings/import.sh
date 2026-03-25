# In Terraform 1.5.0 and later, use an import block to import seerr_emby_settings. For example:
#
# import {
#   to = seerr_emby_settings.example
#   id = "emby"
# }

# Otherwise, use the terraform import command:
terraform import seerr_emby_settings.example emby
