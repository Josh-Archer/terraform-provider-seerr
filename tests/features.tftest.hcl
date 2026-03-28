run "test_features" {
  command = apply
  module {
    source = "./modules/features"
  }

  assert {
    condition     = seerr_request.test.media_id == 550
    error_message = "Request media_id did not match"
  }

  assert {
    condition     = seerr_issue.test.issue_type == 4
    error_message = "Issue type did not match"
  }
}

run "test_status_data_source" {
  command = plan
  module {
    source = "./modules/status"
  }

  assert {
    condition     = data.seerr_service_status.test.version != ""
    error_message = "Service status version is empty"
  }
}
