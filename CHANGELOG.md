# Changelog

All notable changes to `terraform-provider-seerr` are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v1.0.5...v2.0.0) (2026-09-03)


### ⚠ BREAKING CHANGES

* Transition provider to v1.0.0 General Availability with formal SemVer 2.0 API stability and breaking-change guarantees.

### Features

* add `seerr_user` resource for managing users and their notification settings. ([453389c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/453389c90af4234e94203504185818fea51058ec))
* add email and pushbullet settings data sources ([#150](https://github.com/Josh-Archer/terraform-provider-seerr/issues/150)) ([80a076c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80a076ca3826057dd36329b4e3cb4df8a5dce83b))
* Add media requests, issues, and service status features ([#70](https://github.com/Josh-Archer/terraform-provider-seerr/issues/70)) ([cfce309](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cfce30977698c73e9c00fc49a5389033e82d8273))
* add media requests, issues, and service status resources ([fede0b2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fede0b2141ac585ce1565a7023c39e04c81df51d))
* add reusable module ecosystem ([#228](https://github.com/Josh-Archer/terraform-provider-seerr/issues/228)) ([99cdee0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/99cdee04530ab63296736e06d4d3a6e3719cfd98))
* add Seerr modules for ARR servers and notifications ([f987801](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f987801c56f234aac22454e24598ad1b63dcb929))
* Add Seerr user resource and data source with support for Plex import and notification settings management. ([0348176](https://github.com/Josh-Archer/terraform-provider-seerr/commit/034817666479e3313496e08814c52c9977294e7f))
* add seerr_emby_settings resource and data source ([#48](https://github.com/Josh-Archer/terraform-provider-seerr/issues/48)) ([81e5b20](https://github.com/Josh-Archer/terraform-provider-seerr/commit/81e5b20634ae00e5dd8906dbcf9d30dff4030b5a))
* add seerr_issues and seerr_requests data sources ([#45](https://github.com/Josh-Archer/terraform-provider-seerr/issues/45)) ([28f0bc4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/28f0bc4b9e1c1a20e6d442261f5349c0dc69d81c)), closes [#44](https://github.com/Josh-Archer/terraform-provider-seerr/issues/44)
* add seerr_permission_set data source ([#112](https://github.com/Josh-Archer/terraform-provider-seerr/issues/112)) ([ac82806](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ac828064ee9c45784208226527f6a5d9f3583238))
* add Terraform resource and data source for Seerr main settings. ([64470f4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/64470f471cf3d930f9c36d860fa41d4c6eeea7d3))
* add Terraform resource and data source for Seerr main settings. ([bcb847f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bcb847f1a1988441898e1d892c5dee364c3ebfe3))
* add Tier 3 TMDB reference data sources ([#159](https://github.com/Josh-Archer/terraform-provider-seerr/issues/159)) ([#175](https://github.com/Josh-Archer/terraform-provider-seerr/issues/175)) ([eb08106](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eb08106a558491cc299cf93edb9be23c98db498b))
* add Tofu test workflows for CI and agent ([f9b35a6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f9b35a6d2d4e03e1b9d20767823cc5d2d3bfac82))
* brownfield adoption ([2af4a9e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2af4a9e9296856d52fe12ad5248787d271bcf787))
* bulk import and migration CLI tooling with HCL generator and migration guide ([#170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170)) ([#203](https://github.com/Josh-Archer/terraform-provider-seerr/issues/203)) ([37a44e6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/37a44e679ca29211b69e039bcb286f83eeb25bad))
* centralize provider resource and data-source registration metadata ([#113](https://github.com/Josh-Archer/terraform-provider-seerr/issues/113)) ([2fff552](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2fff55225dd12359442dd6ed839cfab303b34ee3))
* **ci:** add RC prerelease workflow, GoReleaser snapshot check, and align with chaptarr ([#214](https://github.com/Josh-Archer/terraform-provider-seerr/issues/214)) ([e64f2f2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e64f2f299abba9d86b577c3193bb963cd782568c))
* **ci:** release automation with release-please, tag-on-merge, and registry verification ([#181](https://github.com/Josh-Archer/terraform-provider-seerr/issues/181)) ([#187](https://github.com/Josh-Archer/terraform-provider-seerr/issues/187)) ([34fb8c6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/34fb8c6e89d7d90f1945bbaa9815c925dbaf4cb3))
* community readiness - issue and PR templates, devcontainer, contributing guide, and governance ([#183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183)) ([#201](https://github.com/Josh-Archer/terraform-provider-seerr/issues/201)) ([83a2b10](https://github.com/Josh-Archer/terraform-provider-seerr/commit/83a2b1036393351c3587246974566a2545baba16))
* **compatibility:** verify current upstream releases ([#223](https://github.com/Josh-Archer/terraform-provider-seerr/issues/223)) ([bfb3e49](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bfb3e495cb90f9bccbf75d108231949c7bdf9c09))
* Developer experience - env vars, retry/backoff, filtered data sources, and arr resolvers ([#164](https://github.com/Josh-Archer/terraform-provider-seerr/issues/164)) ([#179](https://github.com/Josh-Archer/terraform-provider-seerr/issues/179)) ([24c153b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/24c153b9b4b5f38d62bec462e3891be687849c98))
* document and support plex-token bootstrap configuration ([#115](https://github.com/Josh-Archer/terraform-provider-seerr/issues/115)) ([9997af8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9997af8d1f8c33d2107dcda5b9fd029b85cee6c7))
* Emby library enable lists and sync actions (closes [#135](https://github.com/Josh-Archer/terraform-provider-seerr/issues/135)) ([#144](https://github.com/Josh-Archer/terraform-provider-seerr/issues/144)) ([6345ad0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6345ad09e662608917d41c6dfb7355a0425bf707))
* expand jellyfin settings and library parity ([#149](https://github.com/Josh-Archer/terraform-provider-seerr/issues/149)) ([fe9960b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fe9960bef2956afab8bc7b2eb420e7aace608a9f))
* expand network and tautulli settings data sources and unit tests ([#148](https://github.com/Josh-Archer/terraform-provider-seerr/issues/148)) ([0d3bec4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0d3bec4850c31fbe04af04494a1b7397a9a71385))
* HTTP client rate limiting and 429/Retry-After backoff (closes [#134](https://github.com/Josh-Archer/terraform-provider-seerr/issues/134)) ([#143](https://github.com/Josh-Archer/terraform-provider-seerr/issues/143)) ([0e86030](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0e86030fd0015c6c76cdf3c692d41344806d918d))
* implement first-class typed plex settings (Issue [#6](https://github.com/Josh-Archer/terraform-provider-seerr/issues/6)) ([833ee78](https://github.com/Josh-Archer/terraform-provider-seerr/commit/833ee7862f4e1ede74a77c93e4220af509c33cc5))
* implement first-class typed plex settings (Issue [#6](https://github.com/Josh-Archer/terraform-provider-seerr/issues/6)) ([e2ccbc6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e2ccbc6be4c6cbbe4151f7254730f65cbfc759b4))
* implement initial Seerr Terraform provider with API key resource and data sources for main settings and API key. ([cd9dd34](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cd9dd34a7505e5d96600a3e0dde36b8e22070f70))
* implement jellyfin settings resource and data source ([1f37784](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1f37784875fafd44541893689d77a60838ae3ace))
* implement missing notification types and events (Issue [#18](https://github.com/Josh-Archer/terraform-provider-seerr/issues/18)) ([707f0ac](https://github.com/Josh-Archer/terraform-provider-seerr/commit/707f0ac0e3a8b103fe528d95216531c59bdd4242))
* implement OpenTofu Seerr provider with release automation ([241a406](https://github.com/Josh-Archer/terraform-provider-seerr/commit/241a4060f10b07a8b9e226b48a667c425f467391))
* implement per-user notification settings resource and data source ([#151](https://github.com/Josh-Archer/terraform-provider-seerr/issues/151)) ([9d89088](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9d8908840d0ee2406a7f0b45bc4f19f7de947d00))
* implement seerr_job_schedule resource ([#28](https://github.com/Josh-Archer/terraform-provider-seerr/issues/28)) ([f9a19c2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f9a19c27722114099c66468705dc5fd3affe3b45))
* implement tautulli settings and fix notification agent registration ([f23ebfa](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f23ebfa2b9c1c5e9d623aa26459b9eb4b6919f04))
* job run/cancel and notification agent test actions (closes [#123](https://github.com/Josh-Archer/terraform-provider-seerr/issues/123)) ([#141](https://github.com/Josh-Archer/terraform-provider-seerr/issues/141)) ([5addcdd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5addcddc5f98158690ced4ee4040121772af1c83))
* **library:** add Plex and Jellyfin library settings and sync resources ([#121](https://github.com/Josh-Archer/terraform-provider-seerr/issues/121)) ([#131](https://github.com/Josh-Archer/terraform-provider-seerr/issues/131)) ([6b739d5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6b739d53b91d5f2b830cac6cf0f5db54c8a6f52d))
* **observability:** add dashboards, drift detection, and recovery runbooks ([#184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184)) ([5673959](https://github.com/Josh-Archer/terraform-provider-seerr/commit/56739599075f588067667e44c4fbc8b77a931610))
* **openapi:** OpenAPI coverage matrix and CI drift check ([#119](https://github.com/Josh-Archer/terraform-provider-seerr/issues/119)) ([#128](https://github.com/Josh-Archer/terraform-provider-seerr/issues/128)) ([d737ca5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d737ca515daa70e7b6c1eabcaa4ad016db854f03))
* **permissions:** generate PermissionsMap from seerr_permissions.ts ([#109](https://github.com/Josh-Archer/terraform-provider-seerr/issues/109)) ([#129](https://github.com/Josh-Archer/terraform-provider-seerr/issues/129)) ([b62e5f8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b62e5f8ee1999ba083d65da4aa977d29a145eb90))
* Phase 6 - Advanced Resource Lifecycle (request approvals, issue comments, computed attributes, and expanded override rules) ([#177](https://github.com/Josh-Archer/terraform-provider-seerr/issues/177)) ([bb95271](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bb9527108a52a7cc58285b0b91675ca5202e4c7d))
* Phase 8 - Production Readiness (integration tests, changelog, composite example) ([#165](https://github.com/Josh-Archer/terraform-provider-seerr/issues/165)) ([#186](https://github.com/Josh-Archer/terraform-provider-seerr/issues/186)) ([ac32f5a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ac32f5aca2f1737671f224ff48c96f3e2743b518))
* Plex Provider Support ([#41](https://github.com/Josh-Archer/terraform-provider-seerr/issues/41)) ([c2605ec](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c2605ec034a5adc94040bb7e73de3b9134c5f77f))
* promote provider to v1.0.0 General Availability ([#245](https://github.com/Josh-Archer/terraform-provider-seerr/issues/245)) ([b383ae1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b383ae1f99e14540e7289aa1f32069fbb3806cc2))
* provider ergonomics and validation ([#58](https://github.com/Josh-Archer/terraform-provider-seerr/issues/58)) ([ddb253d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ddb253d5dddf9b045bb4f6714cf031e8fb0f08c6))
* provider hardening - import support, validation, defaults, plan modifiers ([#162](https://github.com/Josh-Archer/terraform-provider-seerr/issues/162)) ([#176](https://github.com/Josh-Archer/terraform-provider-seerr/issues/176)) ([97261b1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/97261b1ba6fd92bca9076830b5f08d3ce4bb2288))
* replace custom semver script with mathieudutour/github-tag-action and add Copilot commit instructions ([09375d9](https://github.com/Josh-Archer/terraform-provider-seerr/commit/09375d9bccf9eeb30e3fbadd39cf3fa01f526150))
* **request:** poll async status after create/update ([#90](https://github.com/Josh-Archer/terraform-provider-seerr/issues/90)) ([6be397f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6be397f723fd3ad8d38ad233fc5f9a9d86832eb3))
* **resilience:** graceful 404 state removal and schema resilience ([#169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169)) ([#209](https://github.com/Josh-Archer/terraform-provider-seerr/issues/209)) ([3f26fd6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3f26fd685bd64d0c2598b7ea726983ed5a7ae89d))
* serialize singleton endpoints, parse validation errors, and decouple CI environment ([#262](https://github.com/Josh-Archer/terraform-provider-seerr/issues/262)) ([ff6d75f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff6d75f8de543ff7a8a974a735743322687a2270))
* strongly-typed notification agents ([#9](https://github.com/Josh-Archer/terraform-provider-seerr/issues/9)) ([80211e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80211e749d090ff9484f8bc10c6a9b29db035d46))
* strongly-typed notification agents ([#9](https://github.com/Josh-Archer/terraform-provider-seerr/issues/9)) ([65a8f1a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/65a8f1a20bf5ffc4b64ed7c43e9639350761ed3b))
* upstream compatibility automation - scheduled OpenAPI diff, release watcher, and compatibility matrix ([#182](https://github.com/Josh-Archer/terraform-provider-seerr/issues/182)) ([#195](https://github.com/Josh-Archer/terraform-provider-seerr/issues/195)) ([62a87a3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a87a3e1b80d5bcffe6acfb5b87ce3d0d5a96a6))
* **upstream:** sync OpenAPI spec with upstream develop and update coverage ([#197](https://github.com/Josh-Archer/terraform-provider-seerr/issues/197)) ([#212](https://github.com/Josh-Archer/terraform-provider-seerr/issues/212)) ([f35668f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f35668f8805e7cab303d20522c98ce3c911f7369))
* User Settings Permissions and State Fixes ([#38](https://github.com/Josh-Archer/terraform-provider-seerr/issues/38)) ([025b3c7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/025b3c742056f5eb2f4b1bd5e12c9f1360602531))
* **user-import:** add Plex and Jellyfin user import resources and data sources ([#124](https://github.com/Josh-Archer/terraform-provider-seerr/issues/124)) ([#133](https://github.com/Josh-Archer/terraform-provider-seerr/issues/133)) ([8b979b3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b979b398dc9b40678bdb4ac52f38aa0dccd26f5))
* **user-quota:** add seerr_user_quota resource and data source ([#122](https://github.com/Josh-Archer/terraform-provider-seerr/issues/122)) ([#132](https://github.com/Josh-Archer/terraform-provider-seerr/issues/132)) ([cc63bc0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc63bc0684f80dbe031cec6611f6ab36e6ee6bfa))
* **v0.31.0:** phase 1 stability, unit tests, HCL examples, and CI workflow concurrency ([#166](https://github.com/Josh-Archer/terraform-provider-seerr/issues/166)) ([60687c5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/60687c503e85dab63eb68711f666acf216cdedeb))
* **v0.32.0:** Phase 2 - Observability and bootstrapping data sources ([#172](https://github.com/Josh-Archer/terraform-provider-seerr/issues/172)) ([9aa1503](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9aa1503faf5409bd0b27913b8d9a6fc6da782547))
* **v0.33.0:** add Tier 2 user lookup data sources and watchlist resource ([#158](https://github.com/Josh-Archer/terraform-provider-seerr/issues/158)) ([#173](https://github.com/Josh-Archer/terraform-provider-seerr/issues/173)) ([aad1555](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aad15557dbf706441e117ea83671eb16b2b42442))
* validate published modules and module examples in CI ([#114](https://github.com/Josh-Archer/terraform-provider-seerr/issues/114)) ([4a8fa86](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4a8fa86c8dc026567e76ffa7f3ad585fe71f0130))


### Bug Fixes

* Add `seerr_public_settings` data source and `seerr_main_settings` resource and data source with corresponding documentation, examples, and tests. ([dbfe08f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/dbfe08f3faedf0a120537c604a8ed66867031d5a))
* Add `tfplugindocs` integration and generated documentation for Seerr resources and data sources. ([9822c5b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9822c5bcf54ef2233198a373ee2369f27c300763))
* add GitHub Actions workflow for CI, OpenTofu integration tests, and automatic tagging. ([62a0ab5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a0ab5385be837b6b46a2775f937b7eb818f355))
* Add Radarr and Sonarr server resources with lifecycle tests to t… ([#26](https://github.com/Josh-Archer/terraform-provider-seerr/issues/26)) ([fa38980](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa3898061c625784af52277bdcfe7eae062330a7))
* add Terraform resources for Sonarr and Radarr servers and data sources for their quality profiles. ([7263fea](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7263fea8f65ce9ad5b292275278406f7df1cf109))
* **api_key:** add id attribute to seerr_api_key schema and align import state ([#206](https://github.com/Josh-Archer/terraform-provider-seerr/issues/206)) ([163f4be](https://github.com/Josh-Archer/terraform-provider-seerr/commit/163f4bed58a34229fbfa9031ca816fcdfa6d7734))
* arr scans variable ([7a72022](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7a720229495ac5139be8f51ae75bee1f2e5fc878))
* arr scans variable ([faf4c41](https://github.com/Josh-Archer/terraform-provider-seerr/commit/faf4c417b8f12af18934b8ac8873c18d7420eec8))
* **ci:** correct release-please-action commit sha in release-please.yml ([#188](https://github.com/Josh-Archer/terraform-provider-seerr/issues/188)) ([03e1e35](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e1e354dd87e7d575553e9861c46cfda8d0497c))
* **ci:** create tags for draft releases ([#226](https://github.com/Josh-Archer/terraform-provider-seerr/issues/226)) ([18ed2de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/18ed2de8764b6c8eea77a73f46194203ea696cde))
* **ci:** determine prerelease RC version from conventional commits ([#250](https://github.com/Josh-Archer/terraform-provider-seerr/issues/250)) ([5422bd5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5422bd5ecedd899737d47fb43e6cc6b2a74f3018))
* **ci:** fetch tags after remote creation so GoReleaser can validate HEAD ([#215](https://github.com/Josh-Archer/terraform-provider-seerr/issues/215)) ([04a4105](https://github.com/Josh-Archer/terraform-provider-seerr/commit/04a41057786327b9130607e70b72ef899f58879d))
* **ci:** harden release reconciliation to discover draft releases and self-heal ([#236](https://github.com/Josh-Archer/terraform-provider-seerr/issues/236)) ([63d76fc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/63d76fc34616ca4cf2157d1cfb7cd178ee005c15))
* **ci:** pass explicit tag to goreleaser ([#227](https://github.com/Josh-Archer/terraform-provider-seerr/issues/227)) ([9633ada](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9633adae29b4d881be99692a9864b71a44bef107))
* **ci:** publish releases directly and preserve changelog in GoReleaser ([#249](https://github.com/Josh-Archer/terraform-provider-seerr/issues/249)) ([8908a2c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8908a2cc0253fab7cf61507917fa8d9688f3d785))
* **ci:** query the OpenTofu provider endpoint ([#229](https://github.com/Josh-Archer/terraform-provider-seerr/issues/229)) ([6095c2d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6095c2db2e0f50339e6b28b6f48062b78042b250))
* **ci:** remove environment gate from release.yml for zero-click publishing ([#198](https://github.com/Josh-Archer/terraform-provider-seerr/issues/198)) ([26cbf40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/26cbf4008be4895afe2780007ad522506f6eacbd))
* **ci:** resolve OpenTofu test index evaluation and improve upstream issue deduplication ([#221](https://github.com/Josh-Archer/terraform-provider-seerr/issues/221)) ([b088c6d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b088c6d3b22e80bd431b74a5d483880f03f01a11))
* **ci:** set draft: true for release-please to allow GoReleaser publishing ([#193](https://github.com/Josh-Archer/terraform-provider-seerr/issues/193)) ([f4d86ae](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f4d86ae67231451ac43c6a4ce2bdfc53af3414e0))
* clean up files ([b66a0f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b66a0f525b6824a7ec4334e5d4c8f660ad6dbd00))
* **client:** default empty POST/PUT/PATCH payload to {} and set Content-Type header ([#146](https://github.com/Josh-Archer/terraform-provider-seerr/issues/146)) ([782cbf7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/782cbf76d3cee6e5130f35974dd2355e1292bc54))
* **client:** propagate error when retry delay context is cancelled ([#266](https://github.com/Josh-Archer/terraform-provider-seerr/issues/266)) ([d206aaf](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d206aaf4878e320f6557cff820b5b7f62cf61cb1))
* detect feat/ and feature/ branches in auto-tag minor version bump ([705a184](https://github.com/Josh-Archer/terraform-provider-seerr/commit/705a18405644c1e7d2e07a5cb807282ace58cc4c))
* edge casing ([737b796](https://github.com/Josh-Archer/terraform-provider-seerr/commit/737b796dd5dc2f419e3dc3245cfd7b60c9a8e369))
* finalize seerr_job_schedule and seerr_discover_slider, update n… ([#29](https://github.com/Josh-Archer/terraform-provider-seerr/issues/29)) ([b42ac13](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b42ac137ae3cd25208930e852607b4166d3586c7))
* issue 33 request timeout ([#57](https://github.com/Josh-Archer/terraform-provider-seerr/issues/57)) ([756ea22](https://github.com/Josh-Archer/terraform-provider-seerr/commit/756ea220c70a3bbd6e745d777b5b9cdff0efcfe7))
* **issue:** fail status updates when Seerr returns HTTP errors ([#87](https://github.com/Josh-Archer/terraform-provider-seerr/issues/87)) ([2472d8c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2472d8cc14b068a9a98b37e55220a49aa283c5c1)), closes [#83](https://github.com/Josh-Archer/terraform-provider-seerr/issues/83)
* keep discover slider resource when managed list is empty ([#86](https://github.com/Josh-Archer/terraform-provider-seerr/issues/86)) ([2e263a6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2e263a6f2cce78e8b1da20481104d679d590279e)), closes [#84](https://github.com/Josh-Archer/terraform-provider-seerr/issues/84)
* **library_settings:** normalize empty enabled_libraries slice to non-nil empty set ([#142](https://github.com/Josh-Archer/terraform-provider-seerr/issues/142)) ([ad49fb6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ad49fb6227ee6d5bb3a7b19f145308386cafd68e))
* **library_settings:** normalize singleton ID during import to prevent state mutation drift ([#207](https://github.com/Josh-Archer/terraform-provider-seerr/issues/207)) ([b6e55bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b6e55bc3827714a1104dd9551a8497bbd4f12e3f))
* **lint:** resolve golangci-lint issues to pass CI ([#156](https://github.com/Josh-Archer/terraform-provider-seerr/issues/156)) ([cda0343](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cda0343d378e7b1bf67da5414cfa76f8eb12b64c))
* normalize discover sliders and verify notification deletes ([#63](https://github.com/Josh-Archer/terraform-provider-seerr/issues/63)) ([5a6c6c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5a6c6c06c27f47be6506160c3357975cb5a14c9e))
* notification state handling and bootstrap ephemeral Seerr CI ([#40](https://github.com/Josh-Archer/terraform-provider-seerr/issues/40)) ([f6f2def](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f6f2defe89c57ec2b8cb69cf5b884c6713d93285))
* **notifications:** ensure optional notification attributes are Computed and option keys only sent when set ([#120](https://github.com/Josh-Archer/terraform-provider-seerr/issues/120)) ([#125](https://github.com/Josh-Archer/terraform-provider-seerr/issues/125)) ([edaa6ec](https://github.com/Josh-Archer/terraform-provider-seerr/commit/edaa6ec46323bae30fdf8e8f8f459340210425ee))
* **pagination:** ensure collection data sources complete full pagination without under-paging (closes [#137](https://github.com/Josh-Archer/terraform-provider-seerr/issues/137)) ([#145](https://github.com/Josh-Archer/terraform-provider-seerr/issues/145)) ([8124979](https://github.com/Josh-Archer/terraform-provider-seerr/commit/81249790ac97f6a63589b4d44b4ba745161d3016))
* **pagination:** safeIntFromAny direct int bounds check to fix CodeQL integer conversion alerts ([#155](https://github.com/Josh-Archer/terraform-provider-seerr/issues/155)) ([741b5bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/741b5bcdb8cc7e101d094f73c6475e4b93d76bd5))
* **pagination:** safely bound int64 to int conversions to fix CodeQL alerts ([#154](https://github.com/Josh-Archer/terraform-provider-seerr/issues/154)) ([d3947d3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d3947d3008a635e29d626a6e9321aa84db17cb99))
* panic on empty provider configuration in tests ([efd5762](https://github.com/Josh-Archer/terraform-provider-seerr/commit/efd5762d9f1a7b2c2dc7158d6e1056e938928dbc))
* **plex_library_settings:** omit empty enable query param when enabling zero libraries ([#140](https://github.com/Josh-Archer/terraform-provider-seerr/issues/140)) ([0041db6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0041db60148c90259e670be5fb11fc37dffa721a))
* preserve Seerr base URL subpaths when building API paths ([#85](https://github.com/Josh-Archer/terraform-provider-seerr/issues/85)) ([595c4c2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/595c4c2a47bbad57938216e8a1446d41fda7f5d4)), closes [#82](https://github.com/Josh-Archer/terraform-provider-seerr/issues/82)
* Preserve Seerr server IDs on update ([#25](https://github.com/Josh-Archer/terraform-provider-seerr/issues/25)) ([7ef6f12](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7ef6f12a7f2a6f5d70a9186098fa4a3db7523ca7))
* preserve user data source casing ([#110](https://github.com/Josh-Archer/terraform-provider-seerr/issues/110)) ([7e22892](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7e22892b30fb7ac580f9fda342b545225c75259b))
* prevent discover slider drift on read/delete ([#62](https://github.com/Josh-Archer/terraform-provider-seerr/issues/62)) ([11aa45c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/11aa45cdfe074c1932320f6d7b888bd1dacddf3f))
* Propagate diagnostics correctly from seerr_api_key Update ([#111](https://github.com/Josh-Archer/terraform-provider-seerr/issues/111)) ([6c5d26f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6c5d26ffdacbfdf63a7973e66ab3b1b2a944679e))
* quality profile issue with resources and update docs ([f8f0cbb](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f8f0cbbbc9310855b6b38cff0922fd9d69651b66))
* remove stale notification types state path ([9b76acd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9b76acd990ab5fdfff8173403f0f5b6dc45e4a97))
* resolve client response leak, servarr update state drift, and blocklist immutability ([#258](https://github.com/Josh-Archer/terraform-provider-seerr/issues/258)) ([0c17bc5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0c17bc50cc7c008d87e17d7b261b68af50d98bbf))
* resolve CodeQL incorrect integer conversion in provider env parsing ([#213](https://github.com/Josh-Archer/terraform-provider-seerr/issues/213)) ([35a5c95](https://github.com/Josh-Archer/terraform-provider-seerr/commit/35a5c95b644fbc4e4a60cac5b828c85caced3fce))
* resolve edge cases in retry config, empty arr collections, and synthetic IDs ([#180](https://github.com/Josh-Archer/terraform-provider-seerr/issues/180)) ([ef1d617](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ef1d617d936f68ecf05bcd184716daee0d94d259))
* resolve unknown value after apply and edge cases ([53ba55f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/53ba55fcef11d075d3c0c9da3fa6222c971562a6))
* resources normalize the model in a local copy inside payload ([80d2369](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80d2369530092872a90272aedabe121e0b111421))
* resources normalize the model in a local copy inside payload ([6f05948](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6f0594809f92382ead735f09e7de2d3ce96c86ac))
* restore release and Jellyfin compatibility gates ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7cb5743](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7cb57430b9fbab52b72ce7aa7f3852729fb3d291))
* satisfy lint rule for safe transport cloning ([23c51a9](https://github.com/Josh-Archer/terraform-provider-seerr/commit/23c51a9a587d8813b592f51d8e1e79d66c4df47d))
* schema exact match ([#60](https://github.com/Josh-Archer/terraform-provider-seerr/issues/60)) ([c4c7186](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c4c71865ac553e3690ec6bad9a1eb6c7b8635831))
* **servarr:** validate Arr connectivity via Seerr server proxy test endpoint ([#231](https://github.com/Josh-Archer/terraform-provider-seerr/issues/231)) ([2d128c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2d128c089885ba5bab8033aa516f585ea0767073))
* type radarr and sonarr server schemas ([#59](https://github.com/Josh-Archer/terraform-provider-seerr/issues/59)) ([c5c2aad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c5c2aadc882b247e49c2157c29995a16899d4138))
* typed notification resources/data sources no longer deserialize plan/state through the old generic schema shape ([#36](https://github.com/Josh-Archer/terraform-provider-seerr/issues/36)) ([a5b8e69](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a5b8e69be497e8da20501e648fdb0b1d3b049d51))
* validate examples and schema contracts ([#168](https://github.com/Josh-Archer/terraform-provider-seerr/issues/168)) ([f826b40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f826b40e79004a00fc9f50f9b4d16e2ac4f71cb3))
* versioning ([a0b2510](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a0b251018fc3f5563ee30efd6c4fce9da904d592))


### Documentation

* add HCL example snippets for v0.30.0 data sources (Closes [#152](https://github.com/Josh-Archer/terraform-provider-seerr/issues/152)) ([#153](https://github.com/Josh-Archer/terraform-provider-seerr/issues/153)) ([7bf2965](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7bf296549dd8007af2425f26548e07f24612ad41))
* Add import example ([#306](https://github.com/Josh-Archer/terraform-provider-seerr/issues/306)) ([f6b644e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f6b644efe282c406c3d1b147478e572f99880f28))
* add ROADMAP.md and update progress matrix ([#199](https://github.com/Josh-Archer/terraform-provider-seerr/issues/199)) ([ff71232](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff71232bdc34a18f90e470332887f15c118f298e))
* **auto:** regenerate provider documentation ([03e3637](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e3637461d00c476e9e85f325b2de2bde6d5040))
* **auto:** regenerate provider documentation ([b277888](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b27788883f60b94e1988bc0fad3c25d42f26641c))
* **auto:** regenerate provider documentation ([5a242e2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5a242e29e8916422003e1e5b09f00afce546cb80))
* **auto:** regenerate provider documentation ([fa4ebd7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa4ebd7f7d2d4bfd8e33a4107f931bfef2ee839a))
* **auto:** regenerate provider documentation ([965f19e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/965f19e691977ab84ab001fbb156b77d010cf25d))
* **auto:** regenerate provider documentation ([2f27fbb](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2f27fbbcdc317a7adac5acfecaf4708007cf522d))
* **auto:** regenerate provider documentation ([e3abd18](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3abd185d91b6c574685838211cc551cef6da47a))
* **auto:** regenerate provider documentation ([a3ed5e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a3ed5e84a115d89c7562cbb87f797f3f75b722fa))
* **auto:** regenerate provider documentation ([3931be8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3931be81661d9cb6ebc0c79cd865ca4ddf972270))
* **auto:** regenerate provider documentation ([cb264e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cb264e717c79006195a76fc32927f586bc0f6e13))
* **auto:** regenerate provider documentation ([2f68a3c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2f68a3c6e58372e898c00afa8f50fe37f106c16e))
* clarify reusable module usage ([#230](https://github.com/Josh-Archer/terraform-provider-seerr/issues/230)) ([31e1435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e1435968cc8e3ae592c8124c9db1860ebb4053))
* consolidate roadmap into README ([#237](https://github.com/Josh-Archer/terraform-provider-seerr/issues/237)) ([cc54970](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc54970d7e883cfda0580660f1e559115dd715e7))
* consolidate roadmap phases, wording, and dual Terraform/OpenTofu support ([#200](https://github.com/Josh-Archer/terraform-provider-seerr/issues/200)) ([e340501](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3405010ae94bdee49ba439d7379c28d4f42e5f2))
* enforce generated docs checks via pre-push + workflow ([#34](https://github.com/Josh-Archer/terraform-provider-seerr/issues/34)) ([ee026e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ee026e8eea94fe77d4cf98bf321b3b6dfc1a9f18))
* formalize v1.0.0 stability and breaking-change policy ([#243](https://github.com/Josh-Archer/terraform-provider-seerr/issues/243)) ([a1d3b73](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a1d3b73fae7d761cf6a0d99793d24b7021d7c10b))
* generate documentation for new resources ([55c69e2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/55c69e29716672d1e44c8b3964371349039a19cf))
* generate documentation for new resources and data sources ([15402f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/15402f51726fb705f0aa08fdb39549a76d59582a))
* mark v0.43.0 module release stable ([#241](https://github.com/Josh-Archer/terraform-provider-seerr/issues/241)) ([0ac711f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ac711f63685c962c0047d69190e2b795662357d))
* modernize README and docs for accuracy and conciseness ([#192](https://github.com/Josh-Archer/terraform-provider-seerr/issues/192)) ([8b4cc9b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b4cc9b7c3b8a0e4c0b2aaa001bd96b3f6a195a7))


### Miscellaneous Chores

* add agent instructions for Git workflow. ([#27](https://github.com/Josh-Archer/terraform-provider-seerr/issues/27)) ([6770a65](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6770a65d4754e21de0c11dd4b45f906008cc23db))
* **agents:** add GPG sign bypass instructions to repo workflow ([#116](https://github.com/Josh-Archer/terraform-provider-seerr/issues/116)) ([ff0a74b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff0a74bf05120c4cec901228b96c70df7eb78540))
* **agents:** configure subagents to sign commits using private key directly ([#117](https://github.com/Josh-Archer/terraform-provider-seerr/issues/117)) ([c179d8e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c179d8ee1fce8637d15c7baa87622ce31e7f8d34))
* **agents:** use global GPG/SSH key settings for signing commits ([#118](https://github.com/Josh-Archer/terraform-provider-seerr/issues/118)) ([d75d380](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d75d380202767106ca806088e61fcb6cde7aa9a6))
* change to `govet` ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7ad0b83](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7ad0b835d499d9929da0b28a47e653add79ac319))
* **deps:** bump github.com/stretchr/testify ([#126](https://github.com/Josh-Archer/terraform-provider-seerr/issues/126)) ([ccd5e15](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ccd5e1565fb7c430135fac65141786ceec43c954))
* **deps:** bump github.com/stretchr/testify ([#190](https://github.com/Josh-Archer/terraform-provider-seerr/issues/190)) ([088bd87](https://github.com/Josh-Archer/terraform-provider-seerr/commit/088bd87586465808bdf8ecd800639c0b94a65a94))
* **deps:** bump golang.org/x/crypto from 0.49.0 to 0.52.0 ([#92](https://github.com/Josh-Archer/terraform-provider-seerr/issues/92)) ([4eb4435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4eb4435eba68aa7f1b443a54354fc7a7524107bd))
* **deps:** bump golang.org/x/net from 0.54.0 to 0.55.0 ([#98](https://github.com/Josh-Archer/terraform-provider-seerr/issues/98)) ([53cc295](https://github.com/Josh-Archer/terraform-provider-seerr/commit/53cc2956012a457b31161d2e7f07f2df05593e22))
* **deps:** bump google.golang.org/grpc from 1.79.3 to 1.82.1 ([#130](https://github.com/Josh-Archer/terraform-provider-seerr/issues/130)) ([8e1053e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8e1053ead755c8bedd457be59d02d734de3d0ee8))
* **deps:** bump google.golang.org/grpc from 1.82.1 to 1.83.1 ([#264](https://github.com/Josh-Archer/terraform-provider-seerr/issues/264)) ([1d438cd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1d438cd7ebfcf18402718ac7bfa886e93dc4dedc))
* **deps:** bump the github-actions group across 1 directory with 10 updates ([#96](https://github.com/Josh-Archer/terraform-provider-seerr/issues/96)) ([e8aee52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e8aee52caa09bfe7c22c065d99a3166da744a5cf))
* **deps:** bump the github-actions group across 1 directory with 3 updates ([#222](https://github.com/Josh-Archer/terraform-provider-seerr/issues/222)) ([addedab](https://github.com/Josh-Archer/terraform-provider-seerr/commit/addedab1c5f6d72ed2d79a41ac8a7399f48cffcb))
* **deps:** bump the github-actions group across 1 directory with 4 updates ([#191](https://github.com/Josh-Archer/terraform-provider-seerr/issues/191)) ([c6acb52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c6acb52fa5e38f1efc985a7783c822e43c8b802e))
* **deps:** bump the github-actions group with 3 updates ([#174](https://github.com/Josh-Archer/terraform-provider-seerr/issues/174)) ([b620c2e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b620c2ec3259279a7eadf42e41c4dbc16aec3b52))
* **deps:** bump the github-actions group with 5 updates ([#127](https://github.com/Josh-Archer/terraform-provider-seerr/issues/127)) ([82b12de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/82b12de7889f8577cb20e9296547f7470bfaf686))
* **deps:** bump the go-dependencies group across 1 directory with 3 updates ([#95](https://github.com/Josh-Archer/terraform-provider-seerr/issues/95)) ([f7a54d8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f7a54d83cfbc4b0a86eb31b7b7b80fc9ae60250b))
* fix tofu CI - use dev_overrides and skip init for dev provider ([c39ace6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c39ace6e28462600ae079b5c5aec5fc33caf4581))
* fix tofu init in CI by setting job-level config env ([c47fc7a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c47fc7a7cf8d00f0c653e241758ee192d15dd085))
* fix tofu integration tests in CI using dev_overrides ([dbe6b04](https://github.com/Josh-Archer/terraform-provider-seerr/commit/dbe6b0407124acdd3c1effca8b8fa971f132f263))
* integrate Tofu tests into main CI and skip chores for tagging ([fa441e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa441e731d5e243c0c434e9a90d13180cac21c20))
* **main:** release 0.38.1 ([#189](https://github.com/Josh-Archer/terraform-provider-seerr/issues/189)) ([0122c97](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0122c97ba36c0c29090288738973ab921d65fb25))
* **main:** release 0.38.2 ([#194](https://github.com/Josh-Archer/terraform-provider-seerr/issues/194)) ([31e0531](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e05316e0c62160618443f184eff2bd2e04f531))
* **main:** release 0.38.3 ([#204](https://github.com/Josh-Archer/terraform-provider-seerr/issues/204)) ([5c270e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5c270e8f13b5847c6c5942ceb5b0fd30a0ea0b31))
* **main:** release 0.39.0 ([#208](https://github.com/Josh-Archer/terraform-provider-seerr/issues/208)) ([bcd4935](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bcd49352b4bad133b9d544cd826153a8083cf796))
* **main:** release 0.41.0 ([#224](https://github.com/Josh-Archer/terraform-provider-seerr/issues/224)) ([b10ea8c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b10ea8c83e0c072e9fb40f9529ae48a49aeb8613))
* **main:** release 0.42.0 ([#225](https://github.com/Josh-Archer/terraform-provider-seerr/issues/225)) ([d9aa9ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d9aa9adb743a3a26991f99aed948aedef6b6da8f))
* **main:** release 0.42.1 ([#232](https://github.com/Josh-Archer/terraform-provider-seerr/issues/232)) ([73a8c2b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/73a8c2ba8565c4aba4148ed292f7bd7d0cca22c6))
* **main:** release 0.42.2 ([#235](https://github.com/Josh-Archer/terraform-provider-seerr/issues/235)) ([8e6db2f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8e6db2fb14f3d0eef9329294332959646c023823))
* **main:** release 0.42.3 ([#239](https://github.com/Josh-Archer/terraform-provider-seerr/issues/239)) ([f39a71f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f39a71ffcb87e5e0ceb8c188f892856e1ff2a7bf))
* **main:** release 0.43.0 ([#242](https://github.com/Josh-Archer/terraform-provider-seerr/issues/242)) ([6b4c1b6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6b4c1b6bba41ba2d2f8d51967d5a6d0544bb2964))
* **main:** release 0.43.1 ([#244](https://github.com/Josh-Archer/terraform-provider-seerr/issues/244)) ([b82bdb2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b82bdb290d6a23137b42a83dfce602e5ff6c6d25))
* **main:** release 1.0.0 ([#247](https://github.com/Josh-Archer/terraform-provider-seerr/issues/247)) ([8a3c6de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8a3c6deff541dc7e50fc97d5708439f400649dba))
* **main:** release 1.0.1 ([#248](https://github.com/Josh-Archer/terraform-provider-seerr/issues/248)) ([888b4c6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/888b4c63930f2187fe3ecd72637e4204b9addf44))
* **main:** release 1.0.2 ([#259](https://github.com/Josh-Archer/terraform-provider-seerr/issues/259)) ([bb12fe7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bb12fe74ae530348d49b0c7ad9cce08dad14a725))
* **main:** release 1.0.3 ([#261](https://github.com/Josh-Archer/terraform-provider-seerr/issues/261)) ([e47342f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e47342fe847cf0c79b8727dbd5efeca5ceaf237a))
* **main:** release 1.0.4 ([#267](https://github.com/Josh-Archer/terraform-provider-seerr/issues/267)) ([eb3f546](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eb3f546a9253d67052c0efe00307a6ed560e1c02))
* **main:** release 1.0.5 ([#268](https://github.com/Josh-Archer/terraform-provider-seerr/issues/268)) ([046375c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/046375c51a6dada6b0ec196e8c77d0186e0215c8))
* migrate to native skipping via default_bump: false in tag action ([d571bb6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d571bb632e3b34086cbe2b6f7f40b7f33ce7db63))
* **release:** bump version to 0.40.0 ([bf608f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bf608f56c9836c2ba43a244956e4d6f3016d784f))
* **release:** configure release-please draft mode for GoReleaser publishing ([#260](https://github.com/Josh-Archer/terraform-provider-seerr/issues/260)) ([9f84f6f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9f84f6fce4aa86a1395b0c9af83e96db607852e9))
* **release:** configure release-please for v1.0.0 post-major SemVer ([#246](https://github.com/Josh-Archer/terraform-provider-seerr/issues/246)) ([19603c7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/19603c722189f0d31654206c607471c1c476b09e))
* **release:** configure release-please to bump minor version for feat commits pre-1.0 ([#210](https://github.com/Josh-Archer/terraform-provider-seerr/issues/210)) ([aef9ca8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aef9ca88bc47c2ca91f2e9787ce78469184f17aa))
* remove research files ([732e4d1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/732e4d18d83cc2fe69e12a0364653615dc8e5016))
* replace with copyloopvar ([#253](https://github.com/Josh-Archer/terraform-provider-seerr/issues/253)) ([1db5214](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1db52140d3463d1360b812c1cf537bafbff2b8a5))
* trigger signed release after adding gpg secrets ([4851d6e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4851d6e28be6deb4a3609ebc2f98e8ba60d6fb9d))
* Update minimum Go version in module ([#282](https://github.com/Josh-Archer/terraform-provider-seerr/issues/282)) ([4cfdd3b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4cfdd3b4951c90e4759d993a1399a6cbc591c92c))
* use -plugin-dir for tofu init in CI with explicit version ([0ef218f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ef218f673ad85e19b8940745c05185981193604))
* use filesystem mirror for tofu init in CI ([b1d46c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b1d46c06fe5b448859f73b10134d2784e1a86ee6))


### Tests

* align Jellyfin fixture authentication ([#240](https://github.com/Josh-Archer/terraform-provider-seerr/issues/240)) ([a32f3ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a32f3ad64abcc84e0ee165c6e41a7f207c00a9f9))
* **arch:** add Single Egress Rule architecture test and isolate test network fixtures ([#234](https://github.com/Josh-Archer/terraform-provider-seerr/issues/234)) ([ab1841e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ab1841ecac85c517d88a75e3f607e95814d55038))
* implement comprehensive OpenTofu integration tests for all provider features ([cbdb3c3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cbdb3c33e6610ed9bbe74f38712dd21e8113b5d6))


### Build System

* **deps:** bump github.com/cloudflare/circl from 1.6.1 to 1.6.3 ([#346](https://github.com/Josh-Archer/terraform-provider-seerr/issues/346)) ([3277b3c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3277b3cf37b87da164869f8b2d254241a2a51782))
* **deps:** bump github.com/cloudflare/circl in /tools ([#347](https://github.com/Josh-Archer/terraform-provider-seerr/issues/347)) ([a15d93e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a15d93e43295e34ad94ac67cd8a8bba0481de918))
* **deps:** bump github.com/hashicorp/terraform-plugin-framework ([#348](https://github.com/Josh-Archer/terraform-provider-seerr/issues/348)) ([c681e7e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c681e7ea74469ede4e0c90b7f1f11f57e00f798f))
* **deps:** bump github.com/hashicorp/terraform-plugin-go ([#345](https://github.com/Josh-Archer/terraform-provider-seerr/issues/345)) ([ba92b67](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ba92b67ebc725872123b7e1dfa3e58c80bdba20c))
* **deps:** bump the github-actions group across 1 directory with 3 updates ([#349](https://github.com/Josh-Archer/terraform-provider-seerr/issues/349)) ([3957103](https://github.com/Josh-Archer/terraform-provider-seerr/commit/395710343f1240a8a4a6af99bbc10b2383e5fc12))


### Continuous Integration

* add `golangci-lint` + fix lints ([#120](https://github.com/Josh-Archer/terraform-provider-seerr/issues/120)) ([fa8056c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa8056ce94248bf396295c0b208553f655782e91))
* Add GitHub Actions CI workflow for build, test, linting, and automated version tagging. ([f23eaa1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f23eaa1be4b62bb8033704635703ab98c0d33ad5))
* add hosted-runner fallback and remove template dependabot config ([aa59965](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aa5996555a7fabd0f71985ab8fa1709c2edd68d4))
* add OpenAPI coverage check step to CI workflow ([#147](https://github.com/Josh-Archer/terraform-provider-seerr/issues/147)) ([82a753f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/82a753f32123e20d9156da78892e8b72ecbcafaa))
* auto-tag main with semver after successful CI ([cb4faee](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cb4faeeda6c651e164a863a45cc21edee72cc4d3))
* avoid secrets in if expressions for release workflow ([cfb0351](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cfb0351d2a1c627d59195b7e4ca2bbf5cd9732a9))
* disable golangci-lint remote schema verification and add auto-merge for owner PRs ([#202](https://github.com/Josh-Archer/terraform-provider-seerr/issues/202)) ([76be0d0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/76be0d0c7d8e17267689bc8611a12c06f7731dbd))
* dispatch release after auto-tag ([eaccb1a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eaccb1ad26630d3ca67c955fee7046ff89daf0ac))
* dispatch Release after auto-tag on main ([#91](https://github.com/Josh-Archer/terraform-provider-seerr/issues/91)) ([aae4f46](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aae4f46a48d04b05180904d9276ffaa88f339eec))
* integration tests ([9c00c20](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9c00c205680a9a75c471215115269472d16250f7))
* remove legacy auto-tag from test.yml in favor of staged release-please ([#211](https://github.com/Josh-Archer/terraform-provider-seerr/issues/211)) ([b8a2e50](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b8a2e505b2673ddc18754238ceb6f8036564054b))
* replace custom PowerShell semver with mathieudutour/github-tag-action and add Copilot commit instructions ([1724cda](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1724cda5798c287fe72a253ac1375d230d495545))
* replace forked GH action with latest upstream ([#55](https://github.com/Josh-Archer/terraform-provider-seerr/issues/55)) ([3a91275](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3a91275abc3c45c61c888142c64c9d5abef588b2))
* support unsigned release fallback when GPG secrets are unset ([b674265](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b674265f4a0b4dcc178444e1ff907db45dd33be4))

## [1.0.5](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v1.0.4...v1.0.5) (2026-09-03)


### Miscellaneous Chores

* **deps:** bump google.golang.org/grpc from 1.82.1 to 1.83.1 ([#264](https://github.com/Josh-Archer/terraform-provider-seerr/issues/264)) ([1d438cd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1d438cd7ebfcf18402718ac7bfa886e93dc4dedc))

## [1.0.4](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v1.0.3...v1.0.4) (2026-09-03)


### Bug Fixes

* **client:** propagate error when retry delay context is cancelled ([#266](https://github.com/Josh-Archer/terraform-provider-seerr/issues/266)) ([d206aaf](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d206aaf4878e320f6557cff820b5b7f62cf61cb1))

## [1.0.3](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v1.0.2...v1.0.3) (2026-08-31)


### Miscellaneous Chores

* **release:** configure release-please draft mode for GoReleaser publishing ([#260](https://github.com/Josh-Archer/terraform-provider-seerr/issues/260)) ([9f84f6f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9f84f6fce4aa86a1395b0c9af83e96db607852e9))

## [1.0.2](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v1.0.1...v1.0.2) (2026-08-31)


### Bug Fixes

* resolve client response leak, servarr update state drift, and blocklist immutability ([#258](https://github.com/Josh-Archer/terraform-provider-seerr/issues/258)) ([0c17bc5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0c17bc50cc7c008d87e17d7b261b68af50d98bbf))

## [1.0.1](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v1.0.0...v1.0.1) (2026-08-30)


### Bug Fixes

* **ci:** determine prerelease RC version from conventional commits ([#250](https://github.com/Josh-Archer/terraform-provider-seerr/issues/250)) ([5422bd5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5422bd5ecedd899737d47fb43e6cc6b2a74f3018))
* **ci:** publish releases directly and preserve changelog in GoReleaser ([#249](https://github.com/Josh-Archer/terraform-provider-seerr/issues/249)) ([8908a2c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8908a2cc0253fab7cf61507917fa8d9688f3d785))


### Miscellaneous Chores

* **main:** release 0.43.1 ([#244](https://github.com/Josh-Archer/terraform-provider-seerr/issues/244)) ([b82bdb2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b82bdb290d6a23137b42a83dfce602e5ff6c6d25))
* **main:** release 1.0.0 ([#247](https://github.com/Josh-Archer/terraform-provider-seerr/issues/247)) ([8a3c6de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8a3c6deff541dc7e50fc97d5708439f400649dba))

## [1.0.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.43.1...v1.0.0) (2026-08-30)


### ⚠ BREAKING CHANGES

* Transition provider to v1.0.0 General Availability with formal SemVer 2.0 API stability and breaking-change guarantees.

### Features

* add `seerr_user` resource for managing users and their notification settings. ([453389c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/453389c90af4234e94203504185818fea51058ec))
* add email and pushbullet settings data sources ([#150](https://github.com/Josh-Archer/terraform-provider-seerr/issues/150)) ([80a076c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80a076ca3826057dd36329b4e3cb4df8a5dce83b))
* Add media requests, issues, and service status features ([#70](https://github.com/Josh-Archer/terraform-provider-seerr/issues/70)) ([cfce309](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cfce30977698c73e9c00fc49a5389033e82d8273))
* add media requests, issues, and service status resources ([fede0b2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fede0b2141ac585ce1565a7023c39e04c81df51d))
* add reusable module ecosystem ([#228](https://github.com/Josh-Archer/terraform-provider-seerr/issues/228)) ([99cdee0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/99cdee04530ab63296736e06d4d3a6e3719cfd98))
* add Seerr modules for ARR servers and notifications ([f987801](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f987801c56f234aac22454e24598ad1b63dcb929))
* Add Seerr user resource and data source with support for Plex import and notification settings management. ([0348176](https://github.com/Josh-Archer/terraform-provider-seerr/commit/034817666479e3313496e08814c52c9977294e7f))
* add seerr_emby_settings resource and data source ([#48](https://github.com/Josh-Archer/terraform-provider-seerr/issues/48)) ([81e5b20](https://github.com/Josh-Archer/terraform-provider-seerr/commit/81e5b20634ae00e5dd8906dbcf9d30dff4030b5a))
* add seerr_issues and seerr_requests data sources ([#45](https://github.com/Josh-Archer/terraform-provider-seerr/issues/45)) ([28f0bc4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/28f0bc4b9e1c1a20e6d442261f5349c0dc69d81c)), closes [#44](https://github.com/Josh-Archer/terraform-provider-seerr/issues/44)
* add seerr_permission_set data source ([#112](https://github.com/Josh-Archer/terraform-provider-seerr/issues/112)) ([ac82806](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ac828064ee9c45784208226527f6a5d9f3583238))
* add Terraform resource and data source for Seerr main settings. ([64470f4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/64470f471cf3d930f9c36d860fa41d4c6eeea7d3))
* add Terraform resource and data source for Seerr main settings. ([bcb847f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bcb847f1a1988441898e1d892c5dee364c3ebfe3))
* add Tier 3 TMDB reference data sources ([#159](https://github.com/Josh-Archer/terraform-provider-seerr/issues/159)) ([#175](https://github.com/Josh-Archer/terraform-provider-seerr/issues/175)) ([eb08106](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eb08106a558491cc299cf93edb9be23c98db498b))
* add Tofu test workflows for CI and agent ([f9b35a6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f9b35a6d2d4e03e1b9d20767823cc5d2d3bfac82))
* brownfield adoption ([2af4a9e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2af4a9e9296856d52fe12ad5248787d271bcf787))
* bulk import and migration CLI tooling with HCL generator and migration guide ([#170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170)) ([#203](https://github.com/Josh-Archer/terraform-provider-seerr/issues/203)) ([37a44e6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/37a44e679ca29211b69e039bcb286f83eeb25bad))
* centralize provider resource and data-source registration metadata ([#113](https://github.com/Josh-Archer/terraform-provider-seerr/issues/113)) ([2fff552](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2fff55225dd12359442dd6ed839cfab303b34ee3))
* **ci:** add RC prerelease workflow, GoReleaser snapshot check, and align with chaptarr ([#214](https://github.com/Josh-Archer/terraform-provider-seerr/issues/214)) ([e64f2f2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e64f2f299abba9d86b577c3193bb963cd782568c))
* **ci:** release automation with release-please, tag-on-merge, and registry verification ([#181](https://github.com/Josh-Archer/terraform-provider-seerr/issues/181)) ([#187](https://github.com/Josh-Archer/terraform-provider-seerr/issues/187)) ([34fb8c6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/34fb8c6e89d7d90f1945bbaa9815c925dbaf4cb3))
* community readiness - issue and PR templates, devcontainer, contributing guide, and governance ([#183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183)) ([#201](https://github.com/Josh-Archer/terraform-provider-seerr/issues/201)) ([83a2b10](https://github.com/Josh-Archer/terraform-provider-seerr/commit/83a2b1036393351c3587246974566a2545baba16))
* **compatibility:** verify current upstream releases ([#223](https://github.com/Josh-Archer/terraform-provider-seerr/issues/223)) ([bfb3e49](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bfb3e495cb90f9bccbf75d108231949c7bdf9c09))
* Developer experience - env vars, retry/backoff, filtered data sources, and arr resolvers ([#164](https://github.com/Josh-Archer/terraform-provider-seerr/issues/164)) ([#179](https://github.com/Josh-Archer/terraform-provider-seerr/issues/179)) ([24c153b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/24c153b9b4b5f38d62bec462e3891be687849c98))
* document and support plex-token bootstrap configuration ([#115](https://github.com/Josh-Archer/terraform-provider-seerr/issues/115)) ([9997af8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9997af8d1f8c33d2107dcda5b9fd029b85cee6c7))
* Emby library enable lists and sync actions (closes [#135](https://github.com/Josh-Archer/terraform-provider-seerr/issues/135)) ([#144](https://github.com/Josh-Archer/terraform-provider-seerr/issues/144)) ([6345ad0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6345ad09e662608917d41c6dfb7355a0425bf707))
* expand jellyfin settings and library parity ([#149](https://github.com/Josh-Archer/terraform-provider-seerr/issues/149)) ([fe9960b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fe9960bef2956afab8bc7b2eb420e7aace608a9f))
* expand network and tautulli settings data sources and unit tests ([#148](https://github.com/Josh-Archer/terraform-provider-seerr/issues/148)) ([0d3bec4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0d3bec4850c31fbe04af04494a1b7397a9a71385))
* HTTP client rate limiting and 429/Retry-After backoff (closes [#134](https://github.com/Josh-Archer/terraform-provider-seerr/issues/134)) ([#143](https://github.com/Josh-Archer/terraform-provider-seerr/issues/143)) ([0e86030](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0e86030fd0015c6c76cdf3c692d41344806d918d))
* implement first-class typed plex settings (Issue [#6](https://github.com/Josh-Archer/terraform-provider-seerr/issues/6)) ([833ee78](https://github.com/Josh-Archer/terraform-provider-seerr/commit/833ee7862f4e1ede74a77c93e4220af509c33cc5))
* implement first-class typed plex settings (Issue [#6](https://github.com/Josh-Archer/terraform-provider-seerr/issues/6)) ([e2ccbc6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e2ccbc6be4c6cbbe4151f7254730f65cbfc759b4))
* implement initial Seerr Terraform provider with API key resource and data sources for main settings and API key. ([cd9dd34](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cd9dd34a7505e5d96600a3e0dde36b8e22070f70))
* implement jellyfin settings resource and data source ([1f37784](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1f37784875fafd44541893689d77a60838ae3ace))
* implement missing notification types and events (Issue [#18](https://github.com/Josh-Archer/terraform-provider-seerr/issues/18)) ([707f0ac](https://github.com/Josh-Archer/terraform-provider-seerr/commit/707f0ac0e3a8b103fe528d95216531c59bdd4242))
* implement OpenTofu Seerr provider with release automation ([241a406](https://github.com/Josh-Archer/terraform-provider-seerr/commit/241a4060f10b07a8b9e226b48a667c425f467391))
* implement per-user notification settings resource and data source ([#151](https://github.com/Josh-Archer/terraform-provider-seerr/issues/151)) ([9d89088](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9d8908840d0ee2406a7f0b45bc4f19f7de947d00))
* implement seerr_job_schedule resource ([#28](https://github.com/Josh-Archer/terraform-provider-seerr/issues/28)) ([f9a19c2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f9a19c27722114099c66468705dc5fd3affe3b45))
* implement tautulli settings and fix notification agent registration ([f23ebfa](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f23ebfa2b9c1c5e9d623aa26459b9eb4b6919f04))
* job run/cancel and notification agent test actions (closes [#123](https://github.com/Josh-Archer/terraform-provider-seerr/issues/123)) ([#141](https://github.com/Josh-Archer/terraform-provider-seerr/issues/141)) ([5addcdd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5addcddc5f98158690ced4ee4040121772af1c83))
* **library:** add Plex and Jellyfin library settings and sync resources ([#121](https://github.com/Josh-Archer/terraform-provider-seerr/issues/121)) ([#131](https://github.com/Josh-Archer/terraform-provider-seerr/issues/131)) ([6b739d5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6b739d53b91d5f2b830cac6cf0f5db54c8a6f52d))
* **observability:** add dashboards, drift detection, and recovery runbooks ([#184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184)) ([5673959](https://github.com/Josh-Archer/terraform-provider-seerr/commit/56739599075f588067667e44c4fbc8b77a931610))
* **openapi:** OpenAPI coverage matrix and CI drift check ([#119](https://github.com/Josh-Archer/terraform-provider-seerr/issues/119)) ([#128](https://github.com/Josh-Archer/terraform-provider-seerr/issues/128)) ([d737ca5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d737ca515daa70e7b6c1eabcaa4ad016db854f03))
* **permissions:** generate PermissionsMap from seerr_permissions.ts ([#109](https://github.com/Josh-Archer/terraform-provider-seerr/issues/109)) ([#129](https://github.com/Josh-Archer/terraform-provider-seerr/issues/129)) ([b62e5f8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b62e5f8ee1999ba083d65da4aa977d29a145eb90))
* Phase 6 - Advanced Resource Lifecycle (request approvals, issue comments, computed attributes, and expanded override rules) ([#177](https://github.com/Josh-Archer/terraform-provider-seerr/issues/177)) ([bb95271](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bb9527108a52a7cc58285b0b91675ca5202e4c7d))
* Phase 8 - Production Readiness (integration tests, changelog, composite example) ([#165](https://github.com/Josh-Archer/terraform-provider-seerr/issues/165)) ([#186](https://github.com/Josh-Archer/terraform-provider-seerr/issues/186)) ([ac32f5a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ac32f5aca2f1737671f224ff48c96f3e2743b518))
* Plex Provider Support ([#41](https://github.com/Josh-Archer/terraform-provider-seerr/issues/41)) ([c2605ec](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c2605ec034a5adc94040bb7e73de3b9134c5f77f))
* promote provider to v1.0.0 General Availability ([#245](https://github.com/Josh-Archer/terraform-provider-seerr/issues/245)) ([b383ae1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b383ae1f99e14540e7289aa1f32069fbb3806cc2))
* provider ergonomics and validation ([#58](https://github.com/Josh-Archer/terraform-provider-seerr/issues/58)) ([ddb253d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ddb253d5dddf9b045bb4f6714cf031e8fb0f08c6))
* provider hardening - import support, validation, defaults, plan modifiers ([#162](https://github.com/Josh-Archer/terraform-provider-seerr/issues/162)) ([#176](https://github.com/Josh-Archer/terraform-provider-seerr/issues/176)) ([97261b1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/97261b1ba6fd92bca9076830b5f08d3ce4bb2288))
* replace custom semver script with mathieudutour/github-tag-action and add Copilot commit instructions ([09375d9](https://github.com/Josh-Archer/terraform-provider-seerr/commit/09375d9bccf9eeb30e3fbadd39cf3fa01f526150))
* **request:** poll async status after create/update ([#90](https://github.com/Josh-Archer/terraform-provider-seerr/issues/90)) ([6be397f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6be397f723fd3ad8d38ad233fc5f9a9d86832eb3))
* **resilience:** graceful 404 state removal and schema resilience ([#169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169)) ([#209](https://github.com/Josh-Archer/terraform-provider-seerr/issues/209)) ([3f26fd6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3f26fd685bd64d0c2598b7ea726983ed5a7ae89d))
* strongly-typed notification agents ([#9](https://github.com/Josh-Archer/terraform-provider-seerr/issues/9)) ([80211e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80211e749d090ff9484f8bc10c6a9b29db035d46))
* strongly-typed notification agents ([#9](https://github.com/Josh-Archer/terraform-provider-seerr/issues/9)) ([65a8f1a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/65a8f1a20bf5ffc4b64ed7c43e9639350761ed3b))
* upstream compatibility automation - scheduled OpenAPI diff, release watcher, and compatibility matrix ([#182](https://github.com/Josh-Archer/terraform-provider-seerr/issues/182)) ([#195](https://github.com/Josh-Archer/terraform-provider-seerr/issues/195)) ([62a87a3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a87a3e1b80d5bcffe6acfb5b87ce3d0d5a96a6))
* **upstream:** sync OpenAPI spec with upstream develop and update coverage ([#197](https://github.com/Josh-Archer/terraform-provider-seerr/issues/197)) ([#212](https://github.com/Josh-Archer/terraform-provider-seerr/issues/212)) ([f35668f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f35668f8805e7cab303d20522c98ce3c911f7369))
* User Settings Permissions and State Fixes ([#38](https://github.com/Josh-Archer/terraform-provider-seerr/issues/38)) ([025b3c7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/025b3c742056f5eb2f4b1bd5e12c9f1360602531))
* **user-import:** add Plex and Jellyfin user import resources and data sources ([#124](https://github.com/Josh-Archer/terraform-provider-seerr/issues/124)) ([#133](https://github.com/Josh-Archer/terraform-provider-seerr/issues/133)) ([8b979b3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b979b398dc9b40678bdb4ac52f38aa0dccd26f5))
* **user-quota:** add seerr_user_quota resource and data source ([#122](https://github.com/Josh-Archer/terraform-provider-seerr/issues/122)) ([#132](https://github.com/Josh-Archer/terraform-provider-seerr/issues/132)) ([cc63bc0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc63bc0684f80dbe031cec6611f6ab36e6ee6bfa))
* **v0.31.0:** phase 1 stability, unit tests, HCL examples, and CI workflow concurrency ([#166](https://github.com/Josh-Archer/terraform-provider-seerr/issues/166)) ([60687c5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/60687c503e85dab63eb68711f666acf216cdedeb))
* **v0.32.0:** Phase 2 - Observability and bootstrapping data sources ([#172](https://github.com/Josh-Archer/terraform-provider-seerr/issues/172)) ([9aa1503](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9aa1503faf5409bd0b27913b8d9a6fc6da782547))
* **v0.33.0:** add Tier 2 user lookup data sources and watchlist resource ([#158](https://github.com/Josh-Archer/terraform-provider-seerr/issues/158)) ([#173](https://github.com/Josh-Archer/terraform-provider-seerr/issues/173)) ([aad1555](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aad15557dbf706441e117ea83671eb16b2b42442))
* validate published modules and module examples in CI ([#114](https://github.com/Josh-Archer/terraform-provider-seerr/issues/114)) ([4a8fa86](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4a8fa86c8dc026567e76ffa7f3ad585fe71f0130))


### Bug Fixes

* Add `seerr_public_settings` data source and `seerr_main_settings` resource and data source with corresponding documentation, examples, and tests. ([dbfe08f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/dbfe08f3faedf0a120537c604a8ed66867031d5a))
* Add `tfplugindocs` integration and generated documentation for Seerr resources and data sources. ([9822c5b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9822c5bcf54ef2233198a373ee2369f27c300763))
* add GitHub Actions workflow for CI, OpenTofu integration tests, and automatic tagging. ([62a0ab5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a0ab5385be837b6b46a2775f937b7eb818f355))
* Add Radarr and Sonarr server resources with lifecycle tests to t… ([#26](https://github.com/Josh-Archer/terraform-provider-seerr/issues/26)) ([fa38980](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa3898061c625784af52277bdcfe7eae062330a7))
* add Terraform resources for Sonarr and Radarr servers and data sources for their quality profiles. ([7263fea](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7263fea8f65ce9ad5b292275278406f7df1cf109))
* **api_key:** add id attribute to seerr_api_key schema and align import state ([#206](https://github.com/Josh-Archer/terraform-provider-seerr/issues/206)) ([163f4be](https://github.com/Josh-Archer/terraform-provider-seerr/commit/163f4bed58a34229fbfa9031ca816fcdfa6d7734))
* arr scans variable ([7a72022](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7a720229495ac5139be8f51ae75bee1f2e5fc878))
* arr scans variable ([faf4c41](https://github.com/Josh-Archer/terraform-provider-seerr/commit/faf4c417b8f12af18934b8ac8873c18d7420eec8))
* **ci:** correct release-please-action commit sha in release-please.yml ([#188](https://github.com/Josh-Archer/terraform-provider-seerr/issues/188)) ([03e1e35](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e1e354dd87e7d575553e9861c46cfda8d0497c))
* **ci:** create tags for draft releases ([#226](https://github.com/Josh-Archer/terraform-provider-seerr/issues/226)) ([18ed2de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/18ed2de8764b6c8eea77a73f46194203ea696cde))
* **ci:** fetch tags after remote creation so GoReleaser can validate HEAD ([#215](https://github.com/Josh-Archer/terraform-provider-seerr/issues/215)) ([04a4105](https://github.com/Josh-Archer/terraform-provider-seerr/commit/04a41057786327b9130607e70b72ef899f58879d))
* **ci:** harden release reconciliation to discover draft releases and self-heal ([#236](https://github.com/Josh-Archer/terraform-provider-seerr/issues/236)) ([63d76fc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/63d76fc34616ca4cf2157d1cfb7cd178ee005c15))
* **ci:** pass explicit tag to goreleaser ([#227](https://github.com/Josh-Archer/terraform-provider-seerr/issues/227)) ([9633ada](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9633adae29b4d881be99692a9864b71a44bef107))
* **ci:** query the OpenTofu provider endpoint ([#229](https://github.com/Josh-Archer/terraform-provider-seerr/issues/229)) ([6095c2d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6095c2db2e0f50339e6b28b6f48062b78042b250))
* **ci:** remove environment gate from release.yml for zero-click publishing ([#198](https://github.com/Josh-Archer/terraform-provider-seerr/issues/198)) ([26cbf40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/26cbf4008be4895afe2780007ad522506f6eacbd))
* **ci:** resolve OpenTofu test index evaluation and improve upstream issue deduplication ([#221](https://github.com/Josh-Archer/terraform-provider-seerr/issues/221)) ([b088c6d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b088c6d3b22e80bd431b74a5d483880f03f01a11))
* **ci:** set draft: true for release-please to allow GoReleaser publishing ([#193](https://github.com/Josh-Archer/terraform-provider-seerr/issues/193)) ([f4d86ae](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f4d86ae67231451ac43c6a4ce2bdfc53af3414e0))
* clean up files ([b66a0f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b66a0f525b6824a7ec4334e5d4c8f660ad6dbd00))
* **client:** default empty POST/PUT/PATCH payload to {} and set Content-Type header ([#146](https://github.com/Josh-Archer/terraform-provider-seerr/issues/146)) ([782cbf7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/782cbf76d3cee6e5130f35974dd2355e1292bc54))
* detect feat/ and feature/ branches in auto-tag minor version bump ([705a184](https://github.com/Josh-Archer/terraform-provider-seerr/commit/705a18405644c1e7d2e07a5cb807282ace58cc4c))
* edge casing ([737b796](https://github.com/Josh-Archer/terraform-provider-seerr/commit/737b796dd5dc2f419e3dc3245cfd7b60c9a8e369))
* finalize seerr_job_schedule and seerr_discover_slider, update n… ([#29](https://github.com/Josh-Archer/terraform-provider-seerr/issues/29)) ([b42ac13](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b42ac137ae3cd25208930e852607b4166d3586c7))
* issue 33 request timeout ([#57](https://github.com/Josh-Archer/terraform-provider-seerr/issues/57)) ([756ea22](https://github.com/Josh-Archer/terraform-provider-seerr/commit/756ea220c70a3bbd6e745d777b5b9cdff0efcfe7))
* **issue:** fail status updates when Seerr returns HTTP errors ([#87](https://github.com/Josh-Archer/terraform-provider-seerr/issues/87)) ([2472d8c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2472d8cc14b068a9a98b37e55220a49aa283c5c1)), closes [#83](https://github.com/Josh-Archer/terraform-provider-seerr/issues/83)
* keep discover slider resource when managed list is empty ([#86](https://github.com/Josh-Archer/terraform-provider-seerr/issues/86)) ([2e263a6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2e263a6f2cce78e8b1da20481104d679d590279e)), closes [#84](https://github.com/Josh-Archer/terraform-provider-seerr/issues/84)
* **library_settings:** normalize empty enabled_libraries slice to non-nil empty set ([#142](https://github.com/Josh-Archer/terraform-provider-seerr/issues/142)) ([ad49fb6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ad49fb6227ee6d5bb3a7b19f145308386cafd68e))
* **library_settings:** normalize singleton ID during import to prevent state mutation drift ([#207](https://github.com/Josh-Archer/terraform-provider-seerr/issues/207)) ([b6e55bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b6e55bc3827714a1104dd9551a8497bbd4f12e3f))
* **lint:** resolve golangci-lint issues to pass CI ([#156](https://github.com/Josh-Archer/terraform-provider-seerr/issues/156)) ([cda0343](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cda0343d378e7b1bf67da5414cfa76f8eb12b64c))
* normalize discover sliders and verify notification deletes ([#63](https://github.com/Josh-Archer/terraform-provider-seerr/issues/63)) ([5a6c6c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5a6c6c06c27f47be6506160c3357975cb5a14c9e))
* notification state handling and bootstrap ephemeral Seerr CI ([#40](https://github.com/Josh-Archer/terraform-provider-seerr/issues/40)) ([f6f2def](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f6f2defe89c57ec2b8cb69cf5b884c6713d93285))
* **notifications:** ensure optional notification attributes are Computed and option keys only sent when set ([#120](https://github.com/Josh-Archer/terraform-provider-seerr/issues/120)) ([#125](https://github.com/Josh-Archer/terraform-provider-seerr/issues/125)) ([edaa6ec](https://github.com/Josh-Archer/terraform-provider-seerr/commit/edaa6ec46323bae30fdf8e8f8f459340210425ee))
* **pagination:** ensure collection data sources complete full pagination without under-paging (closes [#137](https://github.com/Josh-Archer/terraform-provider-seerr/issues/137)) ([#145](https://github.com/Josh-Archer/terraform-provider-seerr/issues/145)) ([8124979](https://github.com/Josh-Archer/terraform-provider-seerr/commit/81249790ac97f6a63589b4d44b4ba745161d3016))
* **pagination:** safeIntFromAny direct int bounds check to fix CodeQL integer conversion alerts ([#155](https://github.com/Josh-Archer/terraform-provider-seerr/issues/155)) ([741b5bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/741b5bcdb8cc7e101d094f73c6475e4b93d76bd5))
* **pagination:** safely bound int64 to int conversions to fix CodeQL alerts ([#154](https://github.com/Josh-Archer/terraform-provider-seerr/issues/154)) ([d3947d3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d3947d3008a635e29d626a6e9321aa84db17cb99))
* panic on empty provider configuration in tests ([efd5762](https://github.com/Josh-Archer/terraform-provider-seerr/commit/efd5762d9f1a7b2c2dc7158d6e1056e938928dbc))
* **plex_library_settings:** omit empty enable query param when enabling zero libraries ([#140](https://github.com/Josh-Archer/terraform-provider-seerr/issues/140)) ([0041db6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0041db60148c90259e670be5fb11fc37dffa721a))
* preserve Seerr base URL subpaths when building API paths ([#85](https://github.com/Josh-Archer/terraform-provider-seerr/issues/85)) ([595c4c2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/595c4c2a47bbad57938216e8a1446d41fda7f5d4)), closes [#82](https://github.com/Josh-Archer/terraform-provider-seerr/issues/82)
* Preserve Seerr server IDs on update ([#25](https://github.com/Josh-Archer/terraform-provider-seerr/issues/25)) ([7ef6f12](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7ef6f12a7f2a6f5d70a9186098fa4a3db7523ca7))
* preserve user data source casing ([#110](https://github.com/Josh-Archer/terraform-provider-seerr/issues/110)) ([7e22892](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7e22892b30fb7ac580f9fda342b545225c75259b))
* prevent discover slider drift on read/delete ([#62](https://github.com/Josh-Archer/terraform-provider-seerr/issues/62)) ([11aa45c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/11aa45cdfe074c1932320f6d7b888bd1dacddf3f))
* Propagate diagnostics correctly from seerr_api_key Update ([#111](https://github.com/Josh-Archer/terraform-provider-seerr/issues/111)) ([6c5d26f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6c5d26ffdacbfdf63a7973e66ab3b1b2a944679e))
* quality profile issue with resources and update docs ([f8f0cbb](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f8f0cbbbc9310855b6b38cff0922fd9d69651b66))
* remove stale notification types state path ([9b76acd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9b76acd990ab5fdfff8173403f0f5b6dc45e4a97))
* resolve CodeQL incorrect integer conversion in provider env parsing ([#213](https://github.com/Josh-Archer/terraform-provider-seerr/issues/213)) ([35a5c95](https://github.com/Josh-Archer/terraform-provider-seerr/commit/35a5c95b644fbc4e4a60cac5b828c85caced3fce))
* resolve edge cases in retry config, empty arr collections, and synthetic IDs ([#180](https://github.com/Josh-Archer/terraform-provider-seerr/issues/180)) ([ef1d617](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ef1d617d936f68ecf05bcd184716daee0d94d259))
* resolve unknown value after apply and edge cases ([53ba55f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/53ba55fcef11d075d3c0c9da3fa6222c971562a6))
* resources normalize the model in a local copy inside payload ([80d2369](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80d2369530092872a90272aedabe121e0b111421))
* resources normalize the model in a local copy inside payload ([6f05948](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6f0594809f92382ead735f09e7de2d3ce96c86ac))
* restore release and Jellyfin compatibility gates ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7cb5743](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7cb57430b9fbab52b72ce7aa7f3852729fb3d291))
* satisfy lint rule for safe transport cloning ([23c51a9](https://github.com/Josh-Archer/terraform-provider-seerr/commit/23c51a9a587d8813b592f51d8e1e79d66c4df47d))
* schema exact match ([#60](https://github.com/Josh-Archer/terraform-provider-seerr/issues/60)) ([c4c7186](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c4c71865ac553e3690ec6bad9a1eb6c7b8635831))
* **servarr:** validate Arr connectivity via Seerr server proxy test endpoint ([#231](https://github.com/Josh-Archer/terraform-provider-seerr/issues/231)) ([2d128c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2d128c089885ba5bab8033aa516f585ea0767073))
* type radarr and sonarr server schemas ([#59](https://github.com/Josh-Archer/terraform-provider-seerr/issues/59)) ([c5c2aad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c5c2aadc882b247e49c2157c29995a16899d4138))
* typed notification resources/data sources no longer deserialize plan/state through the old generic schema shape ([#36](https://github.com/Josh-Archer/terraform-provider-seerr/issues/36)) ([a5b8e69](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a5b8e69be497e8da20501e648fdb0b1d3b049d51))
* validate examples and schema contracts ([#168](https://github.com/Josh-Archer/terraform-provider-seerr/issues/168)) ([f826b40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f826b40e79004a00fc9f50f9b4d16e2ac4f71cb3))
* versioning ([a0b2510](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a0b251018fc3f5563ee30efd6c4fce9da904d592))


### Documentation

* add HCL example snippets for v0.30.0 data sources (Closes [#152](https://github.com/Josh-Archer/terraform-provider-seerr/issues/152)) ([#153](https://github.com/Josh-Archer/terraform-provider-seerr/issues/153)) ([7bf2965](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7bf296549dd8007af2425f26548e07f24612ad41))
* Add import example ([#306](https://github.com/Josh-Archer/terraform-provider-seerr/issues/306)) ([f6b644e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f6b644efe282c406c3d1b147478e572f99880f28))
* add ROADMAP.md and update progress matrix ([#199](https://github.com/Josh-Archer/terraform-provider-seerr/issues/199)) ([ff71232](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff71232bdc34a18f90e470332887f15c118f298e))
* **auto:** regenerate provider documentation ([03e3637](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e3637461d00c476e9e85f325b2de2bde6d5040))
* **auto:** regenerate provider documentation ([b277888](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b27788883f60b94e1988bc0fad3c25d42f26641c))
* **auto:** regenerate provider documentation ([5a242e2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5a242e29e8916422003e1e5b09f00afce546cb80))
* **auto:** regenerate provider documentation ([fa4ebd7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa4ebd7f7d2d4bfd8e33a4107f931bfef2ee839a))
* **auto:** regenerate provider documentation ([965f19e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/965f19e691977ab84ab001fbb156b77d010cf25d))
* **auto:** regenerate provider documentation ([2f27fbb](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2f27fbbcdc317a7adac5acfecaf4708007cf522d))
* **auto:** regenerate provider documentation ([e3abd18](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3abd185d91b6c574685838211cc551cef6da47a))
* **auto:** regenerate provider documentation ([a3ed5e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a3ed5e84a115d89c7562cbb87f797f3f75b722fa))
* **auto:** regenerate provider documentation ([3931be8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3931be81661d9cb6ebc0c79cd865ca4ddf972270))
* **auto:** regenerate provider documentation ([cb264e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cb264e717c79006195a76fc32927f586bc0f6e13))
* **auto:** regenerate provider documentation ([2f68a3c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2f68a3c6e58372e898c00afa8f50fe37f106c16e))
* clarify reusable module usage ([#230](https://github.com/Josh-Archer/terraform-provider-seerr/issues/230)) ([31e1435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e1435968cc8e3ae592c8124c9db1860ebb4053))
* consolidate roadmap into README ([#237](https://github.com/Josh-Archer/terraform-provider-seerr/issues/237)) ([cc54970](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc54970d7e883cfda0580660f1e559115dd715e7))
* consolidate roadmap phases, wording, and dual Terraform/OpenTofu support ([#200](https://github.com/Josh-Archer/terraform-provider-seerr/issues/200)) ([e340501](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3405010ae94bdee49ba439d7379c28d4f42e5f2))
* enforce generated docs checks via pre-push + workflow ([#34](https://github.com/Josh-Archer/terraform-provider-seerr/issues/34)) ([ee026e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ee026e8eea94fe77d4cf98bf321b3b6dfc1a9f18))
* formalize v1.0.0 stability and breaking-change policy ([#243](https://github.com/Josh-Archer/terraform-provider-seerr/issues/243)) ([a1d3b73](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a1d3b73fae7d761cf6a0d99793d24b7021d7c10b))
* generate documentation for new resources ([55c69e2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/55c69e29716672d1e44c8b3964371349039a19cf))
* generate documentation for new resources and data sources ([15402f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/15402f51726fb705f0aa08fdb39549a76d59582a))
* mark v0.43.0 module release stable ([#241](https://github.com/Josh-Archer/terraform-provider-seerr/issues/241)) ([0ac711f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ac711f63685c962c0047d69190e2b795662357d))
* modernize README and docs for accuracy and conciseness ([#192](https://github.com/Josh-Archer/terraform-provider-seerr/issues/192)) ([8b4cc9b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b4cc9b7c3b8a0e4c0b2aaa001bd96b3f6a195a7))


### Miscellaneous Chores

* add agent instructions for Git workflow. ([#27](https://github.com/Josh-Archer/terraform-provider-seerr/issues/27)) ([6770a65](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6770a65d4754e21de0c11dd4b45f906008cc23db))
* **agents:** add GPG sign bypass instructions to repo workflow ([#116](https://github.com/Josh-Archer/terraform-provider-seerr/issues/116)) ([ff0a74b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff0a74bf05120c4cec901228b96c70df7eb78540))
* **agents:** configure subagents to sign commits using private key directly ([#117](https://github.com/Josh-Archer/terraform-provider-seerr/issues/117)) ([c179d8e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c179d8ee1fce8637d15c7baa87622ce31e7f8d34))
* **agents:** use global GPG/SSH key settings for signing commits ([#118](https://github.com/Josh-Archer/terraform-provider-seerr/issues/118)) ([d75d380](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d75d380202767106ca806088e61fcb6cde7aa9a6))
* change to `govet` ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7ad0b83](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7ad0b835d499d9929da0b28a47e653add79ac319))
* **deps:** bump github.com/stretchr/testify ([#126](https://github.com/Josh-Archer/terraform-provider-seerr/issues/126)) ([ccd5e15](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ccd5e1565fb7c430135fac65141786ceec43c954))
* **deps:** bump github.com/stretchr/testify ([#190](https://github.com/Josh-Archer/terraform-provider-seerr/issues/190)) ([088bd87](https://github.com/Josh-Archer/terraform-provider-seerr/commit/088bd87586465808bdf8ecd800639c0b94a65a94))
* **deps:** bump golang.org/x/crypto from 0.49.0 to 0.52.0 ([#92](https://github.com/Josh-Archer/terraform-provider-seerr/issues/92)) ([4eb4435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4eb4435eba68aa7f1b443a54354fc7a7524107bd))
* **deps:** bump golang.org/x/net from 0.54.0 to 0.55.0 ([#98](https://github.com/Josh-Archer/terraform-provider-seerr/issues/98)) ([53cc295](https://github.com/Josh-Archer/terraform-provider-seerr/commit/53cc2956012a457b31161d2e7f07f2df05593e22))
* **deps:** bump google.golang.org/grpc from 1.79.3 to 1.82.1 ([#130](https://github.com/Josh-Archer/terraform-provider-seerr/issues/130)) ([8e1053e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8e1053ead755c8bedd457be59d02d734de3d0ee8))
* **deps:** bump the github-actions group across 1 directory with 10 updates ([#96](https://github.com/Josh-Archer/terraform-provider-seerr/issues/96)) ([e8aee52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e8aee52caa09bfe7c22c065d99a3166da744a5cf))
* **deps:** bump the github-actions group across 1 directory with 3 updates ([#222](https://github.com/Josh-Archer/terraform-provider-seerr/issues/222)) ([addedab](https://github.com/Josh-Archer/terraform-provider-seerr/commit/addedab1c5f6d72ed2d79a41ac8a7399f48cffcb))
* **deps:** bump the github-actions group across 1 directory with 4 updates ([#191](https://github.com/Josh-Archer/terraform-provider-seerr/issues/191)) ([c6acb52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c6acb52fa5e38f1efc985a7783c822e43c8b802e))
* **deps:** bump the github-actions group with 3 updates ([#174](https://github.com/Josh-Archer/terraform-provider-seerr/issues/174)) ([b620c2e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b620c2ec3259279a7eadf42e41c4dbc16aec3b52))
* **deps:** bump the github-actions group with 5 updates ([#127](https://github.com/Josh-Archer/terraform-provider-seerr/issues/127)) ([82b12de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/82b12de7889f8577cb20e9296547f7470bfaf686))
* **deps:** bump the go-dependencies group across 1 directory with 3 updates ([#95](https://github.com/Josh-Archer/terraform-provider-seerr/issues/95)) ([f7a54d8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f7a54d83cfbc4b0a86eb31b7b7b80fc9ae60250b))
* fix tofu CI - use dev_overrides and skip init for dev provider ([c39ace6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c39ace6e28462600ae079b5c5aec5fc33caf4581))
* fix tofu init in CI by setting job-level config env ([c47fc7a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c47fc7a7cf8d00f0c653e241758ee192d15dd085))
* fix tofu integration tests in CI using dev_overrides ([dbe6b04](https://github.com/Josh-Archer/terraform-provider-seerr/commit/dbe6b0407124acdd3c1effca8b8fa971f132f263))
* integrate Tofu tests into main CI and skip chores for tagging ([fa441e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa441e731d5e243c0c434e9a90d13180cac21c20))
* **main:** release 0.38.1 ([#189](https://github.com/Josh-Archer/terraform-provider-seerr/issues/189)) ([0122c97](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0122c97ba36c0c29090288738973ab921d65fb25))
* **main:** release 0.38.2 ([#194](https://github.com/Josh-Archer/terraform-provider-seerr/issues/194)) ([31e0531](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e05316e0c62160618443f184eff2bd2e04f531))
* **main:** release 0.38.3 ([#204](https://github.com/Josh-Archer/terraform-provider-seerr/issues/204)) ([5c270e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5c270e8f13b5847c6c5942ceb5b0fd30a0ea0b31))
* **main:** release 0.39.0 ([#208](https://github.com/Josh-Archer/terraform-provider-seerr/issues/208)) ([bcd4935](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bcd49352b4bad133b9d544cd826153a8083cf796))
* **main:** release 0.41.0 ([#224](https://github.com/Josh-Archer/terraform-provider-seerr/issues/224)) ([b10ea8c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b10ea8c83e0c072e9fb40f9529ae48a49aeb8613))
* **main:** release 0.42.0 ([#225](https://github.com/Josh-Archer/terraform-provider-seerr/issues/225)) ([d9aa9ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d9aa9adb743a3a26991f99aed948aedef6b6da8f))
* **main:** release 0.42.1 ([#232](https://github.com/Josh-Archer/terraform-provider-seerr/issues/232)) ([73a8c2b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/73a8c2ba8565c4aba4148ed292f7bd7d0cca22c6))
* **main:** release 0.42.2 ([#235](https://github.com/Josh-Archer/terraform-provider-seerr/issues/235)) ([8e6db2f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8e6db2fb14f3d0eef9329294332959646c023823))
* **main:** release 0.42.3 ([#239](https://github.com/Josh-Archer/terraform-provider-seerr/issues/239)) ([f39a71f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f39a71ffcb87e5e0ceb8c188f892856e1ff2a7bf))
* **main:** release 0.43.0 ([#242](https://github.com/Josh-Archer/terraform-provider-seerr/issues/242)) ([6b4c1b6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6b4c1b6bba41ba2d2f8d51967d5a6d0544bb2964))
* **main:** release 0.43.1 ([#244](https://github.com/Josh-Archer/terraform-provider-seerr/issues/244)) ([b82bdb2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b82bdb290d6a23137b42a83dfce602e5ff6c6d25))
* migrate to native skipping via default_bump: false in tag action ([d571bb6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d571bb632e3b34086cbe2b6f7f40b7f33ce7db63))
* **release:** bump version to 0.40.0 ([bf608f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bf608f56c9836c2ba43a244956e4d6f3016d784f))
* **release:** configure release-please for v1.0.0 post-major SemVer ([#246](https://github.com/Josh-Archer/terraform-provider-seerr/issues/246)) ([19603c7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/19603c722189f0d31654206c607471c1c476b09e))
* **release:** configure release-please to bump minor version for feat commits pre-1.0 ([#210](https://github.com/Josh-Archer/terraform-provider-seerr/issues/210)) ([aef9ca8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aef9ca88bc47c2ca91f2e9787ce78469184f17aa))
* remove research files ([732e4d1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/732e4d18d83cc2fe69e12a0364653615dc8e5016))
* replace with copyloopvar ([#253](https://github.com/Josh-Archer/terraform-provider-seerr/issues/253)) ([1db5214](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1db52140d3463d1360b812c1cf537bafbff2b8a5))
* trigger signed release after adding gpg secrets ([4851d6e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4851d6e28be6deb4a3609ebc2f98e8ba60d6fb9d))
* Update minimum Go version in module ([#282](https://github.com/Josh-Archer/terraform-provider-seerr/issues/282)) ([4cfdd3b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4cfdd3b4951c90e4759d993a1399a6cbc591c92c))
* use -plugin-dir for tofu init in CI with explicit version ([0ef218f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ef218f673ad85e19b8940745c05185981193604))
* use filesystem mirror for tofu init in CI ([b1d46c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b1d46c06fe5b448859f73b10134d2784e1a86ee6))


### Tests

* align Jellyfin fixture authentication ([#240](https://github.com/Josh-Archer/terraform-provider-seerr/issues/240)) ([a32f3ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a32f3ad64abcc84e0ee165c6e41a7f207c00a9f9))
* **arch:** add Single Egress Rule architecture test and isolate test network fixtures ([#234](https://github.com/Josh-Archer/terraform-provider-seerr/issues/234)) ([ab1841e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ab1841ecac85c517d88a75e3f607e95814d55038))
* implement comprehensive OpenTofu integration tests for all provider features ([cbdb3c3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cbdb3c33e6610ed9bbe74f38712dd21e8113b5d6))


### Build System

* **deps:** bump github.com/cloudflare/circl from 1.6.1 to 1.6.3 ([#346](https://github.com/Josh-Archer/terraform-provider-seerr/issues/346)) ([3277b3c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3277b3cf37b87da164869f8b2d254241a2a51782))
* **deps:** bump github.com/cloudflare/circl in /tools ([#347](https://github.com/Josh-Archer/terraform-provider-seerr/issues/347)) ([a15d93e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a15d93e43295e34ad94ac67cd8a8bba0481de918))
* **deps:** bump github.com/hashicorp/terraform-plugin-framework ([#348](https://github.com/Josh-Archer/terraform-provider-seerr/issues/348)) ([c681e7e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c681e7ea74469ede4e0c90b7f1f11f57e00f798f))
* **deps:** bump github.com/hashicorp/terraform-plugin-go ([#345](https://github.com/Josh-Archer/terraform-provider-seerr/issues/345)) ([ba92b67](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ba92b67ebc725872123b7e1dfa3e58c80bdba20c))
* **deps:** bump the github-actions group across 1 directory with 3 updates ([#349](https://github.com/Josh-Archer/terraform-provider-seerr/issues/349)) ([3957103](https://github.com/Josh-Archer/terraform-provider-seerr/commit/395710343f1240a8a4a6af99bbc10b2383e5fc12))


### Continuous Integration

* add `golangci-lint` + fix lints ([#120](https://github.com/Josh-Archer/terraform-provider-seerr/issues/120)) ([fa8056c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa8056ce94248bf396295c0b208553f655782e91))
* Add GitHub Actions CI workflow for build, test, linting, and automated version tagging. ([f23eaa1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f23eaa1be4b62bb8033704635703ab98c0d33ad5))
* add hosted-runner fallback and remove template dependabot config ([aa59965](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aa5996555a7fabd0f71985ab8fa1709c2edd68d4))
* add OpenAPI coverage check step to CI workflow ([#147](https://github.com/Josh-Archer/terraform-provider-seerr/issues/147)) ([82a753f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/82a753f32123e20d9156da78892e8b72ecbcafaa))
* auto-tag main with semver after successful CI ([cb4faee](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cb4faeeda6c651e164a863a45cc21edee72cc4d3))
* avoid secrets in if expressions for release workflow ([cfb0351](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cfb0351d2a1c627d59195b7e4ca2bbf5cd9732a9))
* disable golangci-lint remote schema verification and add auto-merge for owner PRs ([#202](https://github.com/Josh-Archer/terraform-provider-seerr/issues/202)) ([76be0d0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/76be0d0c7d8e17267689bc8611a12c06f7731dbd))
* dispatch release after auto-tag ([eaccb1a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eaccb1ad26630d3ca67c955fee7046ff89daf0ac))
* dispatch Release after auto-tag on main ([#91](https://github.com/Josh-Archer/terraform-provider-seerr/issues/91)) ([aae4f46](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aae4f46a48d04b05180904d9276ffaa88f339eec))
* integration tests ([9c00c20](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9c00c205680a9a75c471215115269472d16250f7))
* remove legacy auto-tag from test.yml in favor of staged release-please ([#211](https://github.com/Josh-Archer/terraform-provider-seerr/issues/211)) ([b8a2e50](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b8a2e505b2673ddc18754238ceb6f8036564054b))
* replace custom PowerShell semver with mathieudutour/github-tag-action and add Copilot commit instructions ([1724cda](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1724cda5798c287fe72a253ac1375d230d495545))
* replace forked GH action with latest upstream ([#55](https://github.com/Josh-Archer/terraform-provider-seerr/issues/55)) ([3a91275](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3a91275abc3c45c61c888142c64c9d5abef588b2))
* support unsigned release fallback when GPG secrets are unset ([b674265](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b674265f4a0b4dcc178444e1ff907db45dd33be4))

## [0.43.1](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.43.0...v0.43.1) (2026-08-30)


### Documentation

* formalize v1.0.0 stability and breaking-change policy ([#243](https://github.com/Josh-Archer/terraform-provider-seerr/issues/243)) ([a1d3b73](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a1d3b73fae7d761cf6a0d99793d24b7021d7c10b))


### Miscellaneous Chores

* **main:** release 0.42.3 ([#239](https://github.com/Josh-Archer/terraform-provider-seerr/issues/239)) ([f39a71f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f39a71ffcb87e5e0ceb8c188f892856e1ff2a7bf))
* **main:** release 0.43.0 ([#242](https://github.com/Josh-Archer/terraform-provider-seerr/issues/242)) ([6b4c1b6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6b4c1b6bba41ba2d2f8d51967d5a6d0544bb2964))

## [0.43.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.42.3...v0.43.0) (2026-08-30)


### Features

* add `seerr_user` resource for managing users and their notification settings. ([453389c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/453389c90af4234e94203504185818fea51058ec))
* add email and pushbullet settings data sources ([#150](https://github.com/Josh-Archer/terraform-provider-seerr/issues/150)) ([80a076c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80a076ca3826057dd36329b4e3cb4df8a5dce83b))
* Add media requests, issues, and service status features ([#70](https://github.com/Josh-Archer/terraform-provider-seerr/issues/70)) ([cfce309](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cfce30977698c73e9c00fc49a5389033e82d8273))
* add media requests, issues, and service status resources ([fede0b2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fede0b2141ac585ce1565a7023c39e04c81df51d))
* add reusable module ecosystem ([#228](https://github.com/Josh-Archer/terraform-provider-seerr/issues/228)) ([99cdee0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/99cdee04530ab63296736e06d4d3a6e3719cfd98))
* add Seerr modules for ARR servers and notifications ([f987801](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f987801c56f234aac22454e24598ad1b63dcb929))
* Add Seerr user resource and data source with support for Plex import and notification settings management. ([0348176](https://github.com/Josh-Archer/terraform-provider-seerr/commit/034817666479e3313496e08814c52c9977294e7f))
* add seerr_emby_settings resource and data source ([#48](https://github.com/Josh-Archer/terraform-provider-seerr/issues/48)) ([81e5b20](https://github.com/Josh-Archer/terraform-provider-seerr/commit/81e5b20634ae00e5dd8906dbcf9d30dff4030b5a))
* add seerr_issues and seerr_requests data sources ([#45](https://github.com/Josh-Archer/terraform-provider-seerr/issues/45)) ([28f0bc4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/28f0bc4b9e1c1a20e6d442261f5349c0dc69d81c)), closes [#44](https://github.com/Josh-Archer/terraform-provider-seerr/issues/44)
* add seerr_permission_set data source ([#112](https://github.com/Josh-Archer/terraform-provider-seerr/issues/112)) ([ac82806](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ac828064ee9c45784208226527f6a5d9f3583238))
* add Terraform resource and data source for Seerr main settings. ([64470f4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/64470f471cf3d930f9c36d860fa41d4c6eeea7d3))
* add Terraform resource and data source for Seerr main settings. ([bcb847f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bcb847f1a1988441898e1d892c5dee364c3ebfe3))
* add Tier 3 TMDB reference data sources ([#159](https://github.com/Josh-Archer/terraform-provider-seerr/issues/159)) ([#175](https://github.com/Josh-Archer/terraform-provider-seerr/issues/175)) ([eb08106](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eb08106a558491cc299cf93edb9be23c98db498b))
* add Tofu test workflows for CI and agent ([f9b35a6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f9b35a6d2d4e03e1b9d20767823cc5d2d3bfac82))
* brownfield adoption ([2af4a9e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2af4a9e9296856d52fe12ad5248787d271bcf787))
* bulk import and migration CLI tooling with HCL generator and migration guide ([#170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170)) ([#203](https://github.com/Josh-Archer/terraform-provider-seerr/issues/203)) ([37a44e6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/37a44e679ca29211b69e039bcb286f83eeb25bad))
* centralize provider resource and data-source registration metadata ([#113](https://github.com/Josh-Archer/terraform-provider-seerr/issues/113)) ([2fff552](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2fff55225dd12359442dd6ed839cfab303b34ee3))
* **ci:** add RC prerelease workflow, GoReleaser snapshot check, and align with chaptarr ([#214](https://github.com/Josh-Archer/terraform-provider-seerr/issues/214)) ([e64f2f2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e64f2f299abba9d86b577c3193bb963cd782568c))
* **ci:** release automation with release-please, tag-on-merge, and registry verification ([#181](https://github.com/Josh-Archer/terraform-provider-seerr/issues/181)) ([#187](https://github.com/Josh-Archer/terraform-provider-seerr/issues/187)) ([34fb8c6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/34fb8c6e89d7d90f1945bbaa9815c925dbaf4cb3))
* community readiness - issue and PR templates, devcontainer, contributing guide, and governance ([#183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183)) ([#201](https://github.com/Josh-Archer/terraform-provider-seerr/issues/201)) ([83a2b10](https://github.com/Josh-Archer/terraform-provider-seerr/commit/83a2b1036393351c3587246974566a2545baba16))
* **compatibility:** verify current upstream releases ([#223](https://github.com/Josh-Archer/terraform-provider-seerr/issues/223)) ([bfb3e49](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bfb3e495cb90f9bccbf75d108231949c7bdf9c09))
* Developer experience - env vars, retry/backoff, filtered data sources, and arr resolvers ([#164](https://github.com/Josh-Archer/terraform-provider-seerr/issues/164)) ([#179](https://github.com/Josh-Archer/terraform-provider-seerr/issues/179)) ([24c153b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/24c153b9b4b5f38d62bec462e3891be687849c98))
* document and support plex-token bootstrap configuration ([#115](https://github.com/Josh-Archer/terraform-provider-seerr/issues/115)) ([9997af8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9997af8d1f8c33d2107dcda5b9fd029b85cee6c7))
* Emby library enable lists and sync actions (closes [#135](https://github.com/Josh-Archer/terraform-provider-seerr/issues/135)) ([#144](https://github.com/Josh-Archer/terraform-provider-seerr/issues/144)) ([6345ad0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6345ad09e662608917d41c6dfb7355a0425bf707))
* expand jellyfin settings and library parity ([#149](https://github.com/Josh-Archer/terraform-provider-seerr/issues/149)) ([fe9960b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fe9960bef2956afab8bc7b2eb420e7aace608a9f))
* expand network and tautulli settings data sources and unit tests ([#148](https://github.com/Josh-Archer/terraform-provider-seerr/issues/148)) ([0d3bec4](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0d3bec4850c31fbe04af04494a1b7397a9a71385))
* HTTP client rate limiting and 429/Retry-After backoff (closes [#134](https://github.com/Josh-Archer/terraform-provider-seerr/issues/134)) ([#143](https://github.com/Josh-Archer/terraform-provider-seerr/issues/143)) ([0e86030](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0e86030fd0015c6c76cdf3c692d41344806d918d))
* implement first-class typed plex settings (Issue [#6](https://github.com/Josh-Archer/terraform-provider-seerr/issues/6)) ([833ee78](https://github.com/Josh-Archer/terraform-provider-seerr/commit/833ee7862f4e1ede74a77c93e4220af509c33cc5))
* implement first-class typed plex settings (Issue [#6](https://github.com/Josh-Archer/terraform-provider-seerr/issues/6)) ([e2ccbc6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e2ccbc6be4c6cbbe4151f7254730f65cbfc759b4))
* implement initial Seerr Terraform provider with API key resource and data sources for main settings and API key. ([cd9dd34](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cd9dd34a7505e5d96600a3e0dde36b8e22070f70))
* implement jellyfin settings resource and data source ([1f37784](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1f37784875fafd44541893689d77a60838ae3ace))
* implement missing notification types and events (Issue [#18](https://github.com/Josh-Archer/terraform-provider-seerr/issues/18)) ([707f0ac](https://github.com/Josh-Archer/terraform-provider-seerr/commit/707f0ac0e3a8b103fe528d95216531c59bdd4242))
* implement OpenTofu Seerr provider with release automation ([241a406](https://github.com/Josh-Archer/terraform-provider-seerr/commit/241a4060f10b07a8b9e226b48a667c425f467391))
* implement per-user notification settings resource and data source ([#151](https://github.com/Josh-Archer/terraform-provider-seerr/issues/151)) ([9d89088](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9d8908840d0ee2406a7f0b45bc4f19f7de947d00))
* implement seerr_job_schedule resource ([#28](https://github.com/Josh-Archer/terraform-provider-seerr/issues/28)) ([f9a19c2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f9a19c27722114099c66468705dc5fd3affe3b45))
* implement tautulli settings and fix notification agent registration ([f23ebfa](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f23ebfa2b9c1c5e9d623aa26459b9eb4b6919f04))
* job run/cancel and notification agent test actions (closes [#123](https://github.com/Josh-Archer/terraform-provider-seerr/issues/123)) ([#141](https://github.com/Josh-Archer/terraform-provider-seerr/issues/141)) ([5addcdd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5addcddc5f98158690ced4ee4040121772af1c83))
* **library:** add Plex and Jellyfin library settings and sync resources ([#121](https://github.com/Josh-Archer/terraform-provider-seerr/issues/121)) ([#131](https://github.com/Josh-Archer/terraform-provider-seerr/issues/131)) ([6b739d5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6b739d53b91d5f2b830cac6cf0f5db54c8a6f52d))
* **observability:** add dashboards, drift detection, and recovery runbooks ([#184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184)) ([5673959](https://github.com/Josh-Archer/terraform-provider-seerr/commit/56739599075f588067667e44c4fbc8b77a931610))
* **openapi:** OpenAPI coverage matrix and CI drift check ([#119](https://github.com/Josh-Archer/terraform-provider-seerr/issues/119)) ([#128](https://github.com/Josh-Archer/terraform-provider-seerr/issues/128)) ([d737ca5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d737ca515daa70e7b6c1eabcaa4ad016db854f03))
* **permissions:** generate PermissionsMap from seerr_permissions.ts ([#109](https://github.com/Josh-Archer/terraform-provider-seerr/issues/109)) ([#129](https://github.com/Josh-Archer/terraform-provider-seerr/issues/129)) ([b62e5f8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b62e5f8ee1999ba083d65da4aa977d29a145eb90))
* Phase 6 - Advanced Resource Lifecycle (request approvals, issue comments, computed attributes, and expanded override rules) ([#177](https://github.com/Josh-Archer/terraform-provider-seerr/issues/177)) ([bb95271](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bb9527108a52a7cc58285b0b91675ca5202e4c7d))
* Phase 8 - Production Readiness (integration tests, changelog, composite example) ([#165](https://github.com/Josh-Archer/terraform-provider-seerr/issues/165)) ([#186](https://github.com/Josh-Archer/terraform-provider-seerr/issues/186)) ([ac32f5a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ac32f5aca2f1737671f224ff48c96f3e2743b518))
* Plex Provider Support ([#41](https://github.com/Josh-Archer/terraform-provider-seerr/issues/41)) ([c2605ec](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c2605ec034a5adc94040bb7e73de3b9134c5f77f))
* provider ergonomics and validation ([#58](https://github.com/Josh-Archer/terraform-provider-seerr/issues/58)) ([ddb253d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ddb253d5dddf9b045bb4f6714cf031e8fb0f08c6))
* provider hardening - import support, validation, defaults, plan modifiers ([#162](https://github.com/Josh-Archer/terraform-provider-seerr/issues/162)) ([#176](https://github.com/Josh-Archer/terraform-provider-seerr/issues/176)) ([97261b1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/97261b1ba6fd92bca9076830b5f08d3ce4bb2288))
* replace custom semver script with mathieudutour/github-tag-action and add Copilot commit instructions ([09375d9](https://github.com/Josh-Archer/terraform-provider-seerr/commit/09375d9bccf9eeb30e3fbadd39cf3fa01f526150))
* **request:** poll async status after create/update ([#90](https://github.com/Josh-Archer/terraform-provider-seerr/issues/90)) ([6be397f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6be397f723fd3ad8d38ad233fc5f9a9d86832eb3))
* **resilience:** graceful 404 state removal and schema resilience ([#169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169)) ([#209](https://github.com/Josh-Archer/terraform-provider-seerr/issues/209)) ([3f26fd6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3f26fd685bd64d0c2598b7ea726983ed5a7ae89d))
* strongly-typed notification agents ([#9](https://github.com/Josh-Archer/terraform-provider-seerr/issues/9)) ([80211e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80211e749d090ff9484f8bc10c6a9b29db035d46))
* strongly-typed notification agents ([#9](https://github.com/Josh-Archer/terraform-provider-seerr/issues/9)) ([65a8f1a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/65a8f1a20bf5ffc4b64ed7c43e9639350761ed3b))
* upstream compatibility automation - scheduled OpenAPI diff, release watcher, and compatibility matrix ([#182](https://github.com/Josh-Archer/terraform-provider-seerr/issues/182)) ([#195](https://github.com/Josh-Archer/terraform-provider-seerr/issues/195)) ([62a87a3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a87a3e1b80d5bcffe6acfb5b87ce3d0d5a96a6))
* **upstream:** sync OpenAPI spec with upstream develop and update coverage ([#197](https://github.com/Josh-Archer/terraform-provider-seerr/issues/197)) ([#212](https://github.com/Josh-Archer/terraform-provider-seerr/issues/212)) ([f35668f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f35668f8805e7cab303d20522c98ce3c911f7369))
* User Settings Permissions and State Fixes ([#38](https://github.com/Josh-Archer/terraform-provider-seerr/issues/38)) ([025b3c7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/025b3c742056f5eb2f4b1bd5e12c9f1360602531))
* **user-import:** add Plex and Jellyfin user import resources and data sources ([#124](https://github.com/Josh-Archer/terraform-provider-seerr/issues/124)) ([#133](https://github.com/Josh-Archer/terraform-provider-seerr/issues/133)) ([8b979b3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b979b398dc9b40678bdb4ac52f38aa0dccd26f5))
* **user-quota:** add seerr_user_quota resource and data source ([#122](https://github.com/Josh-Archer/terraform-provider-seerr/issues/122)) ([#132](https://github.com/Josh-Archer/terraform-provider-seerr/issues/132)) ([cc63bc0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc63bc0684f80dbe031cec6611f6ab36e6ee6bfa))
* **v0.31.0:** phase 1 stability, unit tests, HCL examples, and CI workflow concurrency ([#166](https://github.com/Josh-Archer/terraform-provider-seerr/issues/166)) ([60687c5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/60687c503e85dab63eb68711f666acf216cdedeb))
* **v0.32.0:** Phase 2 - Observability and bootstrapping data sources ([#172](https://github.com/Josh-Archer/terraform-provider-seerr/issues/172)) ([9aa1503](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9aa1503faf5409bd0b27913b8d9a6fc6da782547))
* **v0.33.0:** add Tier 2 user lookup data sources and watchlist resource ([#158](https://github.com/Josh-Archer/terraform-provider-seerr/issues/158)) ([#173](https://github.com/Josh-Archer/terraform-provider-seerr/issues/173)) ([aad1555](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aad15557dbf706441e117ea83671eb16b2b42442))
* validate published modules and module examples in CI ([#114](https://github.com/Josh-Archer/terraform-provider-seerr/issues/114)) ([4a8fa86](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4a8fa86c8dc026567e76ffa7f3ad585fe71f0130))


### Bug Fixes

* Add `seerr_public_settings` data source and `seerr_main_settings` resource and data source with corresponding documentation, examples, and tests. ([dbfe08f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/dbfe08f3faedf0a120537c604a8ed66867031d5a))
* Add `tfplugindocs` integration and generated documentation for Seerr resources and data sources. ([9822c5b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9822c5bcf54ef2233198a373ee2369f27c300763))
* add GitHub Actions workflow for CI, OpenTofu integration tests, and automatic tagging. ([62a0ab5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/62a0ab5385be837b6b46a2775f937b7eb818f355))
* Add Radarr and Sonarr server resources with lifecycle tests to t… ([#26](https://github.com/Josh-Archer/terraform-provider-seerr/issues/26)) ([fa38980](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa3898061c625784af52277bdcfe7eae062330a7))
* add Terraform resources for Sonarr and Radarr servers and data sources for their quality profiles. ([7263fea](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7263fea8f65ce9ad5b292275278406f7df1cf109))
* **api_key:** add id attribute to seerr_api_key schema and align import state ([#206](https://github.com/Josh-Archer/terraform-provider-seerr/issues/206)) ([163f4be](https://github.com/Josh-Archer/terraform-provider-seerr/commit/163f4bed58a34229fbfa9031ca816fcdfa6d7734))
* arr scans variable ([7a72022](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7a720229495ac5139be8f51ae75bee1f2e5fc878))
* arr scans variable ([faf4c41](https://github.com/Josh-Archer/terraform-provider-seerr/commit/faf4c417b8f12af18934b8ac8873c18d7420eec8))
* **ci:** correct release-please-action commit sha in release-please.yml ([#188](https://github.com/Josh-Archer/terraform-provider-seerr/issues/188)) ([03e1e35](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e1e354dd87e7d575553e9861c46cfda8d0497c))
* **ci:** create tags for draft releases ([#226](https://github.com/Josh-Archer/terraform-provider-seerr/issues/226)) ([18ed2de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/18ed2de8764b6c8eea77a73f46194203ea696cde))
* **ci:** fetch tags after remote creation so GoReleaser can validate HEAD ([#215](https://github.com/Josh-Archer/terraform-provider-seerr/issues/215)) ([04a4105](https://github.com/Josh-Archer/terraform-provider-seerr/commit/04a41057786327b9130607e70b72ef899f58879d))
* **ci:** harden release reconciliation to discover draft releases and self-heal ([#236](https://github.com/Josh-Archer/terraform-provider-seerr/issues/236)) ([63d76fc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/63d76fc34616ca4cf2157d1cfb7cd178ee005c15))
* **ci:** pass explicit tag to goreleaser ([#227](https://github.com/Josh-Archer/terraform-provider-seerr/issues/227)) ([9633ada](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9633adae29b4d881be99692a9864b71a44bef107))
* **ci:** query the OpenTofu provider endpoint ([#229](https://github.com/Josh-Archer/terraform-provider-seerr/issues/229)) ([6095c2d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6095c2db2e0f50339e6b28b6f48062b78042b250))
* **ci:** remove environment gate from release.yml for zero-click publishing ([#198](https://github.com/Josh-Archer/terraform-provider-seerr/issues/198)) ([26cbf40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/26cbf4008be4895afe2780007ad522506f6eacbd))
* **ci:** resolve OpenTofu test index evaluation and improve upstream issue deduplication ([#221](https://github.com/Josh-Archer/terraform-provider-seerr/issues/221)) ([b088c6d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b088c6d3b22e80bd431b74a5d483880f03f01a11))
* **ci:** set draft: true for release-please to allow GoReleaser publishing ([#193](https://github.com/Josh-Archer/terraform-provider-seerr/issues/193)) ([f4d86ae](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f4d86ae67231451ac43c6a4ce2bdfc53af3414e0))
* clean up files ([b66a0f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b66a0f525b6824a7ec4334e5d4c8f660ad6dbd00))
* **client:** default empty POST/PUT/PATCH payload to {} and set Content-Type header ([#146](https://github.com/Josh-Archer/terraform-provider-seerr/issues/146)) ([782cbf7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/782cbf76d3cee6e5130f35974dd2355e1292bc54))
* detect feat/ and feature/ branches in auto-tag minor version bump ([705a184](https://github.com/Josh-Archer/terraform-provider-seerr/commit/705a18405644c1e7d2e07a5cb807282ace58cc4c))
* edge casing ([737b796](https://github.com/Josh-Archer/terraform-provider-seerr/commit/737b796dd5dc2f419e3dc3245cfd7b60c9a8e369))
* finalize seerr_job_schedule and seerr_discover_slider, update n… ([#29](https://github.com/Josh-Archer/terraform-provider-seerr/issues/29)) ([b42ac13](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b42ac137ae3cd25208930e852607b4166d3586c7))
* issue 33 request timeout ([#57](https://github.com/Josh-Archer/terraform-provider-seerr/issues/57)) ([756ea22](https://github.com/Josh-Archer/terraform-provider-seerr/commit/756ea220c70a3bbd6e745d777b5b9cdff0efcfe7))
* **issue:** fail status updates when Seerr returns HTTP errors ([#87](https://github.com/Josh-Archer/terraform-provider-seerr/issues/87)) ([2472d8c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2472d8cc14b068a9a98b37e55220a49aa283c5c1)), closes [#83](https://github.com/Josh-Archer/terraform-provider-seerr/issues/83)
* keep discover slider resource when managed list is empty ([#86](https://github.com/Josh-Archer/terraform-provider-seerr/issues/86)) ([2e263a6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2e263a6f2cce78e8b1da20481104d679d590279e)), closes [#84](https://github.com/Josh-Archer/terraform-provider-seerr/issues/84)
* **library_settings:** normalize empty enabled_libraries slice to non-nil empty set ([#142](https://github.com/Josh-Archer/terraform-provider-seerr/issues/142)) ([ad49fb6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ad49fb6227ee6d5bb3a7b19f145308386cafd68e))
* **library_settings:** normalize singleton ID during import to prevent state mutation drift ([#207](https://github.com/Josh-Archer/terraform-provider-seerr/issues/207)) ([b6e55bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b6e55bc3827714a1104dd9551a8497bbd4f12e3f))
* **lint:** resolve golangci-lint issues to pass CI ([#156](https://github.com/Josh-Archer/terraform-provider-seerr/issues/156)) ([cda0343](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cda0343d378e7b1bf67da5414cfa76f8eb12b64c))
* normalize discover sliders and verify notification deletes ([#63](https://github.com/Josh-Archer/terraform-provider-seerr/issues/63)) ([5a6c6c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5a6c6c06c27f47be6506160c3357975cb5a14c9e))
* notification state handling and bootstrap ephemeral Seerr CI ([#40](https://github.com/Josh-Archer/terraform-provider-seerr/issues/40)) ([f6f2def](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f6f2defe89c57ec2b8cb69cf5b884c6713d93285))
* **notifications:** ensure optional notification attributes are Computed and option keys only sent when set ([#120](https://github.com/Josh-Archer/terraform-provider-seerr/issues/120)) ([#125](https://github.com/Josh-Archer/terraform-provider-seerr/issues/125)) ([edaa6ec](https://github.com/Josh-Archer/terraform-provider-seerr/commit/edaa6ec46323bae30fdf8e8f8f459340210425ee))
* **pagination:** ensure collection data sources complete full pagination without under-paging (closes [#137](https://github.com/Josh-Archer/terraform-provider-seerr/issues/137)) ([#145](https://github.com/Josh-Archer/terraform-provider-seerr/issues/145)) ([8124979](https://github.com/Josh-Archer/terraform-provider-seerr/commit/81249790ac97f6a63589b4d44b4ba745161d3016))
* **pagination:** safeIntFromAny direct int bounds check to fix CodeQL integer conversion alerts ([#155](https://github.com/Josh-Archer/terraform-provider-seerr/issues/155)) ([741b5bc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/741b5bcdb8cc7e101d094f73c6475e4b93d76bd5))
* **pagination:** safely bound int64 to int conversions to fix CodeQL alerts ([#154](https://github.com/Josh-Archer/terraform-provider-seerr/issues/154)) ([d3947d3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d3947d3008a635e29d626a6e9321aa84db17cb99))
* panic on empty provider configuration in tests ([efd5762](https://github.com/Josh-Archer/terraform-provider-seerr/commit/efd5762d9f1a7b2c2dc7158d6e1056e938928dbc))
* **plex_library_settings:** omit empty enable query param when enabling zero libraries ([#140](https://github.com/Josh-Archer/terraform-provider-seerr/issues/140)) ([0041db6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0041db60148c90259e670be5fb11fc37dffa721a))
* preserve Seerr base URL subpaths when building API paths ([#85](https://github.com/Josh-Archer/terraform-provider-seerr/issues/85)) ([595c4c2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/595c4c2a47bbad57938216e8a1446d41fda7f5d4)), closes [#82](https://github.com/Josh-Archer/terraform-provider-seerr/issues/82)
* Preserve Seerr server IDs on update ([#25](https://github.com/Josh-Archer/terraform-provider-seerr/issues/25)) ([7ef6f12](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7ef6f12a7f2a6f5d70a9186098fa4a3db7523ca7))
* preserve user data source casing ([#110](https://github.com/Josh-Archer/terraform-provider-seerr/issues/110)) ([7e22892](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7e22892b30fb7ac580f9fda342b545225c75259b))
* prevent discover slider drift on read/delete ([#62](https://github.com/Josh-Archer/terraform-provider-seerr/issues/62)) ([11aa45c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/11aa45cdfe074c1932320f6d7b888bd1dacddf3f))
* Propagate diagnostics correctly from seerr_api_key Update ([#111](https://github.com/Josh-Archer/terraform-provider-seerr/issues/111)) ([6c5d26f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6c5d26ffdacbfdf63a7973e66ab3b1b2a944679e))
* quality profile issue with resources and update docs ([f8f0cbb](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f8f0cbbbc9310855b6b38cff0922fd9d69651b66))
* remove stale notification types state path ([9b76acd](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9b76acd990ab5fdfff8173403f0f5b6dc45e4a97))
* resolve CodeQL incorrect integer conversion in provider env parsing ([#213](https://github.com/Josh-Archer/terraform-provider-seerr/issues/213)) ([35a5c95](https://github.com/Josh-Archer/terraform-provider-seerr/commit/35a5c95b644fbc4e4a60cac5b828c85caced3fce))
* resolve edge cases in retry config, empty arr collections, and synthetic IDs ([#180](https://github.com/Josh-Archer/terraform-provider-seerr/issues/180)) ([ef1d617](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ef1d617d936f68ecf05bcd184716daee0d94d259))
* resolve unknown value after apply and edge cases ([53ba55f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/53ba55fcef11d075d3c0c9da3fa6222c971562a6))
* resources normalize the model in a local copy inside payload ([80d2369](https://github.com/Josh-Archer/terraform-provider-seerr/commit/80d2369530092872a90272aedabe121e0b111421))
* resources normalize the model in a local copy inside payload ([6f05948](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6f0594809f92382ead735f09e7de2d3ce96c86ac))
* restore release and Jellyfin compatibility gates ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7cb5743](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7cb57430b9fbab52b72ce7aa7f3852729fb3d291))
* satisfy lint rule for safe transport cloning ([23c51a9](https://github.com/Josh-Archer/terraform-provider-seerr/commit/23c51a9a587d8813b592f51d8e1e79d66c4df47d))
* schema exact match ([#60](https://github.com/Josh-Archer/terraform-provider-seerr/issues/60)) ([c4c7186](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c4c71865ac553e3690ec6bad9a1eb6c7b8635831))
* **servarr:** validate Arr connectivity via Seerr server proxy test endpoint ([#231](https://github.com/Josh-Archer/terraform-provider-seerr/issues/231)) ([2d128c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2d128c089885ba5bab8033aa516f585ea0767073))
* type radarr and sonarr server schemas ([#59](https://github.com/Josh-Archer/terraform-provider-seerr/issues/59)) ([c5c2aad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c5c2aadc882b247e49c2157c29995a16899d4138))
* typed notification resources/data sources no longer deserialize plan/state through the old generic schema shape ([#36](https://github.com/Josh-Archer/terraform-provider-seerr/issues/36)) ([a5b8e69](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a5b8e69be497e8da20501e648fdb0b1d3b049d51))
* validate examples and schema contracts ([#168](https://github.com/Josh-Archer/terraform-provider-seerr/issues/168)) ([f826b40](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f826b40e79004a00fc9f50f9b4d16e2ac4f71cb3))
* versioning ([a0b2510](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a0b251018fc3f5563ee30efd6c4fce9da904d592))


### Documentation

* add HCL example snippets for v0.30.0 data sources (Closes [#152](https://github.com/Josh-Archer/terraform-provider-seerr/issues/152)) ([#153](https://github.com/Josh-Archer/terraform-provider-seerr/issues/153)) ([7bf2965](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7bf296549dd8007af2425f26548e07f24612ad41))
* Add import example ([#306](https://github.com/Josh-Archer/terraform-provider-seerr/issues/306)) ([f6b644e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f6b644efe282c406c3d1b147478e572f99880f28))
* add ROADMAP.md and update progress matrix ([#199](https://github.com/Josh-Archer/terraform-provider-seerr/issues/199)) ([ff71232](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff71232bdc34a18f90e470332887f15c118f298e))
* **auto:** regenerate provider documentation ([03e3637](https://github.com/Josh-Archer/terraform-provider-seerr/commit/03e3637461d00c476e9e85f325b2de2bde6d5040))
* **auto:** regenerate provider documentation ([b277888](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b27788883f60b94e1988bc0fad3c25d42f26641c))
* **auto:** regenerate provider documentation ([5a242e2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5a242e29e8916422003e1e5b09f00afce546cb80))
* **auto:** regenerate provider documentation ([fa4ebd7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa4ebd7f7d2d4bfd8e33a4107f931bfef2ee839a))
* **auto:** regenerate provider documentation ([965f19e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/965f19e691977ab84ab001fbb156b77d010cf25d))
* **auto:** regenerate provider documentation ([2f27fbb](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2f27fbbcdc317a7adac5acfecaf4708007cf522d))
* **auto:** regenerate provider documentation ([e3abd18](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3abd185d91b6c574685838211cc551cef6da47a))
* **auto:** regenerate provider documentation ([a3ed5e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a3ed5e84a115d89c7562cbb87f797f3f75b722fa))
* **auto:** regenerate provider documentation ([3931be8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3931be81661d9cb6ebc0c79cd865ca4ddf972270))
* **auto:** regenerate provider documentation ([cb264e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cb264e717c79006195a76fc32927f586bc0f6e13))
* **auto:** regenerate provider documentation ([2f68a3c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2f68a3c6e58372e898c00afa8f50fe37f106c16e))
* clarify reusable module usage ([#230](https://github.com/Josh-Archer/terraform-provider-seerr/issues/230)) ([31e1435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e1435968cc8e3ae592c8124c9db1860ebb4053))
* consolidate roadmap into README ([#237](https://github.com/Josh-Archer/terraform-provider-seerr/issues/237)) ([cc54970](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc54970d7e883cfda0580660f1e559115dd715e7))
* consolidate roadmap phases, wording, and dual Terraform/OpenTofu support ([#200](https://github.com/Josh-Archer/terraform-provider-seerr/issues/200)) ([e340501](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e3405010ae94bdee49ba439d7379c28d4f42e5f2))
* enforce generated docs checks via pre-push + workflow ([#34](https://github.com/Josh-Archer/terraform-provider-seerr/issues/34)) ([ee026e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ee026e8eea94fe77d4cf98bf321b3b6dfc1a9f18))
* generate documentation for new resources ([55c69e2](https://github.com/Josh-Archer/terraform-provider-seerr/commit/55c69e29716672d1e44c8b3964371349039a19cf))
* generate documentation for new resources and data sources ([15402f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/15402f51726fb705f0aa08fdb39549a76d59582a))
* mark v0.43.0 module release stable ([#241](https://github.com/Josh-Archer/terraform-provider-seerr/issues/241)) ([0ac711f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ac711f63685c962c0047d69190e2b795662357d))
* modernize README and docs for accuracy and conciseness ([#192](https://github.com/Josh-Archer/terraform-provider-seerr/issues/192)) ([8b4cc9b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8b4cc9b7c3b8a0e4c0b2aaa001bd96b3f6a195a7))


### Miscellaneous Chores

* add agent instructions for Git workflow. ([#27](https://github.com/Josh-Archer/terraform-provider-seerr/issues/27)) ([6770a65](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6770a65d4754e21de0c11dd4b45f906008cc23db))
* **agents:** add GPG sign bypass instructions to repo workflow ([#116](https://github.com/Josh-Archer/terraform-provider-seerr/issues/116)) ([ff0a74b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ff0a74bf05120c4cec901228b96c70df7eb78540))
* **agents:** configure subagents to sign commits using private key directly ([#117](https://github.com/Josh-Archer/terraform-provider-seerr/issues/117)) ([c179d8e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c179d8ee1fce8637d15c7baa87622ce31e7f8d34))
* **agents:** use global GPG/SSH key settings for signing commits ([#118](https://github.com/Josh-Archer/terraform-provider-seerr/issues/118)) ([d75d380](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d75d380202767106ca806088e61fcb6cde7aa9a6))
* change to `govet` ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7ad0b83](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7ad0b835d499d9929da0b28a47e653add79ac319))
* **deps:** bump github.com/stretchr/testify ([#126](https://github.com/Josh-Archer/terraform-provider-seerr/issues/126)) ([ccd5e15](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ccd5e1565fb7c430135fac65141786ceec43c954))
* **deps:** bump github.com/stretchr/testify ([#190](https://github.com/Josh-Archer/terraform-provider-seerr/issues/190)) ([088bd87](https://github.com/Josh-Archer/terraform-provider-seerr/commit/088bd87586465808bdf8ecd800639c0b94a65a94))
* **deps:** bump golang.org/x/crypto from 0.49.0 to 0.52.0 ([#92](https://github.com/Josh-Archer/terraform-provider-seerr/issues/92)) ([4eb4435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4eb4435eba68aa7f1b443a54354fc7a7524107bd))
* **deps:** bump golang.org/x/net from 0.54.0 to 0.55.0 ([#98](https://github.com/Josh-Archer/terraform-provider-seerr/issues/98)) ([53cc295](https://github.com/Josh-Archer/terraform-provider-seerr/commit/53cc2956012a457b31161d2e7f07f2df05593e22))
* **deps:** bump google.golang.org/grpc from 1.79.3 to 1.82.1 ([#130](https://github.com/Josh-Archer/terraform-provider-seerr/issues/130)) ([8e1053e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8e1053ead755c8bedd457be59d02d734de3d0ee8))
* **deps:** bump the github-actions group across 1 directory with 10 updates ([#96](https://github.com/Josh-Archer/terraform-provider-seerr/issues/96)) ([e8aee52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/e8aee52caa09bfe7c22c065d99a3166da744a5cf))
* **deps:** bump the github-actions group across 1 directory with 3 updates ([#222](https://github.com/Josh-Archer/terraform-provider-seerr/issues/222)) ([addedab](https://github.com/Josh-Archer/terraform-provider-seerr/commit/addedab1c5f6d72ed2d79a41ac8a7399f48cffcb))
* **deps:** bump the github-actions group across 1 directory with 4 updates ([#191](https://github.com/Josh-Archer/terraform-provider-seerr/issues/191)) ([c6acb52](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c6acb52fa5e38f1efc985a7783c822e43c8b802e))
* **deps:** bump the github-actions group with 3 updates ([#174](https://github.com/Josh-Archer/terraform-provider-seerr/issues/174)) ([b620c2e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b620c2ec3259279a7eadf42e41c4dbc16aec3b52))
* **deps:** bump the github-actions group with 5 updates ([#127](https://github.com/Josh-Archer/terraform-provider-seerr/issues/127)) ([82b12de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/82b12de7889f8577cb20e9296547f7470bfaf686))
* **deps:** bump the go-dependencies group across 1 directory with 3 updates ([#95](https://github.com/Josh-Archer/terraform-provider-seerr/issues/95)) ([f7a54d8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f7a54d83cfbc4b0a86eb31b7b7b80fc9ae60250b))
* fix tofu CI - use dev_overrides and skip init for dev provider ([c39ace6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c39ace6e28462600ae079b5c5aec5fc33caf4581))
* fix tofu init in CI by setting job-level config env ([c47fc7a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c47fc7a7cf8d00f0c653e241758ee192d15dd085))
* fix tofu integration tests in CI using dev_overrides ([dbe6b04](https://github.com/Josh-Archer/terraform-provider-seerr/commit/dbe6b0407124acdd3c1effca8b8fa971f132f263))
* integrate Tofu tests into main CI and skip chores for tagging ([fa441e7](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa441e731d5e243c0c434e9a90d13180cac21c20))
* **main:** release 0.38.1 ([#189](https://github.com/Josh-Archer/terraform-provider-seerr/issues/189)) ([0122c97](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0122c97ba36c0c29090288738973ab921d65fb25))
* **main:** release 0.38.2 ([#194](https://github.com/Josh-Archer/terraform-provider-seerr/issues/194)) ([31e0531](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e05316e0c62160618443f184eff2bd2e04f531))
* **main:** release 0.38.3 ([#204](https://github.com/Josh-Archer/terraform-provider-seerr/issues/204)) ([5c270e8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/5c270e8f13b5847c6c5942ceb5b0fd30a0ea0b31))
* **main:** release 0.39.0 ([#208](https://github.com/Josh-Archer/terraform-provider-seerr/issues/208)) ([bcd4935](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bcd49352b4bad133b9d544cd826153a8083cf796))
* **main:** release 0.41.0 ([#224](https://github.com/Josh-Archer/terraform-provider-seerr/issues/224)) ([b10ea8c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b10ea8c83e0c072e9fb40f9529ae48a49aeb8613))
* **main:** release 0.42.0 ([#225](https://github.com/Josh-Archer/terraform-provider-seerr/issues/225)) ([d9aa9ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d9aa9adb743a3a26991f99aed948aedef6b6da8f))
* **main:** release 0.42.1 ([#232](https://github.com/Josh-Archer/terraform-provider-seerr/issues/232)) ([73a8c2b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/73a8c2ba8565c4aba4148ed292f7bd7d0cca22c6))
* **main:** release 0.42.2 ([#235](https://github.com/Josh-Archer/terraform-provider-seerr/issues/235)) ([8e6db2f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/8e6db2fb14f3d0eef9329294332959646c023823))
* **main:** release 0.42.3 ([#239](https://github.com/Josh-Archer/terraform-provider-seerr/issues/239)) ([f39a71f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f39a71ffcb87e5e0ceb8c188f892856e1ff2a7bf))
* migrate to native skipping via default_bump: false in tag action ([d571bb6](https://github.com/Josh-Archer/terraform-provider-seerr/commit/d571bb632e3b34086cbe2b6f7f40b7f33ce7db63))
* **release:** bump version to 0.40.0 ([bf608f5](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bf608f56c9836c2ba43a244956e4d6f3016d784f))
* **release:** configure release-please to bump minor version for feat commits pre-1.0 ([#210](https://github.com/Josh-Archer/terraform-provider-seerr/issues/210)) ([aef9ca8](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aef9ca88bc47c2ca91f2e9787ce78469184f17aa))
* remove research files ([732e4d1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/732e4d18d83cc2fe69e12a0364653615dc8e5016))
* replace with copyloopvar ([#253](https://github.com/Josh-Archer/terraform-provider-seerr/issues/253)) ([1db5214](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1db52140d3463d1360b812c1cf537bafbff2b8a5))
* trigger signed release after adding gpg secrets ([4851d6e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4851d6e28be6deb4a3609ebc2f98e8ba60d6fb9d))
* Update minimum Go version in module ([#282](https://github.com/Josh-Archer/terraform-provider-seerr/issues/282)) ([4cfdd3b](https://github.com/Josh-Archer/terraform-provider-seerr/commit/4cfdd3b4951c90e4759d993a1399a6cbc591c92c))
* use -plugin-dir for tofu init in CI with explicit version ([0ef218f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ef218f673ad85e19b8940745c05185981193604))
* use filesystem mirror for tofu init in CI ([b1d46c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b1d46c06fe5b448859f73b10134d2784e1a86ee6))


### Tests

* align Jellyfin fixture authentication ([#240](https://github.com/Josh-Archer/terraform-provider-seerr/issues/240)) ([a32f3ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a32f3ad64abcc84e0ee165c6e41a7f207c00a9f9))
* **arch:** add Single Egress Rule architecture test and isolate test network fixtures ([#234](https://github.com/Josh-Archer/terraform-provider-seerr/issues/234)) ([ab1841e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ab1841ecac85c517d88a75e3f607e95814d55038))
* implement comprehensive OpenTofu integration tests for all provider features ([cbdb3c3](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cbdb3c33e6610ed9bbe74f38712dd21e8113b5d6))


### Build System

* **deps:** bump github.com/cloudflare/circl from 1.6.1 to 1.6.3 ([#346](https://github.com/Josh-Archer/terraform-provider-seerr/issues/346)) ([3277b3c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3277b3cf37b87da164869f8b2d254241a2a51782))
* **deps:** bump github.com/cloudflare/circl in /tools ([#347](https://github.com/Josh-Archer/terraform-provider-seerr/issues/347)) ([a15d93e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a15d93e43295e34ad94ac67cd8a8bba0481de918))
* **deps:** bump github.com/hashicorp/terraform-plugin-framework ([#348](https://github.com/Josh-Archer/terraform-provider-seerr/issues/348)) ([c681e7e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/c681e7ea74469ede4e0c90b7f1f11f57e00f798f))
* **deps:** bump github.com/hashicorp/terraform-plugin-go ([#345](https://github.com/Josh-Archer/terraform-provider-seerr/issues/345)) ([ba92b67](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ba92b67ebc725872123b7e1dfa3e58c80bdba20c))
* **deps:** bump the github-actions group across 1 directory with 3 updates ([#349](https://github.com/Josh-Archer/terraform-provider-seerr/issues/349)) ([3957103](https://github.com/Josh-Archer/terraform-provider-seerr/commit/395710343f1240a8a4a6af99bbc10b2383e5fc12))


### Continuous Integration

* add `golangci-lint` + fix lints ([#120](https://github.com/Josh-Archer/terraform-provider-seerr/issues/120)) ([fa8056c](https://github.com/Josh-Archer/terraform-provider-seerr/commit/fa8056ce94248bf396295c0b208553f655782e91))
* Add GitHub Actions CI workflow for build, test, linting, and automated version tagging. ([f23eaa1](https://github.com/Josh-Archer/terraform-provider-seerr/commit/f23eaa1be4b62bb8033704635703ab98c0d33ad5))
* add hosted-runner fallback and remove template dependabot config ([aa59965](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aa5996555a7fabd0f71985ab8fa1709c2edd68d4))
* add OpenAPI coverage check step to CI workflow ([#147](https://github.com/Josh-Archer/terraform-provider-seerr/issues/147)) ([82a753f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/82a753f32123e20d9156da78892e8b72ecbcafaa))
* auto-tag main with semver after successful CI ([cb4faee](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cb4faeeda6c651e164a863a45cc21edee72cc4d3))
* avoid secrets in if expressions for release workflow ([cfb0351](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cfb0351d2a1c627d59195b7e4ca2bbf5cd9732a9))
* disable golangci-lint remote schema verification and add auto-merge for owner PRs ([#202](https://github.com/Josh-Archer/terraform-provider-seerr/issues/202)) ([76be0d0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/76be0d0c7d8e17267689bc8611a12c06f7731dbd))
* dispatch release after auto-tag ([eaccb1a](https://github.com/Josh-Archer/terraform-provider-seerr/commit/eaccb1ad26630d3ca67c955fee7046ff89daf0ac))
* dispatch Release after auto-tag on main ([#91](https://github.com/Josh-Archer/terraform-provider-seerr/issues/91)) ([aae4f46](https://github.com/Josh-Archer/terraform-provider-seerr/commit/aae4f46a48d04b05180904d9276ffaa88f339eec))
* integration tests ([9c00c20](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9c00c205680a9a75c471215115269472d16250f7))
* remove legacy auto-tag from test.yml in favor of staged release-please ([#211](https://github.com/Josh-Archer/terraform-provider-seerr/issues/211)) ([b8a2e50](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b8a2e505b2673ddc18754238ceb6f8036564054b))
* replace custom PowerShell semver with mathieudutour/github-tag-action and add Copilot commit instructions ([1724cda](https://github.com/Josh-Archer/terraform-provider-seerr/commit/1724cda5798c287fe72a253ac1375d230d495545))
* replace forked GH action with latest upstream ([#55](https://github.com/Josh-Archer/terraform-provider-seerr/issues/55)) ([3a91275](https://github.com/Josh-Archer/terraform-provider-seerr/commit/3a91275abc3c45c61c888142c64c9d5abef588b2))
* support unsigned release fallback when GPG secrets are unset ([b674265](https://github.com/Josh-Archer/terraform-provider-seerr/commit/b674265f4a0b4dcc178444e1ff907db45dd33be4))

## [0.42.3](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.42.2...v0.42.3) (2026-08-30)


### Bug Fixes

* restore release and Jellyfin compatibility gates ([#238](https://github.com/Josh-Archer/terraform-provider-seerr/issues/238)) ([7cb5743](https://github.com/Josh-Archer/terraform-provider-seerr/commit/7cb57430b9fbab52b72ce7aa7f3852729fb3d291))


### Documentation

* mark v0.43.0 module release stable ([#241](https://github.com/Josh-Archer/terraform-provider-seerr/issues/241)) ([0ac711f](https://github.com/Josh-Archer/terraform-provider-seerr/commit/0ac711f63685c962c0047d69190e2b795662357d))


### Tests

* align Jellyfin fixture authentication ([#240](https://github.com/Josh-Archer/terraform-provider-seerr/issues/240)) ([a32f3ad](https://github.com/Josh-Archer/terraform-provider-seerr/commit/a32f3ad64abcc84e0ee165c6e41a7f207c00a9f9))

## [0.42.2](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.42.1...v0.42.2) (2026-08-30)


### Bug Fixes

* **ci:** harden release reconciliation to discover draft releases and self-heal ([#236](https://github.com/Josh-Archer/terraform-provider-seerr/issues/236)) ([63d76fc](https://github.com/Josh-Archer/terraform-provider-seerr/commit/63d76fc34616ca4cf2157d1cfb7cd178ee005c15))


### Documentation

* consolidate roadmap into README ([#237](https://github.com/Josh-Archer/terraform-provider-seerr/issues/237)) ([cc54970](https://github.com/Josh-Archer/terraform-provider-seerr/commit/cc54970d7e883cfda0580660f1e559115dd715e7))


### Tests

* **arch:** add Single Egress Rule architecture test and isolate test network fixtures ([#234](https://github.com/Josh-Archer/terraform-provider-seerr/issues/234)) ([ab1841e](https://github.com/Josh-Archer/terraform-provider-seerr/commit/ab1841ecac85c517d88a75e3f607e95814d55038))

## [0.42.1](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.42.0...v0.42.1) (2026-08-30)


### Bug Fixes

* **servarr:** validate Arr connectivity via Seerr server proxy test endpoint ([#231](https://github.com/Josh-Archer/terraform-provider-seerr/issues/231)) ([2d128c0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/2d128c089885ba5bab8033aa516f585ea0767073))

## [0.42.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.41.0...v0.42.0) (2026-08-29)


### Features

* add reusable module ecosystem ([#228](https://github.com/Josh-Archer/terraform-provider-seerr/issues/228)) ([99cdee0](https://github.com/Josh-Archer/terraform-provider-seerr/commit/99cdee04530ab63296736e06d4d3a6e3719cfd98))


### Bug Fixes

* **ci:** create tags for draft releases ([#226](https://github.com/Josh-Archer/terraform-provider-seerr/issues/226)) ([18ed2de](https://github.com/Josh-Archer/terraform-provider-seerr/commit/18ed2de8764b6c8eea77a73f46194203ea696cde))
* **ci:** pass explicit tag to goreleaser ([#227](https://github.com/Josh-Archer/terraform-provider-seerr/issues/227)) ([9633ada](https://github.com/Josh-Archer/terraform-provider-seerr/commit/9633adae29b4d881be99692a9864b71a44bef107))
* **ci:** query the OpenTofu provider endpoint ([#229](https://github.com/Josh-Archer/terraform-provider-seerr/issues/229)) ([6095c2d](https://github.com/Josh-Archer/terraform-provider-seerr/commit/6095c2db2e0f50339e6b28b6f48062b78042b250))


### Documentation

* clarify reusable module usage ([#230](https://github.com/Josh-Archer/terraform-provider-seerr/issues/230)) ([31e1435](https://github.com/Josh-Archer/terraform-provider-seerr/commit/31e1435968cc8e3ae592c8124c9db1860ebb4053))


### Miscellaneous Chores

* **deps:** bump the github-actions group across 1 directory with 3 updates ([#222](https://github.com/Josh-Archer/terraform-provider-seerr/issues/222)) ([addedab](https://github.com/Josh-Archer/terraform-provider-seerr/commit/addedab1c5f6d72ed2d79a41ac8a7399f48cffcb))

## [0.41.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.40.0...v0.41.0) (2026-08-29)


### Features

* **compatibility:** verify current upstream releases ([#223](https://github.com/Josh-Archer/terraform-provider-seerr/issues/223)) ([bfb3e49](https://github.com/Josh-Archer/terraform-provider-seerr/commit/bfb3e495cb90f9bccbf75d108231949c7bdf9c09))

## [0.40.0](https://github.com/Josh-Archer/terraform-provider-seerr/compare/v0.39.1...v0.40.0) (2026-08-28)


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
