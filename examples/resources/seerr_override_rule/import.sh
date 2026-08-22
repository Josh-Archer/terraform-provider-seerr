# In Terraform 1.5.0 and later, use an import block to import seerr_override_rule. For example:
#
# import {
#   to = seerr_override_rule.example
#   id = "1"
# }

# The rule ID.
# Otherwise, use the terraform import command:
terraform import seerr_override_rule.example 1
