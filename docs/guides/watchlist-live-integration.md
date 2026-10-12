# Watchlist live integration

The watchlist live integration exercises the provider against a disposable Seerr 3.5.0 container and a real OpenTofu process. It seeds 42 native watchlist rows directly into the container's SQLite database so the test needs no TMDB or Plex account and cannot change an existing Seerr instance.

Run it from the repository root with Docker, Go, and OpenTofu 1.8.0 available:

```shell
python3 ./scripts/test_watchlist_live.py
```

The script checks both the authenticated-user and explicit-user data source reads, imports an item from the final watchlist page, confirms an unchanged plan, deletes that item through Seerr's API, then checks that refresh-only removes that state entry while preserving the other 41 items. It starts a uniquely named Compose project with its own volume and removes that project in a `finally` cleanup path. It refuses `SEERR_URL` or `SEERR_API_KEY` environment variables; the disposable local container is always the target.

On failure, container logs, Compose logs, OpenTofu state, and plan data remain under `test-artifacts/watchlist-live/` for diagnosis. The GitHub Actions `Watchlist Live Integration (Seerr 3.5.0)` job runs this test on pull requests and main pushes using a GitHub-hosted runner, and automatic merging waits for it to pass.
