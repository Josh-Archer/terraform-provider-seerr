# In Terraform 1.5.0 and later, use an import block to import seerr_jellyfin_library_settings. For example:
#
# import {
#   to = seerr_jellyfin_library_settings.example
#   id = "jellyfin_library_settings"
# }

# Otherwise, use the terraform import command:
terraform import seerr_jellyfin_library_settings.example jellyfin_library_settings
