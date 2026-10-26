# Lifecycle contract coverage

Inventory of import/refresh, null versus empty, external deletion, retries, and secret handling for Seerr resources. The reusable suite is `internal/provider/lifecycle_contract_test.go` (`TestLifecycleContract*`). It uses `httptest` and the Plugin Framework `Read` / `ImportState` calls. It does not start Seerr and it does not set `TF_ACC`. Older unit tests stay in place.

Data sources are out of scope except where they share the HTTP client. Acceptance tests that require `TF_ACC` and a live instance are listed as `acc` and are not part of this suite.

## Legend

| Mark | Meaning |
|---|---|
| `suite` | Exercised by `TestLifecycleContract*` |
| `unit` | Older httptest or pure unit test, left in place |
| `schema` | Schema assertion only |
| `acc` | Acceptance test only (`TF_ACC`, live Seerr) |
| `client` | Covered by the shared client, not by a resource-level test |
| `code` | Implemented in `Read`, with no dedicated httptest |
| `n/a` | Contract does not apply |
| `gap` | Not covered |

## Shared rules

Every registered resource implements `resource.ResourceWithImportState` (`TestLifecycleContractResourcesSupportImport`). Most imports are `ImportStatePassthroughID` on `id`. Exceptions are called out per resource.

The client retries GET (and other safe methods) on 429, 502, 503, and 504. `max_retries` 0 is a single attempt. POST is not retried. That behavior is pinned in `client_retry_test.go`. The suite repeats it on a sample of resource `Read` methods. A failed read keeps state.

`TestLifecycleContractSensitiveSchema` fails when a resource gains or drops a `Sensitive` attribute without updating the golden list. Diagnostics on the suite's error paths must not contain the provider API key or the stored secret under test.

HTTP 404 on `Read` usually calls `RemoveResource` and returns no error. These resources do something else:

- `seerr_api_object` removes state only when `suppress_not_found` is true. Otherwise it returns an error and keeps state.
- `seerr_plex_library_sync`, `seerr_jellyfin_library_sync`, and `seerr_emby_library_sync` return an error and keep state.
- `seerr_watchlist` keeps state and returns an error when the watchlist route 404s or 405s. A complete empty page removes state.
- `seerr_user_invitation` never calls HTTP. `Read` always drops state with a warning. `ImportState` always errors.
- `seerr_discover_slider` removes state on a settings 404. An empty managed slider list stays in state (`TestDiscoverSliderReadKeepsEmptyManagedList`).

Omitted JSON and JSON `null` usually become a null attribute. `""` is an empty string. A few resources collapse those cases; they are called out below.

## Coverage by resource

