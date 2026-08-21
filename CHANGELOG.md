# Changelog

All notable changes to `terraform-provider-seerr` are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.38.1](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.38.0...v0.38.1) (2026-08-20)


### Features

* **ci:** release automation with release-please, tag-on-merge, and registry verification ([#181](https://github.com/Josh-Archer/terraform-provider-seerr/issues/181)) ([#187](https://github.com/Josh-Archer/terraform-provider-seerr/issues/187)) ([34fb8c6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/34fb8c6e89d7d90f1945bbaa9815c925dbaf4cb3))


### Bug Fixes

* **ci:** correct release-please-action commit sha in release-please.yml ([#188](https://github.com/Josh-Archer/terraform-provider-seerr/issues/188)) ([03e1e35](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e1e354dd87e7d575553e9861c46cfda8d0497c))


### Documentation

* modernize README and docs for accuracy and conciseness ([#192](https://github.com/Josh-Archer/terraform-provider-seerr/issues/192)) ([8b4cc9b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b4cc9b7c3b8a0e4c0b2aaa001bd96b3f6a195a7))

## 0.38.0 (Unreleased)

### Features
- **Production Readiness Suite**: Added comprehensive `.tftest.hcl` OpenTofu test suites for `seerr_issue`, `seerr_request`, `seerr_override_rule`, and `seerr_blocklist` with full CRUD, no-drift idempotency, and status updates (#165).
- **Composite Example Module**: Added `examples/modules/full-media-stack` showcasing complete integration of Seerr with Plex/Jellyfin, Radarr, Sonarr, override rules, discover sliders, and Discord/Ntfy notification channels.

### Improvements
- **Mock Media Service**: Added endpoints for root folder, tag, and language profile mocking in test sidecars.
- **Integration Test Coverage**: Expanded `stable_filters` in CI test runner to validate lifecycle suites across all core and operations resources.

---

## 0.37.1 (2026-08-20)

### Bug Fixes
- **Retry Disabling**: Fixed `NewClient` check from `maxRetries <= 0` to `maxRetries < 0` so that `maxRetries = 0` (or `SEERR_MAX_RETRIES=0`) correctly executes exactly 1 attempt without retries.
- **Empty Collection Null-Safety**: Initialized slices with `make([]..., 0, len(results))` across all five Arr data sources (`seerr_sonarr_root_folders`, `seerr_radarr_root_folders`, `seerr_sonarr_tags`, `seerr_radarr_tags`, `seerr_sonarr_language_profiles`) to return empty lists rather than null state when no items exist upstream.
- **Synthetic ID Defensive Checks**: Added `!IsUnknown()` guards in `seerr_users` and `seerr_requests` data source ID generation to prevent plan-time evaluation panics.

---

## 0.37.0 (2026-08-19)

### Features
- **Environment Variable Fallbacks**: Added automatic fallback support for `SEERR_URL`, `SEERR_API_KEY`, and `SEERR_INSECURE_SKIP_VERIFY` (#164).
- **Configurable HTTP Retry & Backoff**: Added `max_retries` and `retry_backoff_seconds` provider schema attributes with exponential backoff and rate limit (`HTTP 429` / `Retry-After`) handling.
- **Filtered Users Data Source**: Added `filter_user_type` and `filter_permissions_has` attributes to `seerr_users` data source.
- **Filtered Requests Data Source**: Added `filter_status`, `filter_media_type`, `filter_requested_by_id`, and `total` count attribute to `seerr_requests` data source.
- **Arr Entity Resolvers**: Added 5 dedicated data sources for direct Sonarr/Radarr querying:
  - `seerr_sonarr_root_folders`
  - `seerr_radarr_root_folders`
  - `seerr_sonarr_tags`
  - `seerr_radarr_tags`
  - `seerr_sonarr_language_profiles`

---

## 0.36.0 (2026-08-15)

### Features
- **Request Approval Lifecycle**: Added `seerr_request_approval` resource for declarative approve/decline workflow management (#163).
- **Issue Comments**: Added `seerr_issue_comment` resource with full CRUD for automated comments on Seerr issues.
- **Computed Attributes on Requests & Issues**: Added `created_at`, `updated_at`, `season_count`, `requested_by`, and `created_by` nested objects on `seerr_request` and `seerr_issue`.
- **Expanded Override Rules**: Added support for routing by `user_roles`, `genres`, `tag_ids`, and `languages` in `seerr_override_rule`.

---

## 0.35.0 (2026-08-15)

### Improvements
- **Import Support**: Added `ImportState` implementations to `seerr_api_key` and `seerr_discover_slider` (#162).
- **Schema Validation**: Added port range (1-65535) and hostname/URL validators across Radarr, Sonarr, and Tautulli resources.
- **Plan Modifiers & Defaults**: Added missing schema defaults and `UseStateForUnknown` / `RequiresReplace` plan modifiers across settings and user resources.

---

## 0.34.0 (2026-08-14)

### Features
- **TMDB Reference Data Sources**: Added `seerr_genres`, `seerr_languages`, and `seerr_regions` data sources for looking up reference metadata IDs (#159).

---

## 0.33.0 (2026-08-14)

### Features
- **User Lookup & Watchlists**: Added `seerr_users`, `seerr_user_invitations`, and `seerr_user_quota` data sources along with the `seerr_watchlist` resource (#158).

---

## 0.32.0 (2026-08-09)

### Features
- **Observability Data Sources**: Added `seerr_about`, `seerr_plex_devices`, and `seerr_pushover_sounds` data sources for infrastructure observability (#172).

---

## 0.31.1 (2026-08-08)

### Bug Fixes
- **Contract Validation**: Fixed schema contracts and validated HCL examples across all typed resources (#168).

---

## 0.31.0 (2026-08-07)

### Features
- **Stability & Hardening**: Comprehensive unit tests, HCL examples for every resource, and CI workflow concurrency groups (#166).

---

## 0.30.1 (2026-08-02)

### Bug Fixes
- **Linter**: Resolved golangci-lint static check issues (#156).

---

## 0.30.0 (2026-07-30)

### Features
- **User Notifications**: Added `seerr_user_notification_settings` resource and data source (#151).

---

## 0.29.0 (2026-07-29)

### Features
- **Jellyfin Settings**: Expanded Jellyfin settings and library parity (#149).

---

## 0.28.0 (2026-07-28)

### Features
- **Network & Tautulli**: Expanded network and Tautulli settings data sources and unit tests (#148).

---

## 0.27.1 (2026-07-28)

### Bug Fixes
- **Client Payloads**: Default empty POST/PUT/PATCH payload to `{}` and enforce `Content-Type: application/json` header (#146).

---

## 0.27.0 (2026-07-28)

### Features
- **Emby Libraries**: Added Emby library enablement and sync resources (#144).

---

## 0.26.0 (2026-07-26)

### Features
- **Rate Limiting**: Added HTTP client rate limiting and 429 / `Retry-After` header backoff (#143).

---

## 0.25.0 (2026-07-26)

### Features
- **Job Runs & Notification Tests**: Added `seerr_job_run` and `seerr_notification_agent_test` resources for triggering operational actions (#141).

---

## 0.24.0 (2026-07-26)

### Features
- **User Import**: Added Plex and Jellyfin user batch import resources and data sources (`seerr_user_import_plex`, `seerr_user_import_jellyfin`) (#133).

---

## 0.23.0 (2026-07-25)

### Features
- **User Quotas**: Added dedicated `seerr_user_quota` resource and data source for managing per-user request limits (#132).

---

## 0.22.0 (2026-07-25)

### Features
- **Library Sync**: Added Plex and Jellyfin library settings and manual library sync resources (#131).

---

## 0.21.0 (2026-07-21)

### Features
- **Module Validation**: Added automated CI validation for published modules and examples (#114).

---

## 0.20.0 (2026-07-16)

### Features
- **Async Status Polling**: Added asynchronous request status wait polling on create and update (#90).
