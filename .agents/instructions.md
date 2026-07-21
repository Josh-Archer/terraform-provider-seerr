# Agent Instructions

## Git Workflow

- **NEVER** commit directly to the `main` or `master` branches.
- **ALWAYS** create a new branch for your work (e.g., `git checkout -b feature/issue-name` or `git checkout -b fix/issue-name`).
- **ALWAYS** ensure that Git commit GPG/SSH signing is enabled and uses the default global signing key (e.g. from Bitwarden/ssh-agent).
- **IMPORTANT**: Ensure your SSH agent (e.g., Bitwarden vault) is unlocked when tasks are running so commits can be signed successfully without blocking.
- **ALWAYS** submit a Pull Request (PR) using the `gh` CLI (e.g., `gh pr create`) after pushing your branch to the remote repository. Let the user review the PR.
- **ALWAYS** run `go generate ./...` as the final code-change step before wrapping up so generated provider docs stay in sync.
