# User can be imported by ID
terraform import seerr_user.example 1

# Additionally, the user can be imported by username
terraform import seerr_user.example jdoe

# Or the user can be imported by their email address
terraform import seerr_user.example jdoe@example.com
