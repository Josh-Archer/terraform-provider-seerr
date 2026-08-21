# Project Roadmap: `terraform-provider-seerr`

Tracked officially in [GitHub Discussion #161](https://github.com/Josh-Archer/terraform-provider-seerr/discussions/161).

---

## 📊 Phase Progress Matrix

| Phase | Version | Issue | Key Theme | Status |
|:---|:---|:---|:---|:---|
| **1** | `v0.31.0` | — | Stability and Quality | ✅ Completed |
| **2** | `v0.32.0` | — | Observability Data Sources | ✅ Completed |
| **3** | `v0.33.0` | — | User Management and Watchlists | ✅ Completed |
| **4** | `v0.34.0` | — | TMDB Reference Helpers | ✅ Completed |
| **5** | `v0.35.0` | [#162](https://github.com/Josh-Archer/terraform-provider-seerr/issues/162) | Provider Hardening (Import, Defaults, Validation) | ✅ Completed |
| **6** | `v0.36.0` | [#163](https://github.com/Josh-Archer/terraform-provider-seerr/issues/163) | Advanced Resource Lifecycle (Approvals, Comments) | ✅ Completed |
| **7** | `v0.37.0` | [#164](https://github.com/Josh-Archer/terraform-provider-seerr/issues/164) | Developer Experience & Arr Resolvers | ✅ Completed (`v0.37.1`) |
| **8** | `v0.38.0` | [#165](https://github.com/Josh-Archer/terraform-provider-seerr/issues/165) | Production Readiness & Live Homelab Validation | ✅ Completed (`v0.38.0`) |
| **8b** | `v0.38.1` | [#181](https://github.com/Josh-Archer/terraform-provider-seerr/issues/181) | Release Please & Zero-Click Automated Publishing | ✅ Completed (`v0.38.1`) |
| **10** | `v0.38.2` | [#182](https://github.com/Josh-Archer/terraform-provider-seerr/issues/182) | Upstream Compatibility Automation (OpenAPI Drift) | ✅ Completed (`v0.38.2`) |
| **15** | `v0.38.1` | [#185](https://github.com/Josh-Archer/terraform-provider-seerr/issues/185) | Dual Registry Publishing (HashiCorp + OpenTofu) | ✅ Completed |
| **12** | `v0.39.0` | [#183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183) | Community Readiness & Contributor Onboarding | 🔵 Next |
| **11** | `v0.40.0` | [#170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170) | Import and Migration Tooling (Bulk CLI) | 📅 Planned |
| **9** | `v0.41.0` | [#169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169) | State Resilience & Upgraders | 📅 Planned |
| **13** | `v0.42.0` | [#184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184) | Observability & Disaster Recovery | 📅 Planned |
| **14** | `v0.43.0` | [#171](https://github.com/Josh-Archer/terraform-provider-seerr/issues/171) | Module Ecosystem & Stack Presets | 📅 Planned |
| **v1.0** | `v1.0.0` | — | Stability Guarantee & General Availability | 🏁 Planned |

---

## 🎯 Completed Milestone Highlights

### 1. Dual Registry Publishing ([Issue #185](https://github.com/Josh-Archer/terraform-provider-seerr/issues/185))
- Published and actively verified on **[HashiCorp Terraform Registry](https://registry.terraform.io/providers/josh-archer/seerr)** and **[OpenTofu Registry](https://registry.opentofu.org/providers/josh-archer/seerr)**.
- Full RSA 4096-bit GPG signature verification across all releases.

### 2. Automated Release Pipeline ([Issue #181](https://github.com/Josh-Archer/terraform-provider-seerr/issues/181))
- Automated release workflow using [Google Release Please](https://github.com/googleapis/release-please).
- Zero-click publishing triggered automatically upon merging rolling release PRs.
- Automatic SemVer tagging and structured CHANGELOG generation.

### 3. Upstream Compatibility Automation ([Issue #182](https://github.com/Josh-Archer/terraform-provider-seerr/issues/182))
- Weekly automated OpenAPI schema drift detection (`.github/workflows/openapi-drift.yml`).
- Daily automated upstream release watcher (`.github/workflows/upstream-watch.yml`).
- Version compatibility guide in [`docs/guides/compatibility.md`](docs/guides/compatibility.md) and [`COMPATIBILITY.md`](COMPATIBILITY.md).

### 4. Comprehensive Coverage & Hardening
- 50 managed resources and 69 data sources covering 100% of applicable OpenAPI configuration endpoints.
- Dynamic entity resolvers for Radarr and Sonarr quality profiles, root folders, tags, and language profiles.
- 11 typed notification agents with live test execution triggers.

---

## 🔮 Upcoming Phases

### Phase 12 — Community Readiness ([Issue #183](https://github.com/Josh-Archer/terraform-provider-seerr/issues/183))
- Structured GitHub Issue and Pull Request templates.
- VS Code dev container and local Docker Compose test environment.
- Contributor onboarding guide (`CONTRIBUTING.md`), `CODEOWNERS`, and Code of Conduct.

### Phase 11 — Import & Migration Tooling ([Issue #170](https://github.com/Josh-Archer/terraform-provider-seerr/issues/170))
- Bulk configuration exporter CLI (`tools/importer`) generating Terraform HCL and `import` blocks from a live Seerr instance.
- Step-by-step migration guide for onboarding existing self-hosted Seerr deployments.

### Phase 9 — State Resilience & Diff Suppression ([Issue #169](https://github.com/Josh-Archer/terraform-provider-seerr/issues/169))
- State upgraders for schema migrations across version upgrades.
- Robust 404 handling across all resources for graceful state reconciliation on external deletions.

### Phase 13 — Observability & Disaster Recovery ([Issue #184](https://github.com/Josh-Archer/terraform-provider-seerr/issues/184))
- State backup, restoration, and disaster recovery runbooks.
- Pre-built Grafana/Prometheus dashboard references for Seerr API activity.

### Phase 14 — Module Ecosystem ([Issue #171](https://github.com/Josh-Archer/terraform-provider-seerr/issues/171))
- Curated composite modules for turnkey family media server deployments, arr stack configurations, and notification routing profiles.

### v1.0.0 — General Availability
- Zero schema breaking changes guarantee.
- Comprehensive end-to-end acceptance test coverage across multiple Seerr versions.
