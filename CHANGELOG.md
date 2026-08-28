# Changelog

All notable changes to `terraform-provider-seerr` are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.39.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.38.3...v0.39.0) (2026-08-28)


### Features

* **ci:** add RC prerelease workflow, GoReleaser snapshot check, and align with chaptarr ([#214](https://github.com/Josh-Archer/terraform-provider-seerr/issues/214)) ([e64f2f2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e64f2f299abba9d86b577c3193bb963cd782568c))
* **observability:** add dashboards, drift detection, and recovery runbooks ([#184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184)) ([5673959](https://github.com/Josh-Archer/terraform-provider-seerr/commit/56739599075f588067667e44c4fbc8b77a931610))
* **resilience:** graceful 404 state removal and schema resilience ([#169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169)) ([#209](https://github.com/Josh-Archer/terraform-provider-seerr/issues/209)) ([3f26fd6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3f26fd685bd64d0c2598b7ea726983ed5a7ae89d))
* **upstream:** sync OpenAPI spec with upstream develop and update coverage ([#197](https://github.com/Josh-Archer/terraform-provider-seerr/issues/197)) ([#212](https://github.com/Josh-Archer/terraform-provider-seerr/issues/212)) ([f35668f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f35668f8805e7cab303d20522c98ce3c911f7369))


### Bug Fixes

* **ci:** fetch tags after remote creation so GoReleaser can validate HEAD ([#215](https://github.com/Josh-Archer/terraform-provider-seerr/issues/215)) ([04a4105](https://github.com/Josh-Archer/terraform-provider-seerr/commit/04a41057786327b9130607e70b72ef899f58879d))
* **ci:** resolve OpenTofu test index evaluation and improve upstream issue deduplication ([#221](https://github.com/Josh-Archer/terraform-provider-seerr/issues/221)) ([b088c6d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b088c6d3b22e80bd431b74a5d483880f03f01a11))
* resolve CodeQL incorrect integer conversion in provider env parsing ([#213](https://github.com/Josh-Archer/terraform-provider-seerr/issues/213)) ([35a5c95](https://github.com/Josh-Archer/terraform-provider-seerr/commit/35a5c95b644fbc4e4a60cac5b828c85caced3fce))


### Miscellaneous Chores

* **release:** configure release-please to bump minor version for feat commits pre-1.0 ([#210](https://github.com/Josh-Archer/terraform-provider-seerr/issues/210)) ([aef9ca8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aef9ca88bc47c2ca91f2e9787ce78469184f17aa))


### Continuous Integration

* remove legacy auto-tag from test.yml in favor of staged release-please ([#211](https://github.com/Josh-Archer/terraform-provider-seerr/issues/211)) ([b8a2e50](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b8a2e505b2673ddc18754238ceb6f8036564054b))

## [0.38.3](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.38.2...v0.38.3) (2026-08-22)


### Bug Fixes

* **api_key:** add id attribute to seerr_api_key schema and align import state ([#206](https://github.com/Josh-Archer/terraform-provider-seerr/issues/206)) ([163f4be](https://github.com/Josh-Archer/terraform-provider-seerr/commit/163f4bed58a34229fbfa9031ca816fcdfa6d7734))
* **library_settings:** normalize singleton ID during import to prevent state mutation drift ([#207](https://github.com/Josh-Archer/terraform-provider-seerr/issues/207)) ([b6e55bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b6e55bc3827714a1104dd9551a8497bbd4f12e3f))


### Miscellaneous Chores

* **deps:** bump the github-actions group across 1 directory with 4 updates ([#191](https://github.com/Josh-Archer/terraform-provider-seerr/issues/191)) ([c6acb52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c6acb52fa5e38f1efc985a7783c822e43c8b802e))


### Continuous Integration

* disable golangci-lint remote schema verification and add auto-merge for owner PRs ([#202](https://github.com/Josh-Archer/terraform-provider-seerr/issues/202)) ([76be0d0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/76be0d0c7d8e17267689bc8611a12c06f7731dbd))

## [0.38.2](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.38.1...v0.38.2) (2026-08-22)


### Features

* bulk import and migration CLI tooling with HCL generator and migration guide ([#170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170)) ([#203](https://github.com/Josh-Archer/terraform-provider-seerr/issues/203)) ([37a44e6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/37a44e679ca29211b69e039bcb286f83eeb25bad))
* community readiness - issue and PR templates, devcontainer, contributing guide, and governance ([#183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183)) ([#201](https://github.com/Josh-Archer/terraform-provider-seerr/issues/201)) ([83a2b10](https://github.com/Josh-Archer/terraform-provider-seerr/commit/83a2b1036393351c3587246974566a2545baba16))
* upstream compatibility automation - scheduled OpenAPI diff, release watcher, and compatibility matrix ([#182](https://github.com/Josh-Archer/terraform-provider-seerr/issues/182)) ([#195](https://github.com/Josh-Archer/terraform-provider-seerr/issues/195)) ([62a87a3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a87a3e1b80d5bcffe6acfb5b87ce3d0d5a96a6))


### Bug Fixes

* **ci:** remove environment gate from release.yml for zero-click publishing ([#198](https://github.com/Josh-Archer/terraform-provider-seerr/issues/198)) ([26cbf40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/26cbf4008be4895afe2780007ad522506f6eacbd))
* **ci:** set draft: true for release-please to allow GoReleaser publishing ([#193](https://github.com/Josh-Archer/terraform-provider-seerr/issues/193)) ([f4d86ae](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f4d86ae67231451ac43c6a4ce2bdfc53af3414e0))


### Documentation

* add ROADMAP.md and update progress matrix ([#199](https://github.com/Josh-Archer/terraform-provider-seerr/issues/199)) ([ff71232](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff71232bdc34a18f90e470332887f15c118f298e))
* consolidate roadmap phases, wording, and dual Terraform/OpenTofu support ([#200](https://github.com/Josh-Archer/terraform-provider-seerr/issues/200)) ([e340501](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3405010ae94bdee49ba439d7379c28d4f42e5f2))


### Miscellaneous Chores

* **deps:** bump github.com/stretchr/testify ([#190](https://github.com/Josh-Archer/terraform-provider-seerr/issues/190)) ([088bd87](https://github.com/Josh-Archer/terraform-provider-seerr/commit/088bd87586465808bdf8ecd800639c0b94a65a94))

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
