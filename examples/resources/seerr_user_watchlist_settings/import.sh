# In Terraform 1.5.0 and later, use an import block to import seerr_user_watchlist_settings. For example:
#
# import {
#   to = seerr_user_watchlist_settings.example
#   id = "1"
# }

# The ID of the user.
# Otherwise, use the terraform import command:
terraform import seerr_user_watchlist_settings.example 1
