data "seerr_permission_set" "power_user" {
  request       = true
  request_movie = true
  request_tv    = true
  request_4k    = true
  auto_approve  = true
}

resource "seerr_user_permissions" "example" {
  user_id     = 123
  permissions = data.seerr_permission_set.power_user.permissions
}
