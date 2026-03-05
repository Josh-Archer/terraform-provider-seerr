resource "seerr_api_object" "main_settings" {
  path              = "/api/v1/settings/main"
  read_method       = "GET"
  create_method     = "POST"
  update_method     = "POST"
  delete_method     = "POST"
  suppress_not_found = false

  request_body_json = jsonencode({
    applicationTitle = "Seerr"
    locale           = "en"
  })
}

