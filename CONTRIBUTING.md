# Contributing

Contributions are welcome through fork-based pull requests.

1. Fork the repository and create a focused branch.
2. Run `bash ./scripts/test-all-locally.sh` before opening a pull request.
3. Open a pull request against `main` and link the issue it addresses.
4. Respond to review feedback and keep the branch up to date.

Pull-request CI is deliberately untrusted: it runs on `ubuntu-latest`, does not receive repository secrets, and must not be used to test private infrastructure. Integration tests use an ephemeral local Seerr target when credentials are not provided.

Trusted integration runs happen only from `main`, schedules, or manually dispatched workflows. Maintainers may configure the `TRUSTED_RUNNER_LABEL` repository or organization variable to route those runs to UECB or an isolated self-hosted runner. Do not use that runner for pull requests, and do not give it access to unrelated home-network systems.

Please do not include credentials, tokens, private endpoints, generated state, or personal data in issues or pull requests.
