# In Terraform 1.5.0 and later, use an import block to import seerr_watchlist. For example:
#
# import {
#   to = seerr_watchlist.example
#   id = "1"
# }

# The watchlist item ID.
# Otherwise, use the terraform import command:
terraform import seerr_watchlist.example 1
