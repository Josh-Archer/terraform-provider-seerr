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

Clone the repository and run the importer pointing at your live server.

**Recommended:** Set credentials via environment variables to prevent exposing your API key in shell history and process tables (`ps aux`):

```bash
# Clone the provider repository
git clone https://github.com/Josh-Archer/terraform-provider-seerr.git
cd terraform-provider-seerr

# Configure environment variables (preferred for security)
export SEERR_URL="http://192.168.1.100:5055"
export SEERR_API_KEY="YOUR_SEERR_API_KEY"

# Run the bulk importer CLI
go run ./tools/importer --out-dir "./my-seerr-tf"
```

Command-line flags are also supported:

```bash
go run ./tools/importer \
  --url "http://192.168.1.100:5055" \
  --api-key "YOUR_SEERR_API_KEY" \
  --out-dir "./my-seerr-tf"
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
🔒 Generated Sensitive variables: my-seerr-tf/variables.tf
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
| `--api-key` | `$SEERR_API_KEY` | API Key (setting `$SEERR_API_KEY` environment variable is preferred over command-line arguments) |
| `--out-dir` | `.` | Directory to write generated `.tf` files (created with restrictive mode `0700`) |
| `--format` | `all` | Output format: `all`, `hcl`, `variables`, `imports`, or `script` |
| `--provider-header` | `true` | Include `terraform {}` and `provider "seerr" {}` block in `main.tf` |
| `--timeout` | `30` | HTTP timeout in seconds for API queries |
| `--overwrite` / `--force` | `false` | Overwrite existing files in `out-dir` (refuses overwrite by default to prevent accidental data loss) |

> [!NOTE]
> All generated configuration files are written with restrictive owner-only permissions: mode `0600` for `.tf` files, mode `0700` for `import.sh`, and mode `0700` for created directories.

---

## 🔒 Security Best Practices: State Storage & Credentials

### Secret Handling in Generated Configuration
When importing live instances, the importer automatically classifies credentials and sensitive settings (such as notification passwords, API keys, tokens, and PGP private keys). Rather than embedding plaintext secrets into your HCL files:
- Credentials are replaced with variable references (`var.<name>`).
- Sensitive variable declarations are written to `variables.tf` with `sensitive = true`.
- Fetched credential values are never written to HCL, scripts, or console diagnostics.

### State File Security
While variable references protect your source code repositories from plaintext secrets, **Terraform and OpenTofu persist all resource attribute values (including sensitive variables) in plaintext within the state file (`.tfstate`) when applied**.

To keep your infrastructure and credentials secure:
1. **Use Encrypted Remote State**: Store your `.tfstate` in an encrypted remote backend with access controls:
   - **AWS S3**: Enable server-side encryption with AWS KMS (`sse_algorithm = "aws:kms"`) and enforce TLS via bucket policies.
   - **OpenTofu State Encryption**: OpenTofu 1.7+ supports native state encryption at rest (AES-GCM or PBKDF2).
   - **Terraform Cloud / HCP Terraform**: Encrypts state at rest and in transit automatically.
   - **Google Cloud Storage (GCS)** or **Azure Blob Storage**: Enable customer-managed encryption keys.
2. **Restrict Backend Permissions**: Limit access to the state storage bucket or backend strictly to authorized CI/CD deployment pipelines and administrators.
3. **Never Commit State Files**: Ensure `.tfstate` and `.tfstate.backup` are added to your `.gitignore` file.
4. **Pass Credentials Securely**: Provide sensitive variable values via `TF_VAR_<name>` environment variables or your CI/CD secret manager rather than committing `.tfvars` files to git.

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
