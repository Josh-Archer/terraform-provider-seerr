"""Go builds must work on shared CI runners without HOME or GOPATH."""
import os
import pathlib
import re
import shutil
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]

class GoCacheEnvironmentTests(unittest.TestCase):
    def test_codeql_bootstraps_cache_before_go_setup_and_extraction(self):
        workflow = (ROOT / '.github/workflows/codeql.yml').read_text()
        bootstrap = workflow.index('uses: ./.github/actions/ensure-go-cache')
        self.assertLess(bootstrap, workflow.index('uses: actions/setup-go@'))
        self.assertLess(bootstrap, workflow.index('uses: github/codeql-action/init@'))
        self.assertIn("if: matrix.language == 'go'\n        uses: ./.github/actions/ensure-go-cache", workflow)

    def test_opentofu_uses_native_cli_without_node_wrapper(self):
        workflow = (ROOT / '.github/workflows/test.yml').read_text()
        setups = re.findall(r'uses: opentofu/setup-opentofu@[^\n]+\n        with:\n          tofu_version: [^\n]+\n          tofu_wrapper: false', workflow)
        self.assertEqual(len(setups), 6)

    def test_existing_caches_are_preserved(self):
        source = (ROOT / '.github/actions/ensure-go-cache/action.yml').read_text()
        script = '\n'.join(line[8:] for line in source.split('      run: |\n', 1)[1].splitlines())
        with tempfile.TemporaryDirectory() as directory:
            output = pathlib.Path(directory) / 'env'
            env = dict(os.environ, RUNNER_TEMP=directory, GITHUB_ENV=str(output),
                       GOMODCACHE='/configured module cache', GOCACHE='/configured build cache')
            subprocess.run(['/bin/bash', '-c', script], env=env, check=True)
            self.assertEqual(output.read_text(), 'GOMODCACHE=/configured module cache\nGOCACHE=/configured build cache\n')

    def test_job_local_caches_support_real_build_without_home(self):
        workflow = (ROOT / '.github/workflows/test.yml').read_text()
        self.assertNotIn('runner.temp', workflow)
        self.assertEqual(workflow.count('uses: ./.github/actions/ensure-go-cache'), 8)
        source = (ROOT / '.github/actions/ensure-go-cache/action.yml').read_text()
        script = '\n'.join(line[8:] for line in source.split('      run: |\n', 1)[1].splitlines())
        go = shutil.which('go')
        self.assertIsNotNone(go, 'Go must be installed for this regression')
        with tempfile.TemporaryDirectory(prefix='go cache test ') as directory:
            root = pathlib.Path(directory)
            (root / 'go.mod').write_text('module cache-regression\n\ngo 1.20\n')
            (root / 'main.go').write_text('package main\nfunc main() {}\n')
            env = dict(os.environ)
            for key in ('HOME', 'USERPROFILE', 'GOPATH', 'XDG_CACHE_HOME'):
                env.pop(key, None)
            for key in ('GOMODCACHE', 'GOCACHE'):
                env.pop(key, None)
            env.update(RUNNER_TEMP=str(root), GITHUB_ENV=str(root / 'github-env'),
                       GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off', CGO_ENABLED='0')
            subprocess.run(['/bin/bash', '-c', script], env=env, check=True)
            exported = dict(line.split('=', 1) for line in (root / 'github-env').read_text().splitlines())
            self.assertEqual(exported, {'GOMODCACHE': str(root / 'seerr-go-mod'),
                                        'GOCACHE': str(root / 'seerr-go-build')})
            env.update(exported)
            subprocess.run([go, 'build', '-o', str(root / 'app'), '.'],
                           cwd=root, env=env, check=True, timeout=120)
            self.assertTrue((root / 'app').is_file())
            self.assertTrue((root / 'seerr-go-build').is_dir())

if __name__ == '__main__':
    unittest.main()
