import importlib.util
import json
import pathlib
import subprocess
import unittest
from unittest import mock

ROOT = pathlib.Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location('health', ROOT / 'scripts/ci_container_health.py')
health = importlib.util.module_from_spec(spec)
spec.loader.exec_module(health)
ID = 'a' * 64


class HealthTests(unittest.TestCase):
    def collect(self, outputs, times=None):
        with mock.patch.object(health.shutil, 'which', return_value=None), mock.patch.object(health.subprocess, 'run') as run:
            run.side_effect = [subprocess.CompletedProcess([], 0, stdout=x) if isinstance(x, str) else x for x in outputs]
            if times is None:
                result = health.collect('/tmp/compose.yml')
            else:
                with mock.patch.object(health.time, 'monotonic', side_effect=times):
                    result = health.collect('/tmp/compose.yml')
            return result, run.call_args_list

    def test_allowlist_excludes_sensitive_fields(self):
        result, calls = self.collect([ID, json.dumps({'status': 'exited', 'exit_code': 137, 'health': 'unhealthy', 'env': 'SECRET', 'output': 'SECRET'})])
        self.assertEqual(result['containers'], [{'status': 'exited', 'exit_code': 137, 'health': 'unhealthy'}])
        self.assertNotIn('SECRET', json.dumps(result))
        self.assertIn('--format', calls[1].args[0])
        self.assertNotIn('.Config', health.STATE_FORMAT)
        self.assertNotIn('.Log', health.STATE_FORMAT)

    def test_invalid_identifier_never_inspected(self):
        result, calls = self.collect(['--env-secret'])
        self.assertEqual(len(calls), 1)
        self.assertEqual(result['warnings'], ['invalid_container_id'])

    def test_raw_failure_not_disclosed(self):
        result, _ = self.collect([subprocess.CalledProcessError(1, ['docker'], stderr='SECRET')])
        self.assertEqual(result['warnings'], ['diagnostic_unavailable'])
        self.assertNotIn('SECRET', json.dumps(result))

    def test_timeout_budget_is_shared(self):
        result, calls = self.collect([ID, subprocess.TimeoutExpired(['docker'], 1)], [100, 101, 119])
        self.assertEqual([x.kwargs['timeout'] for x in calls], [19, 1])
        self.assertEqual(result['warnings'], ['diagnostic_timeout'])

    def test_expired_budget_does_not_start_inspect(self):
        result, calls = self.collect([ID], [100, 101, 121])
        self.assertEqual(len(calls), 1)
        self.assertEqual(result['warnings'], ['diagnostic_timeout'])

    def test_invalid_state_is_redacted(self):
        for state in ['not json SECRET', '[]', '{"status":"SECRET","health":"healthy","exit_code":0}']:
            result, _ = self.collect([ID, state])
            self.assertEqual(result['warnings'], ['invalid_state'])
            self.assertNotIn('SECRET', json.dumps(result))

    def test_container_limit(self):
        state = '{"status":"running","health":"healthy","exit_code":0}'
        result, calls = self.collect([' '.join([ID] * 20)] + [state] * 16)
        self.assertEqual(len(calls), 17)
        self.assertEqual(len(result['containers']), 16)
        self.assertEqual(result['warnings'], ['container_limit'])

    def test_failure_capture_and_ci_wiring(self):
        script = (ROOT / 'scripts/test-integration.sh').read_text()
        self.assertIn('ci_container_health.py', script)
        self.assertIn('exit $exit_code', script)
        self.assertIn('test_ci_container_health.py', (ROOT / '.github/workflows/test.yml').read_text())


class ResourceSnapshotTests(unittest.TestCase):
    def test_snapshot_allowlists_numeric_fields(self):
        from types import SimpleNamespace
        disk = SimpleNamespace(f_blocks=100, f_frsize=4096, f_bavail=30, f_favail=50)
        text = 'MemTotal: 1024 kB\nMemAvailable: 12 kB\nSwapTotal: 512 kB\nSwapFree: 256 kB\nSecretToken: 999 kB\nMemFree: 99 kB\n'
        with mock.patch.object(health.os, 'cpu_count', return_value=2), mock.patch.object(health.os, 'getloadavg', return_value=(1., 2., 3.)), mock.patch.object(health.os, 'statvfs', return_value=disk), mock.patch.object(health.pathlib.Path, 'read_text', return_value=text):
            result = health.resource_snapshot()
        self.assertEqual(result, {'cpu_count': 2, 'load_average': [1., 2., 3.], 'workspace_bytes_total': 409600, 'workspace_bytes_available': 122880, 'workspace_inodes_available': 50, 'memory_bytes': {'MemTotal': 1048576, 'MemAvailable': 12288, 'SwapTotal': 524288, 'SwapFree': 262144}})
        self.assertNotIn('Secret', json.dumps(result))

    def test_unavailable_metrics_do_not_break_collection(self):
        with mock.patch.object(health.os, 'cpu_count', return_value=None), mock.patch.object(health.os, 'getloadavg', side_effect=OSError), mock.patch.object(health.os, 'statvfs', side_effect=OSError), mock.patch.object(health.pathlib.Path, 'read_text', side_effect=PermissionError):
            self.assertEqual(health.resource_snapshot(), {})

    def test_invalid_memory_values_not_copied(self):
        text = 'MemTotal: SECRET kB\nSwapFree: -1 kB\nMemAvailable: 2 MB\n'
        with mock.patch.object(health.pathlib.Path, 'read_text', return_value=text):
            self.assertNotIn('memory_bytes', health.resource_snapshot())
