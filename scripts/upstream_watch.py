#!/usr/bin/env python3
"""Reconcile the upstream release status comment on compatibility issue #314."""

from __future__ import annotations

import json
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Sequence


REPOSITORY = "Josh-Archer/terraform-provider-seerr"
TRACKING_ISSUE = 314
STATUS_MARKER = "<!-- upstream-watch:status -->"
COMPATIBILITY_FILE = Path(__file__).resolve().parents[1] / ".github/upstream-compatibility.json"
RELEASE_REPOSITORIES = {
    "seerr": "seerr-team/seerr",
    "overseerr": "sct/overseerr",
}


class UpstreamCheckError(RuntimeError):
    """Raised when a release endpoint fails or returns no usable release tag."""


@dataclass(frozen=True)
class ReleaseStatus:
    latest: dict[str, str]
    verified: dict[str, str]

    @property
    def needs_verification(self) -> bool:
        return any(self.latest[name] != self.verified[name] for name in self.latest)


def run_gh(args: Sequence[str]) -> str:
    """Run gh and return stdout, preserving failures instead of inventing a tag."""
    try:
        result = subprocess.run(
            ["gh", *args], check=True, capture_output=True, text=True
        )
    except (OSError, subprocess.CalledProcessError) as exc:
        raise UpstreamCheckError(f"GitHub CLI request failed: {' '.join(args)}: {exc}") from exc
    return result.stdout.strip()


def load_verified() -> dict[str, str]:
    try:
        metadata = json.loads(COMPATIBILITY_FILE.read_text(encoding="utf-8"))
        verified = {name: entry["release"] for name, entry in metadata.items()}
    except (OSError, json.JSONDecodeError, KeyError, TypeError) as exc:
        raise UpstreamCheckError(f"cannot read verified upstream releases: {exc}") from exc
    required = {"seerr", "jellyseerr", "overseerr"}
    if set(verified) != required or any(not isinstance(tag, str) or not tag for tag in verified.values()):
        raise UpstreamCheckError("verified release metadata must contain non-empty tags for all projects")
    return verified


def fetch_latest_tag(repository: str) -> str:
    tag = run_gh(["api", f"repos/{repository}/releases/latest", "--jq", ".tag_name"])
    if not tag or tag == "null":
        raise UpstreamCheckError(f"latest release response from {repository} had no tag_name")
    return tag


def check_releases() -> ReleaseStatus:
    verified = load_verified()
    latest = {
        name: fetch_latest_tag(repository)
        for name, repository in RELEASE_REPOSITORIES.items()
    }
    # Jellyseerr redirects to Seerr, so querying releases/latest would report a
    # Seerr v3 tag. Keep the last release from the legacy repository instead.
    latest["jellyseerr"] = "v2.7.3"
    return ReleaseStatus(latest=latest, verified=verified)


def render_status(status: ReleaseStatus) -> str:
    if status.needs_verification:
        headline = "**Verification needed:** one or more upstream releases are newer than the tested versions."
        action = (
            "### Follow-up for each changed upstream\n"
            "- Verify the provider test suite against the release image listed in "
            "[the compatibility matrix](https://github.com/Josh-Archer/terraform-provider-seerr/blob/main/.github/workflows/test.yml).\n"
            "- Compare the upstream OpenAPI specification and update the "
            "[compatibility guide](https://github.com/Josh-Archer/terraform-provider-seerr/blob/main/docs/guides/compatibility.md) "
            "after verification.\n"
        )
    else:
        headline = "**Current:** all latest upstream releases match the verified versions; no compatibility action is needed."
        action = ""

    lines = [
        STATUS_MARKER,
        "## Upstream Release Status",
        "",
        headline,
        "",
        "| Upstream project | Latest release | Verified release | Evidence |",
        "|---|---|---|---|",
    ]
    evidence = {
        "seerr": "[release](https://github.com/seerr-team/seerr/releases/latest)",
        "jellyseerr": "[legacy release](https://github.com/seerr-team/seerr/releases/tag/v2.7.3)",
        "overseerr": "[release](https://github.com/sct/overseerr/releases/latest)",
    }
    labels = {"seerr": "Seerr (unified)", "jellyseerr": "Jellyseerr (legacy)", "overseerr": "Overseerr"}
    for name in ("seerr", "jellyseerr", "overseerr"):
        lines.append(
            f"| {labels[name]} | `{status.latest[name]}` | `{status.verified[name]}` | {evidence[name]} |"
        )
    if action:
        lines.extend(["", action.rstrip()])
    lines.extend(["", "Tracker: #314. This comment is updated in place by the upstream watch workflow."])
    return "\n".join(lines) + "\n"


def find_status_comment(comments: list[dict[str, Any]]) -> dict[str, Any] | None:
    marked = [comment for comment in comments if STATUS_MARKER in comment.get("body", "")]
    if len(marked) > 1:
        raise UpstreamCheckError(
            f"issue #{TRACKING_ISSUE} has {len(marked)} comments with the status marker; refusing to create another"
        )
    return marked[0] if marked else None


def reconcile_comment(body: str) -> str:
    raw_comments = run_gh(
        [
            "api",
            f"repos/{REPOSITORY}/issues/{TRACKING_ISSUE}/comments?per_page=100",
        ]
    )
    try:
        comments = json.loads(raw_comments)
        if not isinstance(comments, list):
            raise ValueError("comments response was not a JSON array")
    except (json.JSONDecodeError, ValueError) as exc:
        raise UpstreamCheckError(f"cannot read comments on issue #{TRACKING_ISSUE}: {exc}") from exc

    existing = find_status_comment(comments)
    if existing and existing.get("body") == body:
        return "unchanged"
    if existing:
        comment_id = existing.get("id")
        if not isinstance(comment_id, int):
            raise UpstreamCheckError("marked status comment has no valid comment id")
        run_gh(
            [
                "api",
                f"repos/{REPOSITORY}/issues/comments/{comment_id}",
                "-X",
                "PATCH",
                "-f",
                f"body={body}",
            ]
        )
        return "updated"

    run_gh(
        [
            "api",
            f"repos/{REPOSITORY}/issues/{TRACKING_ISSUE}/comments",
            "-X",
            "POST",
            "-f",
            f"body={body}",
        ]
    )
    return "created"


def main() -> None:
    status = check_releases()
    body = render_status(status)
    result = reconcile_comment(body)
    print(f"Upstream status {result} on issue #{TRACKING_ISSUE}.")


if __name__ == "__main__":
    main()
