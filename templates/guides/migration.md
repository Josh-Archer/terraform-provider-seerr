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

Once imported, delete `imports.tf` (or keep it if using Terraform 1.5+ declarative imports), and commit your `main.tf` to your Git repository.

Your entire Seerr stack is now fully managed as Infrastructure as Code! 🎉

---

## 🛠️ CLI Options Reference

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--url` | `$SEERR_URL` or `http://localhost:5055` | Base URL of your live Seerr server |
| `--api-key` | `$SEERR_API_KEY` | API Key with administrator permissions |
| `--out-dir` | `.` | Directory to write generated `.tf` files |
| `--format` | `all` | Output format: `all`, `hcl`, `imports`, or `script` |
| `--provider-header` | `true` | Include `terraform {}` and `provider "seerr" {}` block in `main.tf` |
| `--timeout` | `30` | HTTP timeout in seconds for API queries |

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