| Resource | Import / refresh | Null vs empty | External deletion | Retries | Secrets |
|---|---|---|---|---|---|
| `seerr_api_object` | `suite`. Import sets `id` and `path`. Refresh stores the body. Import does not set `read_method`; the suite sets it before refresh. | `suite`. Body `""` stays empty. Body `{}` is stored as `{}`. | `suite`. 404 removes state only when `suppress_not_found` is true. | `client` | none |
| `seerr_job_schedule` | `suite`. Import id is the job id. Refresh sets `schedule` and `previous_schedule`. `unit` `TestJobScheduleRefreshSetsPreviousScheduleOnFirstRead`. | `suite`. `""`, JSON `null`, and an omitted cron all keep the prior schedule. | `suite`. HTTP 404 removes state. A job id missing from the array removes state. | `client` | none |
| `seerr_main_settings` | `suite`. Refresh forces id `main` and maps `applicationTitle` / `applicationUrl`. `unit` `TestMainSettingsRefreshStateReadsCanonicalValues`. | `suite` and `unit` `TestMainSettingsApplyDecodedSettingsClearsMissingValues`. Omitted strings become null. `""` stays empty. | `suite` and `unit` `TestResilience404Reconciliation`. | `suite` | none |
| `seerr_plex_settings` | `suite`. Refresh forces id `plex`. `acc` `TestAccPlexSettingsResource`. | `suite`. Omitted `name` becomes null. `""` stays empty. `unit` `TestPlexSettingsApplyDecodedSettingsClearsMissingOptionalValues`. | `suite` and `unit` `TestResilience404Reconciliation`. | `client` | none. The Plex token is a provider attribute. |
| `seerr_jellyfin_settings` | Passthrough import. Full import-then-refresh is a `gap`. | `suite` for `url_base`: omitted becomes null, `""` stays empty. `unit` apply-decoded test. | `suite` and `unit` `TestResilience404Reconciliation`. | `suite` | `schema` `api_key`. `suite`: omitted, JSON `null`, and `""` keep the prior key. A non-empty key replaces it. `unit` apply-decoded test. |
| `seerr_emby_settings` | Passthrough import. Full import-then-refresh is a `gap`. | `unit` apply-decoded test clears omitted optionals. | `suite` and `unit` `TestResilience404Reconciliation`. | `client` | `schema` `api_key`. `suite`: `""` keeps the prior key. `unit` apply-decoded test. |
| `seerr_notification_discord` | `suite` rejects an import id other than the agent name (asserted on gotify; the check is shared). Per-agent import-then-refresh is a `gap`. | `unit` `TestNormalizeNotificationModelAllAgentsOmittedAndPartialFields`. | `code` 404 removes state. `unit` `TestNotificationAgentMissingReturnsTrueFor404` covers the helper, not `Read`. | `client` | `schema` `discord.webhook_url`. |
| `seerr_notification_slack` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `slack.webhook_url`. |
| `seerr_notification_email` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. Empty secret strings are restored from prior state when that state is non-empty. | `code` 404 removes state. | `client` | `schema` `email.auth_pass`, `email.pgp_private_key`, `email.pgp_password`. HTTP refresh of those secrets is a `gap`. |
| `seerr_notification_lunasea` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `lunasea.webhook_url`. |
| `seerr_notification_telegram` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `telegram.bot_api`. HTTP refresh is a `gap`. |
| `seerr_notification_pushbullet` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `pushbullet.access_token`. HTTP refresh is a `gap`. |
| `seerr_notification_pushover` | Same shared import rule. `acc` `TestAccNotificationPushoverResource`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `pushover.access_token`, `pushover.user_token`. HTTP refresh is a `gap`. |
| `seerr_notification_ntfy` | Same shared import rule. `acc` `TestAccNotificationNtfyResource`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `ntfy.password`, `ntfy.token`. HTTP refresh is a `gap`. |
| `seerr_notification_webhook` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. | `code` 404 removes state. | `client` | `schema` `webhook.auth_header`, `webhook.webhook_url`. HTTP refresh is a `gap`. |
| `seerr_notification_gotify` | `suite`. Import id `gotify` refreshes `url` and `token`. Import id `slack` errors. | `suite`. An API `url` of `""` stays empty. An API `token` of `""` keeps the prior token. `unit` normalize test. | `code` 404 removes state. The gotify case is not in the deletion table. | `client` | `schema` `gotify.token`. `suite` HTTP refresh preserves the prior token when the API sends `""`. |
| `seerr_notification_webpush` | Same shared import rule. Per-agent refresh is a `gap`. | `unit` normalize test. The nested object has no attributes. | `code` 404 removes state. | `client` | none |
| `seerr_radarr_server` | `suite`. Import `3` sets `id` and `server_id`. Refresh leaves `api_key` null when state has no key, including when the API sends one. `tags: []` is an empty list. | `suite` and `unit` `TestReadRadarrStateFromJSON_*`. JSON `null` tags stay null. `[]` is an empty list. | `suite`. HTTP 404 removes state. An id missing from the array removes state. `unit` `TestResilience404Reconciliation`. | `client` | `schema` `api_key`. `suite`: state key wins over the upstream key, including upstream `""`. The JSON helper can fill an unset key; `Read` then copies the key from prior state. |
| `seerr_sonarr_server` | `suite`. Import `4` sets `id` and `server_id`. Same API-key rule as Radarr. `unit` JSON reader tests. | `suite`. `tags` and `anime_tags` keep null and `[]` distinct. `unit` `TestReadSonarrStateFromJSON_*`. | `suite` for 404 and for an id missing from the array. `unit` `TestResilience404Reconciliation`. | `suite` | `schema` `api_key`. `suite` keeps the state key over `upstream-secret`. |
| `seerr_user_permissions` | Passthrough import. Import-then-refresh is a `gap`. | `gap` | `unit` `TestResource404StateRemoval` and `TestResilience404Reconciliation`. | `client` | none |
| `seerr_user_watchlist_settings` | Passthrough import. Import-then-refresh is a `gap`. | `gap` | `unit` `TestResource404StateRemoval` and `TestResilience404Reconciliation`. | `client` | none |
| `seerr_api_key` | `suite` and `unit` `TestAPIKeyResourceImportAndReadLifecycle`. Refresh stores the server key, which must differ from the provider key. | `n/a` (computed key and status). | `suite` and `unit` `TestAPIKeyResourceRead404RemovesResource`. A 500 keeps state. | `suite` | `schema` `api_key`. `suite`: a 500 diagnostic contains neither the provider key nor the stored key. |
| `seerr_user` | `acc` `TestAccUserResource`. Unit import-then-refresh is a `gap`. | `unit` payload tests omit unset permissions and preserve remote notification values. Full `Read` null-versus-empty is a `gap`. | `unit` `TestResource404StateRemoval` and `TestResilience404Reconciliation`. | `client` | `schema` `plex_auth_token`, `jellyfin_password`, and the three `notification_settings` tokens. HTTP refresh of those secrets is a `gap`. |
| `seerr_discover_slider` | `unit` `TestDiscoverSliderResourceImportState`. | `unit` `TestDiscoverSliderReadNormalizesBlankOptionalFields`. An empty managed list stays in state. | `code` settings 404 removes state. Empty managed list does not. | `client` | none |
| `seerr_tautulli_settings` | Passthrough import. Import-then-refresh is a `gap`. | `unit` apply-decoded test. | `suite` and `unit` `TestResilience404Reconciliation`. | `client` | `schema` `api_key`. `suite`: `""` keeps the prior key. `unit` apply-decoded test. |
| `seerr_user_settings_permissions` | `acc` `TestAccUserSettingsPermissionsResource`. | `unit` `TestUserSettingsPermissionsReadPermissionsPreservesMissingValues`. | `unit` `TestResource404StateRemoval`. | `client` | none |
| `seerr_request` | `suite`. Import id is the request id. Refresh maps `media.id` to `seerr_media_id`. `unit` `TestRequestReadPopulatesComputedIDs`. | `suite`. Omitted `createdAt` is null. `""` stays empty. | `suite` and `unit` `TestRequestReadMarksStateMissingOn404`. | `client` for GET. Status polling is a separate loop (`TestWaitForRequestStatus*`). | none |
| `seerr_media_status` | Passthrough import. Import-then-refresh is a `gap`. | `unit` `TestMediaStatusReadPreservesReadOnlyStatuses`. | `code` 404 removes state. `TestMediaStatusTerraformProtocol` talks to httptest and sets `TF_ACC` itself. | `client` | none |
| `seerr_request_retry` | Passthrough import. Import-then-refresh is a `gap`. | `n/a` | `unit` `TestResource404StateRemoval`. | `client` | none |
| `seerr_issue` | `suite`. Import id is the issue id. Refresh maps `media.id` and `createdAt`. `unit` `TestIssueReadPopulatesMediaID`. | `suite`. Omitted `createdAt` is null. `""` stays empty. `comments` JSON `null` and `[]` both set `comments_count` to 0. | `suite` and `unit` `TestIssueReadMarksStateMissingOn404`. | `client` | none |
| `seerr_backup_settings` | `suite`. Any import id refreshes to id `backups`. `unit` `TestBackupSettingsRefreshState`. | `suite`. Omitted retention and storage become null. `""` stays empty and retention `0` stays 0. `unit` `TestBackupSettingsApplyDecodedSettingsClearsMissingValues`. | `suite` and `unit` `TestResilience404Reconciliation`. | `suite` | none |
| `seerr_user_invitation` | `suite` and `unit` `TestUserInvitationResourceUnsupported`. Import always errors and makes no request. | `n/a` | `suite`. `Read` drops state with a warning and makes no request. There is no invitation API. | `n/a` | `n/a` |
| `seerr_network_settings` | Passthrough import. Import-then-refresh is a `gap`. | `suite` for a managed proxy hostname: omitted is null, `""` stays empty. `unit` `TestRefreshNetworkSettingsLeavesOmittedNestedBlocksUnmanaged` leaves unmanaged `proxy` and `dns_cache` nil. | `suite` and `unit` `TestResilience404Reconciliation`. | `client` | `schema` `proxy.password`. `suite`: an omitted password keeps the prior secret; explicit `""` replaces it. `unit` `TestRefreshNetworkSettingsPreservesManagedNestedBlocksAndProxyPassword`. |
| `seerr_metadata_settings` | `suite`. Refresh forces id `metadata`. `unit` `TestMetadataSettingsResourceReadUsesMetadatasEndpoint`. | `suite`. Omitted `anime` and JSON `null` are null. `tv` `""` stays empty. | `suite` and `unit` `TestResilience404Reconciliation`. | `client` | none |
| `seerr_override_rule` | Passthrough import. Import-then-refresh is a `gap`. | `unit` `TestApplyOverrideRuleBodyMapsFields` covers field mapping. Null versus `""` is a `gap`. | `code`. 404, or a missing id, removes state. | `client` | none |
| `seerr_blocklist` | `suite`. Import id is `<media_type>:<tmdb_id>`. Refresh reads `/api/v1/blocklist/{tmdbId}`. Omitted `createdAt` is null. `unit` `TestApplyBlocklistMapMapsFields`. | `suite`. Omitted `title` is null. `""` stays empty. | `suite` and `unit` `TestResource404StateRemoval`. | `client` | none |
| `seerr_plex_library_settings` | `unit` `TestPlexLibrarySettingsImportIDConsistency`. | `unit` library enablement tests. An empty enabled list disables every library. | `code` 404 removes state. | `client` | none |
| `seerr_jellyfin_library_settings` | Passthrough import. Import-id unit test is a `gap`. | `unit` `TestJellyfinLibrarySettingsCreateThenReadPersistsEnabledLibraries` and shared enablement tests. | `code` 404 removes state. | `client` | none |
| `seerr_emby_library_settings` | `unit` `TestEmbyLibrarySettingsImportIDConsistency`. | `unit` `TestEmbyLibrarySettingsResourceUpdateAndRead` and shared enablement tests. | `code` 404 removes state. | `client` | none |
| `seerr_plex_library_sync` | Passthrough-style import writes `id` only. Import-then-refresh is a `gap`. `unit` `TestPlexLibrarySyncTrigger`. | `n/a` for scan status. `triggers` is `Map(string)` and must be a typed null when unset. | `suite`. 404 returns an error and keeps state. | `client` | none |
| `seerr_jellyfin_library_sync` | Same as Plex library sync. `unit` `TestJellyfinLibrarySyncTrigger`. | Same typed-null `triggers` rule. | `suite`. 404 returns an error and keeps state. | `client` | none |
| `seerr_emby_library_sync` | Same as Plex library sync. `unit` `TestEmbyLibrarySyncTriggerAndStatus`. | Same typed-null `triggers` rule. | `suite`. 404 returns an error and keeps state. | `client` | none |
| `seerr_user_quota` | `acc` `TestAccUserQuotaResource`. | `unit` `TestApplyQuotaResponseDefaultsToZeroWhenAbsent`. | `unit` `TestResource404StateRemoval` and `TestResilience404Reconciliation`. | `client` | none |
| `seerr_user_import_plex` | Passthrough import. Import-then-refresh is a `gap`. `unit` `TestPlexUserImportPerformImport`. | `gap` | `code` 404 removes state. | `client` | none |
| `seerr_user_import_jellyfin` | Passthrough import. Import-then-refresh is a `gap`. `unit` `TestJellyfinUserImportPerformImport`. | `gap` | `code` 404 removes state. | `client` | none |
| `seerr_job_run` | Passthrough import. Import-then-refresh is a `gap`. `unit` `TestJobRunExecuteAndCancel`. | `n/a` | `gap` beyond the execute/cancel test's HTTP statuses. | `client` | none |
| `seerr_notification_agent_test` | Passthrough import. This is a one-shot action. `unit` `TestNotificationAgentTestExecute`. | `n/a` | `gap` | `client` | none |
| `seerr_user_notification_settings` | `suite`. Import `9` sets `id` and `user_id`. Refresh keeps a token written onto state after import. `acc` `TestAccUserNotificationSettingsResource`. | `suite`. A body of `{}` keeps prior `notification_types`. `"notificationTypes":{}` nulls `email` and still leaves the block. `unit` `TestPopulateUserNotificationSettingsFromMap`. | `suite` and `unit` `TestResilience404Reconciliation`. | `client` | `schema` `pushbullet_access_token`, `pushover_application_token`, `pushover_user_key`. `suite`: JSON `null` and omitted keys keep the prior token. Explicit `""` clears all three. |
| `seerr_watchlist` | `suite`. Import id is `<tmdb_id>:<media_type>`, for example `550:movie`. An empty `overview` stays null. | `suite` for empty overview. `unit` pagination and identity tests. | `suite` and `unit` `TestWatchlistResource_Read_404PreservesState`, `TestWatchlistResourceReadRemovesOnlyAfterCompleteAbsence`. Route 404 keeps state and errors. A complete empty page removes state. | `suite` retries the watchlist GET. `client` covers the shared loop. | none |
| `seerr_request_approval` | Passthrough import. Import-then-refresh is a `gap`. | `unit` `TestRequestApprovalNormalizeStatus`. | `unit` `TestResource404StateRemoval` and `TestRequestApprovalRead404MarksStateMissing`. | `client` | none |
| `seerr_issue_comment` | Passthrough import. Import-then-refresh is a `gap`. `unit` `TestIssueCommentReadPopulatesFields`. | `gap` | `unit` `TestIssueCommentReadMarksMissingOn404`. | `client` | none |

## Notification Read and secrets

Each notification resource schema contains only that agent's nested attribute. `Read` used to `Get` the full `NotificationAgentModel`, which cannot succeed, so secret normalization never saw prior state. `Read` now uses `readNotificationClientModel`, the same path as create and plan. `TestLifecycleContractSecrets/gotify_empty_token_preserves_state` covers that path over HTTP.

`normalizeNotificationModel` treats an empty secret string like null and restores a non-empty prior value. `normalizeSetField` turns an empty notification-type set back into null when the prior set was null.

## Remaining gaps

- Resource-level retry tests cover `seerr_api_key`, `seerr_main_settings`, `seerr_backup_settings`, `seerr_jellyfin_settings`, `seerr_sonarr_server`, and `seerr_watchlist`. Every other resource relies on `client_retry_test.go`.
- `seerr_user` import/refresh is acceptance-only.
- Library settings, override rules, media status, user import, job run, the notification-agent test action, and the notification agents other than gotify are not in the HTTP import-then-refresh table.
- Issue `comments` JSON `null` and `[]` both become `comments_count` 0. They are not distinct in state.
