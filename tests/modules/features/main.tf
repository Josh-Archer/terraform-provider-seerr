terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

resource "seerr_request" "test" {
  media_type = "movie"
  media_id   = 550 # Fight Club
}

resource "seerr_issue" "test" {
  issue_type = 4 # Other
  message    = "Test issue from tofu"
  media_id   = seerr_request.test.seerr_media_id
}

output "request_id" {
  value = seerr_request.test.id
}

output "issue_id" {
  value = seerr_issue.test.id
}
