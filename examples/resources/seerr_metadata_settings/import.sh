# In Terraform 1.5.0 and later, use an import block to import seerr_metadata_settings. For example:
#
# import {
#   to = seerr_metadata_settings.example
#   id = "metadata"
# }

# Otherwise, use the terraform import command:
terraform import seerr_metadata_settings.example metadata
