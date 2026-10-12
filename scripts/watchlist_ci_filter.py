#!/usr/bin/env python3
"""Select the live watchlist test without skipping post-merge validation."""

import subprocess
import sys


def should_run(event: str, paths: list[str]) -> bool:
    if event == "push":
        return True
    if event != "pull_request":
        return False
    exact = {"main.go", "go.mod", "go.sum", ".github/workflows/test.yml"}
    prefixes = ("internal/provider/", "scripts/", "tests/")
    return any(path in exact or path.startswith(prefixes) for path in paths)


def pr_paths() -> list[str]:
    # Checkout's default PR ref is GitHub's merge commit. Its first parent is
    # the base tree; --no-renames includes both paths when code moves away.
    parents = subprocess.check_output(
        ["git", "show", "-s", "--format=%P", "HEAD"], text=True, timeout=30
    ).split()
    if len(parents) != 2:
        raise RuntimeError("PR filtering requires GitHub's two-parent merge commit")
    changed = subprocess.check_output(
        ["git", "diff", "--no-ext-diff", "--name-only", "--no-renames", "-z", "HEAD^1", "HEAD"],
        timeout=30,
    )
    return [path.decode("utf-8", errors="surrogateescape") for path in changed.split(b"\0") if path]


def main() -> None:
    event = sys.argv[1]
    paths = pr_paths() if event == "pull_request" else []
    print("true" if should_run(event, paths) else "false")


if __name__ == "__main__":
    main()
