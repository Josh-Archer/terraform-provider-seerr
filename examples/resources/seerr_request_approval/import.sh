# In Terraform 1.5.0 and later, use an import block to import seerr_request_approval. For example:
#
# import {
#   to = seerr_request_approval.example
#   id = "1"
# }

# The request ID to approve/decline.
# Otherwise, use the terraform import command:
terraform import seerr_request_approval.example 1
