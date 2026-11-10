# Adopting an Existing Instance (Secret-Safe Runbook)

Adopt `terraform-provider-seerr` on an **existing** Seerr, Jellyseerr, or Overseerr
instance without hand-writing HCL or re-creating your settings.

This is the short operator path. It ties together three things that already exist:

- [`tools/importer`](../../tools/importer) discovery and generation (including
  secret-reference output and discovery pagination).
- The full [`Migration & Bulk Adoption Guide`](migration.md) for flag reference
  and resources the importer intentionally leaves unmanaged.
- The provider's import support (`import` blocks or `terraform import`).

For background on what the importer covers, start here. For the exhaustive flag
table and the list of settings left for manual adoption, see
[`migration.md`](migration.md).

> [!CAUTION]
> Validate this runbook against a **disposable or staging** instance first
> (for example a local container you can throw away). Do not run adoption
> against your production home Seerr until the disposable run is clean.

## Prerequisites

- Go toolchain (to run the importer) and OpenTofu (`tofu`) or Terraform
  (`terraform`) 1.5+.
- Network access to the disposable Seerr instance and an administrator API key.
- An empty output directory (or pass `-overwrite` / `-force`, see below).

## Step 1: Point the importer at the disposable instance

Export credentials as environment variables. Do **not** pass `-api-key` on the
command line: shell history and the process list keep command arguments.

```bash
export SEERR_URL="http://localhost:5055"
export SEERR_API_KEY="YOUR_SEERR_API_KEY"

go run ./tools/importer --out-dir "./seerr-adopt"
```

Current flags (see `go run ./tools/importer -h` for the source of truth):

| Flag | Default | Meaning |
| :--- | :--- | :--- |
| `--url` | `$SEERR_URL` or `http://localhost:5055` | Base URL of the instance |
| `--api-key` | `$SEERR_API_KEY` | Prefer the env var (see above) |
| `--out-dir` | `.` | Output directory (created `0700`) |
| `--format` | `all` | `all`, `hcl`, `imports`, or `script` |
| `--provider-header` | `true` | Include `terraform {}` + `provider "seerr" {}` in `main.tf` |
| `--timeout` | `30` | HTTP timeout in seconds |
| `--allow-partial` | `false` | Write discovered-so-far on incomplete discovery (see below) |
| `--overwrite` / `--force` | `false` | Replace existing generated files (otherwise the run refuses and exits non-zero) |

Expect a discovery summary on stdout:

```text
🔍 Connecting to Seerr at http://localhost:5055...
✅ Discovered 14 live resources across your Seerr instance!

  • seerr_main_settings            : 1
  • seerr_radarr_server            : 2
  ...
📄 Wrote seerr-adopt/main.tf
📄 Wrote seerr-adopt/variables.tf
📄 Wrote seerr-adopt/imports.tf
📄 Wrote seerr-adopt/import.sh
```

## Step 2: Confirm the generated HCL is secret-safe

Fetched credentials are **never written as literals**. `main.tf` references
variables; `variables.tf` declares them with no defaults and no values:

```hcl
# main.tf (generated; attributes are column-aligned)
resource "seerr_radarr_server" "hd" {
  api_key                   = var.seerr_radarr_server_hd_api_key
  hostname                  = "radarr.local"
  port                      = 7878
}
```

```hcl
# variables.tf (generated)
variable "seerr_radarr_server_hd_api_key" {
  type        = string
  sensitive   = true
  description = "Sensitive value for seerr_radarr_server.hd api_key. Set TF_VAR_seerr_radarr_server_hd_api_key or an uncommitted tfvars file. State stores this value in plaintext; use an encrypted backend."
}
```

Check before you proceed:

```bash
# From the repo root, after generating into ./seerr-adopt:
grep -rEi 'api[_-]?key\s*=\s*"[^"]+"|webhook[_-]?url\s*=\s*"https?://|bot[_-]?api\s*=\s*"[^"]+"|access[_-]?token\s*=\s*"[^"]+"' seerr-adopt/main.tf seerr-adopt/variables.tf \
  && echo "FAIL: possible credential literal in generated files" \
  || echo "OK: no credential literals found"
```

If that grep matches, **stop**: do not commit or apply. File a bug with the
resource type and attribute (redact the value). The automated stand-in for this
check is `TestOutputPermissionsAndOverwriteControl` in
`tools/importer/importer_output_test.go`, which fails if a fetched credential
reaches any generated file or CLI output.

