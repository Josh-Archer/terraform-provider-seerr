run "override_rule_lifecycle" {
  command = apply

  variables {
    genre    = "18"
    language = "en"
    keywords = "heist"
  }

  module {
    source = "./modules/override_rule"
  }

  assert {
    condition     = seerr_override_rule.test.genre == var.genre
    error_message = "Override rule genre did not match expected value"
  }

  assert {
    condition     = seerr_override_rule.test.language == var.language
    error_message = "Override rule language did not match expected value"
  }

  assert {
    condition     = seerr_override_rule.test.keywords == var.keywords
    error_message = "Override rule keywords did not match expected value"
  }
}

run "override_rule_no_drift" {
  command = plan
}

run "override_rule_update" {
  command = apply

  variables {
    genre    = "28"
    language = "fr"
    keywords = "action"
  }

  module {
    source = "./modules/override_rule"
  }

  assert {
    condition     = seerr_override_rule.test.genre == "28"
    error_message = "Override rule genre update failed"
  }

  assert {
    condition     = seerr_override_rule.test.language == "fr"
    error_message = "Override rule language update failed"
  }

  assert {
    condition     = seerr_override_rule.test.keywords == "action"
    error_message = "Override rule keywords update failed"
  }
}
