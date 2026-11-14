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
    def test_job_local_caches_support_real_build_without_home(self):
        workflow = (ROOT / '.github/workflows/test.yml').read_text()
        jobs = re.split(r'(?m)(?=^  [a-z][a-z0-9-]*:\n)', workflow.split('jobs:\n', 1)[1])
        jobs = [job for job in jobs if job.strip()]
        self.assertEqual(len(jobs), 8)
        for job in jobs:
            self.assertRegex(job, r'(?m)^    env:\n      GOMODCACHE: \$\{\{ runner\.temp \}\}/seerr-go-mod\n      GOCACHE: \$\{\{ runner\.temp \}\}/seerr-go-build$')
        go = shutil.which('go')
        self.assertIsNotNone(go, 'Go must be installed for this regression')
        with tempfile.TemporaryDirectory(prefix='go cache test ') as directory:
            root = pathlib.Path(directory)
            (root / 'go.mod').write_text('module cache-regression\n\ngo 1.20\n')
            (root / 'main.go').write_text('package main\nfunc main() {}\n')
            env = dict(os.environ)
            for key in ('HOME', 'USERPROFILE', 'GOPATH', 'XDG_CACHE_HOME'):
                env.pop(key, None)
            env.update(GOMODCACHE=str(root / 'seerr-go-mod'),
                       GOCACHE=str(root / 'seerr-go-build'),
                       GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off', CGO_ENABLED='0')
            subprocess.run([go, 'build', '-o', str(root / 'app'), '.'],
                           cwd=root, env=env, check=True, timeout=120)
            self.assertTrue((root / 'app').is_file())
            self.assertTrue((root / 'seerr-go-build').is_dir())

if __name__ == '__main__':
    unittest.main()
