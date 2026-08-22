# In Terraform 1.5.0 and later, use an import block to import seerr_emby_library_settings. For example:
#
# import {
#   to = seerr_emby_library_settings.example
#   id = "1"
# }

# The library ID.
# Otherwise, use the terraform import command:
terraform import seerr_emby_library_settings.example 1
