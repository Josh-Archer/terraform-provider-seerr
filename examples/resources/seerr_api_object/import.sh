# In Terraform 1.5.0 and later, use an import block to import seerr_api_object. For example:
#
# import {
#   to = seerr_api_object.example
#   id = "GET:/api/v1/status"
# }

# Otherwise, use the terraform import command:
terraform import seerr_api_object.example GET:/api/v1/status
