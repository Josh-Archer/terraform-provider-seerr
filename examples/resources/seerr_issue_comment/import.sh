# In Terraform 1.5.0 and later, use an import block to import seerr_issue_comment. For example:
#
# import {
#   to = seerr_issue_comment.example
#   id = "1"
# }

# The comment ID.
# Otherwise, use the terraform import command:
terraform import seerr_issue_comment.example 1
