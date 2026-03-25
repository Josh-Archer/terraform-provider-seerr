# In Terraform 1.5.0 and later, use an import block to import seerr_sonarr_server. For example:
#
# import {
#   to = seerr_sonarr_server.example
#   id = "0"
# }

# The ID of the server. For the first server, the ID is `0`.
# Otherwise, use the terraform import command:
terraform import seerr_sonarr_server.example 0
