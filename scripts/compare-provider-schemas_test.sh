#!/usr/bin/env bash
# Unit / functional test suite for compare-provider-schemas.sh and tools/schemacmp
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

tmp_dir="$(mktemp -d "/tmp/schemacmp-test.XXXXXX")"
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT

base_schema="${tmp_dir}/base.json"
cat > "${base_schema}" <<'JSON'
{
  "format_version": "1.0",
  "resources": {
    "seerr_user": {
      "attributes": {
        "id": {"type": "string", "computed": true},
        "email": {"type": "string", "required": true},
        "username": {"type": "string", "optional": true, "computed": true},
        "password": {"type": "string", "optional": true, "sensitive": true}
      }
    }
  },
  "data_sources": {
    "seerr_about": {
      "attributes": {
        "version": {"type": "string", "computed": true}
      }
    }
  }
}
JSON

# Test 1: Clean baseline vs itself passes
report_1="${tmp_dir}/report1.md"
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${base_schema}" \
  -report "${report_1}" \
  -quiet
test -f "${report_1}"
grep -q "100% Backward Compatible" "${report_1}"
echo "PASS: identical schema is 100% compatible"

# Test 2: Incompatible removal fails
cand_removed="${tmp_dir}/cand_removed.json"
cat > "${cand_removed}" <<'JSON'
{
  "format_version": "1.0",
  "resources": {
    "seerr_user": {
      "attributes": {
        "id": {"type": "string", "computed": true},
        "username": {"type": "string", "optional": true}
      }
    }
  },
  "data_sources": {
    "seerr_about": {
      "attributes": {
        "version": {"type": "string", "computed": true}
      }
    }
  }
}
JSON

report_2="${tmp_dir}/report2.md"
set +e
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${cand_removed}" \
  -report "${report_2}" \
  -quiet 2>/dev/null
rc=$?
set -e
test "${rc}" -ne 0
grep -q "ATTRIBUTE_REMOVED" "${report_2}"
grep -q "email" "${report_2}"
echo "PASS: incompatible attribute removal is detected"

# Test 3: Incompatible type change fails
cand_type="${tmp_dir}/cand_type.json"
cat > "${cand_type}" <<'JSON'
{
  "format_version": "1.0",
  "resources": {
    "seerr_user": {
      "attributes": {
        "id": {"type": "string", "computed": true},
        "email": {"type": "number", "required": true},
        "username": {"type": "string", "optional": true},
        "password": {"type": "string", "optional": true, "sensitive": true}
      }
    }
  },
  "data_sources": {
    "seerr_about": {
      "attributes": {
        "version": {"type": "string", "computed": true}
      }
    }
  }
}
JSON

report_3="${tmp_dir}/report3.md"
set +e
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${cand_type}" \
  -report "${report_3}" \
  -quiet 2>/dev/null
rc=$?
set -e
test "${rc}" -ne 0
grep -q "TYPE_MODIFIED" "${report_3}"
echo "PASS: incompatible type change is detected"

# Test 4: Newly required attribute fails
cand_new_req="${tmp_dir}/cand_new_req.json"
cat > "${cand_new_req}" <<'JSON'
{
  "format_version": "1.0",
  "resources": {
    "seerr_user": {
      "attributes": {
        "id": {"type": "string", "computed": true},
        "email": {"type": "string", "required": true},
        "username": {"type": "string", "required": true},
        "password": {"type": "string", "optional": true, "sensitive": true}
      }
    }
  },
  "data_sources": {
    "seerr_about": {
      "attributes": {
        "version": {"type": "string", "computed": true}
      }
    }
  }
}
JSON

report_4="${tmp_dir}/report4.md"
set +e
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${cand_new_req}" \
  -report "${report_4}" \
  -quiet 2>/dev/null
rc=$?
set -e
test "${rc}" -ne 0
grep -q "NEWLY_REQUIRED" "${report_4}"
echo "PASS: newly required attribute is detected"

# Test 5: Lost sensitive marking fails
cand_sensitive="${tmp_dir}/cand_sensitive.json"
cat > "${cand_sensitive}" <<'JSON'
{
  "format_version": "1.0",
  "resources": {
    "seerr_user": {
      "attributes": {
        "id": {"type": "string", "computed": true},
        "email": {"type": "string", "required": true},
        "username": {"type": "string", "optional": true},
        "password": {"type": "string", "optional": true, "sensitive": false}
      }
    }
  },
  "data_sources": {
    "seerr_about": {
      "attributes": {
        "version": {"type": "string", "computed": true}
      }
    }
  }
}
JSON

report_5="${tmp_dir}/report5.md"
set +e
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${cand_sensitive}" \
  -report "${report_5}" \
  -quiet 2>/dev/null
rc=$?
set -e
test "${rc}" -ne 0
grep -q "LOST_SENSITIVE" "${report_5}"
echo "PASS: lost sensitive marking is detected"

# Test 6: Backward compatible addition passes
cand_compat="${tmp_dir}/cand_compat.json"
cat > "${cand_compat}" <<'JSON'
{
  "format_version": "1.0",
  "resources": {
    "seerr_user": {
      "attributes": {
        "id": {"type": "string", "computed": true},
        "email": {"type": "string", "required": true},
        "username": {"type": "string", "optional": true, "computed": true},
        "password": {"type": "string", "optional": true, "sensitive": true},
        "avatar": {"type": "string", "optional": true}
      }
    },
    "seerr_invite": {
      "attributes": {
        "id": {"type": "string", "computed": true}
      }
    }
  },
  "data_sources": {
    "seerr_about": {
      "attributes": {
        "version": {"type": "string", "computed": true}
      }
    }
  }
}
JSON

report_6="${tmp_dir}/report6.md"
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${cand_compat}" \
  -report "${report_6}" \
  -quiet
grep -q "Backward-Compatible Additions" "${report_6}"
grep -q "seerr_invite" "${report_6}"
echo "PASS: backward compatible addition passes"

# Test 7: Policy enforcement - intentional breaking change with major bump and migration note passes
note_file="${tmp_dir}/note.md"
cat > "${note_file}" <<'NOTE'
# Migration to v3
Email field has been removed in favor of username authentication.
NOTE

report_7="${tmp_dir}/report7.md"
go run ./tools/schemacmp \
  -baseline-file "${base_schema}" \
  -candidate "${cand_removed}" \
  -previous "2.2.2" \
  -target-version "3.0.0" \
  -migration-note "${note_file}" \
  -allow-breaking \
  -report "${report_7}" \
  -quiet
grep -q "Policy Approved" "${report_7}"
echo "PASS: intentional breaking change with major bump and migration note passes"

# Test 8: Unavailable registry handling with --allow-offline succeeds
go run ./tools/schemacmp \
  -previous "99.99.99" \
  -allow-offline \
  -quiet 2>/dev/null
echo "PASS: unavailable registry handled gracefully with --allow-offline"

# Test 9: compare-provider-schemas.sh runs cleanly on workspace provider
export SEERR_SCHEMA_BASELINE_FILE="${repo_root}/testdata/provider_schema_baseline.json"
export SEERR_SCHEMA_REPORT_PATH="${tmp_dir}/workspace_report.md"
bash "${repo_root}/scripts/compare-provider-schemas.sh" -quiet
test -f "${tmp_dir}/workspace_report.md"
grep -q "100% Backward Compatible" "${tmp_dir}/workspace_report.md"
echo "PASS: scripts/compare-provider-schemas.sh succeeds on current repository"

echo "ALL SCHEMA COMPARISON TESTS PASSED."
