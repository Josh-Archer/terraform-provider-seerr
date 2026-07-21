# Agent Instructions

## Git Workflow

- **NEVER** commit directly to the `main` or `master` branches.
- **ALWAYS** create a new branch for your work (e.g., `git checkout -b feature/issue-name` or `git checkout -b fix/issue-name`).
- **ALWAYS** configure SSH commit signing locally in your workspace/branch to use your SSH private key file directly. This prevents commit operations from blocking or failing due to missing agent keys/biometrics inside sandboxed environments:
  - On Windows: run `git config user.signingkey "$env:USERPROFILE/.ssh/id_ed25519"` (or specify your active private key path)
  - On macOS/Linux: run `git config user.signingkey "$HOME/.ssh/id_ed25519"` (or specify your active private key path)
  - Ensure local `commit.gpgsign` is set to `true` via `git config commit.gpgsign true`
- **ALWAYS** submit a Pull Request (PR) using the `gh` CLI (e.g., `gh pr create`) after pushing your branch to the remote repository. Let the user review the PR.
- **ALWAYS** run `go generate ./...` as the final code-change step before wrapping up so generated provider docs stay in sync.
