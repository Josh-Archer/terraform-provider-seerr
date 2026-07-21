# Agent Instructions

## Git Workflow

- **NEVER** commit directly to the `main` or `master` branches.
- **ALWAYS** create a new branch for your work (e.g., `git checkout -b feature/issue-name` or `git checkout -b fix/issue-name`).
- **ALWAYS** run `git config commit.gpgsign false` locally in your workspace/branch immediately after checking it out so that commit operations do not fail/block due to missing GPG/SSH agent keys inside sandboxed environments.
- **ALWAYS** submit a Pull Request (PR) using the `gh` CLI (e.g., `gh pr create`) after pushing your branch to the remote repository. Let the user review the PR.
- **ALWAYS** run `go generate ./...` as the final code-change step before wrapping up so generated provider docs stay in sync.
