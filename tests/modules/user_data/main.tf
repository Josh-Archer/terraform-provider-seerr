variable "email" { type = string }

data "seerr_user" "test" {
  email = var.email
}

output "id" {
  value = data.seerr_user.test.id
}
