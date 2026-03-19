# In Terraform 1.5.0 and later, use an import block to import seerr_user. For example:
#
# import {
#   to = seerr_user.example
#   id = "1

# Additionally, you can import by username or email:
# terraform import seerr_user.example jdoe
# terraform import seerr_user.example jdoe@example.com"
# }
#
# Otherwise, use the terraform import command:
terraform import seerr_user.example 1

# Additionally, you can import by username or email:
# terraform import seerr_user.example jdoe
# terraform import seerr_user.example jdoe@example.com
