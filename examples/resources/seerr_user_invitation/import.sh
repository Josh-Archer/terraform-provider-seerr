# In Terraform 1.5.0 and later, use an import block to import seerr_user_invitation. For example:
#
# import {
#   to = seerr_user_invitation.example
#   id = "1"
# }

# The invitation ID.
# Otherwise, use the terraform import command:
terraform import seerr_user_invitation.example 1
