# Migration & Bulk Adoption Guide

This guide walks you through adopting `terraform-provider-seerr` on an **existing, live** Seerr, Jellyseerr, or Overseerr instance without having to write Terraform/OpenTofu HCL by hand or re-create your settings.

---

## ⚡ Overview

With the automated **Bulk Importer CLI (`tools/importer`)**, you can inspect your live server and generate:
1. **Idiomatic Terraform/OpenTofu HCL (`main.tf`)** matching your existing configuration.
2. **Modern `import { ... }` blocks (`imports.tf`)** compatible with OpenTofu and Terraform 1.5+.
3. **Legacy import commands (`import.sh`)** for scriptable CLI workflows.

---

## 🚀 Step-by-Step Walkthrough

### Step 1: Obtain Your Seerr API Key & URL

1. Log in to your Seerr/Jellyseerr web interface as an Administrator.
2. Navigate to **Settings** ➔ **General**.
3. Copy your **API Key** and note your server URL (e.g. `http://192.168.1.100:5055` or `https://seerr.yourdomain.com`).

---

### Step 2: Run the Bulk Importer CLI

Clone the repository and run the importer pointing at your live server:

```bash
# Clone the provider repository
git clone https://github.com/Josh-Archer/terraform-provider-seerr.git
cd terraform-provider-seerr

# Run the bulk importer CLI
go run ./tools/importer \
  --url "http://192.168.1.100:5055" \
  --api-key "YOUR_SEERR_API_KEY" \
  --out-dir "./my-seerr-tf"
```

Or configure environment variables:

```bash
export SEERR_URL="http://192.168.1.100:5055"
export SEERR_API_KEY="YOUR_SEERR_API_KEY"

go run ./tools/importer --out-dir "./my-seerr-tf"
```

The CLI will connect to your instance, discover all configured resources, and output summary metrics:

```text
🔍 Connecting to Seerr at http://192.168.1.100:5055...
✅ Discovered 14 live resources across your Seerr instance!

  • seerr_main_settings            : 1
  • seerr_plex_settings            : 1
  • seerr_radarr_server            : 2
  • seerr_sonarr_server            : 2
  • seerr_notification_discord     : 1
  • seerr_notification_email       : 1
  • seerr_override_rule            : 2
  • seerr_user                     : 5

📄 Generated HCL resources: my-seerr-tf/main.tf
📦 Generated Import blocks:  my-seerr-tf/imports.tf
🚀 Generated Import script:  my-seerr-tf/import.sh
```

---

### Step 3: Initialize OpenTofu or Terraform

Switch into your newly generated configuration directory:

```bash
cd my-seerr-tf
```

#### OpenTofu
```bash
tofu init
```

#### Terraform
```bash
terraform init
```

---

### Step 4: Verify Clean Plan (Zero Drift)

Run a plan to verify that the generated configuration accurately matches your live server:

```bash
tofu plan
# or
terraform plan
```

The plan will indicate that the resources will be imported into state with **0 to add, 0 to change, 0 to destroy**.

---

### Step 5: Apply & Lock In State

Apply the configuration to import all live resources into your state file:

```bash
tofu apply
# or
terraform apply
```

Before the first plan, supply sensitive variables. The importer writes `variables.tf` and references `var.<name>` from `main.tf`. It does not write API keys or bot tokens into generated files. Put the live values in an uncommitted `terraform.tfvars` file, or export `TF_VAR_<name>`.

Once imported, run plan again. A clean plan (`0 to add, 0 to change, 0 to destroy`) means Terraform will not update the instance. Delete `imports.tf` after the import (or keep it if you use Terraform 1.5+ declarative imports). Do not commit tfvars files that contain credentials.

Settings that are not in the generated configuration stay on the server, because the importer never sends updates and a clean plan does not apply changes.

---

## Sensitive values and resource addresses

Fetched credentials become sensitive variables with no default. Terraform and OpenTofu still store those values in state. Use an encrypted remote backend, or OpenTofu state encryption.

Resource addresses are sanitized. When two names sanitize to the same identifier, the lower import ID keeps the base name and the rest get a numeric suffix (`radarr_hd`, `radarr_hd_2`). The suffix is stable for the same import IDs.

Users are read with `take=100` until a short page, a reported total, a repeated page of IDs, or 1000 pages. Re-running discovery does not invent a second resource for an ID that was already seen.

`--format` must be `all`, `hcl`, `imports`, or `script`. Any other value exits non-zero before the importer contacts Seerr. Write failures, including an output path that is not a directory, also exit non-zero.

HTTP 404 means that endpoint is not on this server and is skipped. Authentication failures (401/403), transport failures, and other HTTP errors abort discovery and write nothing. `--allow-partial` is the explicit opt-in that still writes the resources discovered before those errors.

---

## Settings the importer does not adopt

Adopt these by hand. Each resource's import syntax is under [`docs/resources`](../resources). Leaving them out of the generated configuration is what keeps their live values in place:

- `seerr_metadata_settings`
- `seerr_network_settings` (proxy, DNS cache, and timeouts). A partial network resource can null live proxy settings, so a populated network endpoint is reported for manual adoption instead of emitted.
- `seerr_discover_slider`
- `seerr_override_rule`
- `seerr_backup_settings`
- `seerr_job_schedule` and `seerr_job_run`
- `seerr_blocklist`
- `seerr_issue` and `seerr_issue_comment`
- `seerr_request`, `seerr_request_approval`, and `seerr_request_retry`
- `seerr_api_key`
- Plex, Jellyfin, and Emby library settings and library sync
- `seerr_user_quota`, `seerr_user_permissions`, `seerr_user_notification_settings`, `seerr_user_watchlist_settings`, and `seerr_user_invitation`
- `seerr_watchlist`
- Notification and Arr connection test actions

Deprecated main-setting aliases `region`, `image_proxy`, and `plex_login` are not emitted. They conflict with `streaming_region` / `discover_region`, `cache_images`, and `media_server_login`. If a later plan shows one of those aliases drifting, set the current attribute and leave the deprecated one out of the configuration.

---

## 🛠️ CLI Options Reference

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--url` | `$SEERR_URL` or `http://localhost:5055` | Base URL of your live Seerr server |
| `--api-key` | `$SEERR_API_KEY` | API Key with administrator permissions. Prefer `SEERR_API_KEY` so the key is not stored in shell history. |
| `--out-dir` | `.` | Directory to write generated `.tf` files |
| `--format` | `all` | Output format: `all`, `hcl`, `imports`, or `script`. Any other value exits non-zero before discovery. |
| `--provider-header` | `true` | Include `terraform {}` and `provider "seerr" {}` block in `main.tf` |
| `--timeout` | `30` | HTTP timeout in seconds for API queries |
| `--allow-partial` | `false` | Write partial HCL when authentication, transport, or server errors make discovery incomplete |

---

## 📌 Manual Resource Import Reference

If you prefer to import individual resources manually instead of bulk importing, every resource supports standard `import { ... }` blocks and `terraform import` commands:

### Example: Import Settings
```hcl
import {
  to = seerr_main_settings.main
  id = "main"
}

import {
  to = seerr_plex_settings.plex
  id = "plex"
}
```

### Example: Import Radarr / Sonarr Server
```hcl
# The ID is the 0-indexed server ID from Seerr
import {
  to = seerr_radarr_server.hd
  id = "0"
}
```

### Example: Import User
```hcl
# Import by numeric ID, email, or username
import {
  to = seerr_user.admin
  id = "1"
}
```

See individual resource documentation under [`docs/resources`](../resources) for exact import IDs and syntax.
