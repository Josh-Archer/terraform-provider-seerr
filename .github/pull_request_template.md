## Summary

<!-- Briefly describe the purpose and outcome of this pull request. -->

Fixes # <!-- INSERT ISSUE NUMBER (e.g. Closes #123) -->

---

## Type of Change

- [ ] `feat`: New resource, data source, or provider capability
- [ ] `fix`: Bug fix in resource lifecycle, schema, or error handling
- [ ] `docs`: Documentation updates, guide additions, or example fixes
- [ ] `test`: Unit, integration, or acceptance test improvements
- [ ] `refactor`: Internal code refactoring without schema changes
- [ ] `ci`: Workflow, release, or automation pipeline adjustments

---

## Key Changes

<!-- Provide a bulleted summary of key changes and architectural decisions. -->
- 

---

## Verification & Testing

<!-- Describe how you verified your changes. -->
- [ ] Ran `gofmt -s -w .` (Go code formatting)
- [ ] Ran unit tests: `go test -v -count=1 ./...`
- [ ] Ran fast merge gate: `bash ./scripts/test-all-locally.sh`
- [ ] Generated docs / schema sync: `go generate ./...`
- [ ] Tested with live instance (OpenTofu / Terraform) *(if applicable)*

---

## Checklist

- [ ] My commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) (e.g., `feat: ...`, `fix: ...`).
- [ ] Relevant documentation under `docs/` and templates under `templates/` are updated.
- [ ] Sensitive tokens, passwords, and API keys are redacted from tests and examples.
