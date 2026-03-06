# GitHub Copilot Instructions

## Commit Message Conventions

This repository uses [Conventional Commits](https://www.conventionalcommits.org/) to drive automatic semantic versioning via `mathieudutour/github-tag-action`. Every commit message **must** follow the format below so that merges to `main` produce the correct version bump.

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types and their version impact

| Type | Example | Version bump |
|------|---------|-------------|
| `feat` | `feat: add user watchlist resource` | **minor** (0.x.0) |
| `feat!` or `BREAKING CHANGE` footer | `feat!: remove legacy API support` | **major** (x.0.0) |
| `fix` | `fix: correct status_code handling` | patch (0.0.x) |
| `chore`, `ci`, `docs`, `refactor`, `test`, `style`, `build`, `perf` | `chore: update dependencies` | patch (0.0.x) |

### Rules

1. **Always use a conventional commit prefix** on every commit and PR merge commit (squash/merge).
2. **New resources, data sources, or provider capabilities** → use `feat:`.
3. **Bug fixes** → use `fix:`.
4. **Breaking API or schema changes** → use `feat!:` or add `BREAKING CHANGE: <description>` in the commit footer.
5. **Tooling, CI, docs, refactors with no user-facing change** → use `chore:`, `ci:`, `docs:`, or `refactor:`.

### Examples

```
feat: add seerr_user_permissions resource
fix: resolve nil pointer in notification agent read
feat(plex): add plex_settings data source
chore: update golangci-lint to v2.5.0
ci: switch auto-tag to mathieudutour/github-tag-action
feat!: rename seerr_api_object id field to object_id

BREAKING CHANGE: the `id` attribute on seerr_api_object has been renamed to `object_id`
```