Generated files are owner-only (`0700` for the directory and `import.sh`,
`0600` for `.tf` files). Even so, treat them as sensitive: the `var.*`
references are safe to keep, but the live values you supply next are not.

## Step 3: Wire the secret variables (without committing them)

Supply every variable in `variables.tf` at plan time. Pick one:

```bash
# Option A: per-process environment (preferred for one-off adoption)
export TF_VAR_seerr_radarr_server_hd_api_key="live-value-here"

# Option B: an uncommitted tfvars file
cat > seerr-adopt/terraform.tfvars <<'EOF'
seerr_radarr_server_hd_api_key = "live-value-here"
EOF
```

> [!WARNING]
> Never commit `terraform.tfvars`, `*.tfvars`, state files, or any file holding
> live secret values — even if generation looks correct today. A future
> regression could write a literal where a `var.*` reference belongs, and a
> committed tfvars file would still leak. Keep these paths git-ignored and
> local-only. Terraform/OpenTofu also store sensitive values in state in
> plaintext: use an encrypted remote backend (and OpenTofu state encryption
> where available). See [`migration.md`](migration.md) for backend examples.

## Step 4: Import, then expect a clean re-plan

```bash
cd seerr-adopt
tofu init        # or: terraform init
tofu apply       # imports via imports.tf, then manages live settings
tofu plan        # must be clean (or: terraform plan / terraform apply)
```

Clean re-plan expectation:

```text
No changes. Your infrastructure matches the configuration.

Tofu has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.
```

i.e. **0 to add, 0 to change, 0 to destroy**. If the plan proposes changes,
the generated HCL does not match the live instance — do not apply. Re-check
secret wiring first (an empty variable is the most common cause), then the
failure modes below.

Prefer `import` blocks (`imports.tf`, Terraform 1.5+ / OpenTofu) over the
legacy `import.sh` script. After a successful import, delete `imports.tf` (or
keep it if you use declarative imports intentionally).

## Failure modes you may hit (already handled — just react)

| What you see | What it means | What to do |
| :--- | :--- | :--- |
| `incomplete discovery: ... re-run with -allow-partial` and nothing written | Auth (401/403), transport, or server error aborted discovery | Fix URL/key/network, re-run. Only add `-allow-partial` as a deliberate opt-in to keep discovered-so-far. |
| `Partial discovery: N error(s) ignored because -allow-partial is set` | Output covers only what was reachable | Treat the result as incomplete: diff against the instance, fix the listed failures, re-run without the flag. |
| `Skipped N unsupported endpoints (HTTP 404)` | That endpoint is not on this server build | Normal across Seerr / Jellyseerr / Overseerr versions. Nothing to do. |
| `Manual adoption required for settings left unmanaged:` | Populated network / discover / override settings were seen but not emitted (emitting a partial resource would null live settings) | Adopt those by hand per [`migration.md`](migration.md#settings-the-importer-does-not-adopt). Leaving them out keeps live values in place. |
| `radarr_hd_2`-style names | Two names sanitized to the same identifier; lowest import ID keeps the base name, rest get `_N` suffixes (stable for the same IDs) | Keep the generated names; update your references to match. |
| `refusing to overwrite existing file ... (use -overwrite to replace it)` | Your `out-dir` already has generated files | Re-run into an empty directory, or pass `-overwrite` / `-force` deliberately. |
| `invalid output format ...` before any connection attempt | Bad `--format` value | Use `all`, `hcl`, `imports`, or `script`. |
| Users stop paginating early | User listing is read `take`-pages at a time until a short page, a reported total, a repeated page, or the page cap | Re-run; if it persists, the run is incomplete — do not treat it as full adoption. |

## Validating without a live server

No disposable instance handy? The importer integration tests are the automated
stand-in for this runbook:

```bash
go test ./tools/importer/...
```

- `TestImporterCLIOutputValidatesWithLocalProvider` runs the real CLI against a
  fixture server and validates the generated HCL with the local provider.
- `TestOutputPermissionsAndOverwriteControl` proves fetched credentials reach
  neither generated files nor CLI output, and covers the overwrite-refusal flow.
- `TestCLIAPIErrorPathDoesNotLeakCredentials` proves error payloads (which can
  echo credentials) are omitted from CLI output.
- `TestSensitiveAttributeMappingMatchesProviderSchema` keeps the
  secret-reference classification in sync with provider schemas.

If you changed provider schemas or importer behavior, those tests — not a
production instance — are where you re-prove adoption.
