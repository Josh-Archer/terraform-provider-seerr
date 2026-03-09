run "api_object_lifecycle" {
  command = apply

  variables {
    path         = "/api/v1/settings/main"
    request_body = jsonencode({ applicationTitle = "tofu_api_test" })
  }

  module {
    source = "./modules/api_object"
  }

  assert {
    condition     = seerr_api_object.test.status_code == 200
    error_message = "API object status_code was not 200"
  }
}

run "api_request_data_source" {
  command = plan

  variables {
    endpoint = "/api/v1/status"
  }

  module {
    source = "./modules/api_request_data"
  }

  assert {
    condition     = contains(keys(jsondecode(data.seerr_api_request.test.response_json)), "version")
    error_message = "API request data source response did not contain version"
  }
}



