terraform {
  required_version = ">= 1.5.0"
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = ">= 0.38.0"
    }
  }
}

provider "seerr" {
  url     = var.seerr_url
  api_key = var.seerr_api_key
}

# -----------------------------------------------------------------------------
# 1. Server Core Health & Version Telemetry
# -----------------------------------------------------------------------------

data "seerr_about" "server" {}

data "seerr_service_status" "status" {}

# -----------------------------------------------------------------------------
# 2. Workload & User Telemetry
# -----------------------------------------------------------------------------

data "seerr_users" "all" {}

data "seerr_requests" "pending" {
  filter_status = 1 # 1 = Pending Approval
}

data "seerr_requests" "all" {}

# -----------------------------------------------------------------------------
# 3. Scheduled Background Jobs & Backup Health
# -----------------------------------------------------------------------------

data "seerr_jobs" "background_jobs" {}

data "seerr_backup_settings" "backup" {}
