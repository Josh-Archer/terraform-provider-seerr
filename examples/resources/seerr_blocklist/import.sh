# In Terraform 1.5.0 and later, use an import block to import seerr_blocklist. For example:
#
# import {
#   to = seerr_blocklist.example
#   id = "1"
# }

# The TMDB ID of the blocked media.
# Otherwise, use the terraform import command:
terraform import seerr_blocklist.example 1
