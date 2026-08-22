# In Terraform 1.5.0 and later, use an import block to import seerr_api_key. For example:
#
# import {
#   to = seerr_api_key.example
#   id = "api_key"
# }

# Otherwise, use the terraform import command:
terraform import seerr_api_key.example api_key
