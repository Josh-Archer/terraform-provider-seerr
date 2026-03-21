# In Terraform 1.5.0 and later, use an import block to import seerr_main_settings. For example:
#
# import {
#   to = seerr_main_settings.example
#   id = "main"
# }

# Otherwise, use the terraform import command:
terraform import seerr_main_settings.example main
