# In Terraform 1.5.0 and later, use an import block to import seerr_emby_library_settings. For example:
#
# import {
#   to = seerr_emby_library_settings.example
#   id = "emby_library_settings"
# }

# Otherwise, use the terraform import command:
terraform import seerr_emby_library_settings.example emby_library_settings
