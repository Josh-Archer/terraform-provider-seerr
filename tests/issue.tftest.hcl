run "issue_lifecycle" {
  command = apply

  variables {
    issue_type = 4
    message    = "Audio track desync"
    media_id   = 1
    status     = 1
  }

  module {
    source = "./modules/issue"
  }

  assert {
    condition     = seerr_issue.test.issue_type == var.issue_type
    error_message = "Issue type did not match expected value"
  }

  assert {
    condition     = seerr_issue.test.message == var.message
    error_message = "Issue message did not match expected value"
  }

  assert {
    condition     = seerr_issue.test.media_id == var.media_id
    error_message = "Issue media_id did not match expected value"
  }

  assert {
    condition     = seerr_issue.test.status == var.status
    error_message = "Issue status did not match expected value"
  }
}

run "issue_no_drift" {
  command = plan
}

run "issue_update_status" {
  command = apply

  variables {
    issue_type = 4
    message    = "Audio track desync"
    media_id   = 1
    status     = 2
  }

  module {
    source = "./modules/issue"
  }

  assert {
    condition     = seerr_issue.test.status == 2
    error_message = "Issue status update to resolved failed"
  }
}
