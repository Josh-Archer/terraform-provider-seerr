#!/usr/bin/env python3
"""Fail when the tested Seerr versions drift from .github/upstream-compatibility.json."""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
COMPAT = ROOT / ".github" / "upstream-compatibility.json"
WORKFLOW = ROOT / ".github" / "workflows" / "test.yml"
COMPOSE = ROOT / "tests" / "docker-compose.yml"
COMPAT_DOC = ROOT / "COMPATIBILITY.md"


def check(compat: dict, workflow: str, compose: str, compat_doc: str) -> list[str]:
    errors = []
    matrix_images = set(re.findall(r"^\s+image:\s*(\S+)\s*$", workflow, re.M))
    for name, entry in compat.items():
        image = entry["image"]
        if image not in matrix_images:
            errors.append(f"{name}: verified image {image} is not in a test.yml integration matrix")
        if entry["release"] not in compat_doc:
            errors.append(f"{name}: verified release {entry['release']} is not listed in COMPATIBILITY.md")
    seerr_image = compat["seerr"]["image"]
    for default in re.findall(r"SEERR_TEST_IMAGE \|\| '([^']+)'", workflow):
        if default != seerr_image:
            errors.append(f"test.yml SEERR_TEST_IMAGE default {default} != verified {seerr_image}")
    for default in re.findall(r"SEERR_TEST_IMAGE:-([^}\s]+)}", compose):
        if default != seerr_image:
            errors.append(f"tests/docker-compose.yml default {default} != verified {seerr_image}")
    return errors


def main() -> int:
    errors = check(
        json.loads(COMPAT.read_text(encoding="utf-8")),
        WORKFLOW.read_text(encoding="utf-8"),
        COMPOSE.read_text(encoding="utf-8"),
        COMPAT_DOC.read_text(encoding="utf-8"),
    )
    for error in errors:
        print(f"compat matrix drift: {error}", file=sys.stderr)
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
