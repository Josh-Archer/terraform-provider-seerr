run "api_object_lifecycle" {
  command = apply

  variables {
    name     = "tofu_api_test"
    endpoint = "/api/v1/settings/main"
    payload  = jsonencode({ applicationTitle = "tofu_api_test" })
  }

  module {
    source = "./modules/api_object"
  }

  assert {
    condition     = seerr_api_object.test.name == var.name
    error_message = "API object name did not match expected value"
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



