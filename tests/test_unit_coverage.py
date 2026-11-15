import os
import pathlib
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]

class UnitCoverageTests(unittest.TestCase):
    def run_gate(self, coverage=True, fail=False):
        with tempfile.TemporaryDirectory(prefix='unit coverage ') as tmp:
            root = pathlib.Path(tmp)
            go = root / 'fake go'
            go.write_text('''#!/usr/bin/env python3
import os,sys,pathlib,json
with open(os.environ['CALLS'],'a') as f: f.write(json.dumps(sys.argv[1:])+'\\n')
if sys.argv[1]=='test':
 if os.environ.get('FAIL')=='1': sys.exit(23)
 for arg in sys.argv:
  if arg.startswith('-coverprofile='): pathlib.Path(arg.split('=',1)[1]).write_text('mode: set\\n')
else: print('total: (statements) 62.3%')
''')
            go.chmod(0o755)
            env = os.environ.copy()
            env.pop('SEERR_COVERAGE_DIR', None)
            env.update(GO_BIN=str(go), CALLS=str(root/'calls'), GITHUB_STEP_SUMMARY=str(root/'summary'), FAIL='1' if fail else '0')
            if coverage: env['SEERR_COVERAGE_DIR'] = str(root/'coverage with spaces')
            result = subprocess.run(['bash', str(ROOT/'scripts/test-unit-coverage.sh')], env=env, capture_output=True, text=True)
            summary = (root/'summary').read_text() if (root/'summary').exists() else ''
            return result.returncode, (root/'calls').read_text(), summary

    def test_disabled_preserves_unit_command(self):
        code, calls, summary = self.run_gate(coverage=False)
        self.assertEqual(code, 0)
        self.assertEqual(calls, '["test", "./..."]\n')
        self.assertEqual(summary, '')

    def test_profile_and_summary_paths_with_spaces(self):
        code, calls, summary = self.run_gate()
        self.assertEqual(code, 0)
        self.assertIn('-coverprofile=', calls)
        self.assertIn('62.3%', summary)
        self.assertEqual(len(calls.splitlines()), 2)

    def test_failed_tests_remain_failed_without_success_summary(self):
        code, calls, summary = self.run_gate(fail=True)
        self.assertEqual(code, 23)
        self.assertEqual(len(calls.splitlines()), 1)
        self.assertEqual(summary, '')

    def test_workflow_uploads_evidence_without_permission_expansion(self):
        workflow = (ROOT/'.github/workflows/test.yml').read_text()
        self.assertIn('SEERR_COVERAGE_DIR: test-artifacts/unit-coverage', workflow)
        self.assertIn('name: unit-test-coverage', workflow)
        self.assertIn('path: test-artifacts/unit-coverage/', workflow)
        self.assertIn('run bash "${repo_root}/scripts/test-unit-coverage.sh"', (ROOT/'scripts/test-all-locally.sh').read_text())
