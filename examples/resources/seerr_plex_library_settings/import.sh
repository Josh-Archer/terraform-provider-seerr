# In Terraform 1.5.0 and later, use an import block to import seerr_plex_library_settings. For example:
#
# import {
#   to = seerr_plex_library_settings.example
#   id = "plex_library_settings"
# }

# Otherwise, use the terraform import command:
terraform import seerr_plex_library_settings.example plex_library_settings
