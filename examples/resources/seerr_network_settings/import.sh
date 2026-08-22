# In Terraform 1.5.0 and later, use an import block to import seerr_network_settings. For example:
#
# import {
#   to = seerr_network_settings.example
#   id = "network"
# }

# Otherwise, use the terraform import command:
terraform import seerr_network_settings.example network
