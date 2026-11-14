"""Offline regression tests for Compose installer runner environment handling."""
import pathlib
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]

class ComposeInstallerTests(unittest.TestCase):
    def test_install_and_next_step_config_discovery(self):
        source = (ROOT / '.github/actions/ensure-docker-compose/action.yml').read_text()
        script = '\n'.join(line[8:] for line in source.split('      run: |\n', 1)[1].splitlines())
        for mode in ('home', 'temp', 'config', 'empty'):
            with self.subTest(mode=mode), tempfile.TemporaryDirectory(prefix='compose test ') as directory:
                root = pathlib.Path(directory)
                bin_dir = root / 'bin'
                bin_dir.mkdir()
                (bin_dir / 'docker').write_text('#!/bin/bash\n[[ -f "${DOCKER_CONFIG:-${HOME:-}/.docker}/cli-plugins/docker-compose" ]]\n')
                (bin_dir / 'curl').write_text('#!/bin/bash\nwhile [[ $# -gt 0 ]]; do if [[ "$1" == -o ]]; then shift; touch "$1"; exit; fi; shift; done\nexit 1\n')
                for binary in bin_dir.iterdir():
                    binary.chmod(0o755)
                env = {'PATH': str(bin_dir) + ':/usr/bin:/bin', 'RUNNER_TEMP': directory,
                       'GITHUB_ENV': str(root / 'env'), 'COMPOSE_VERSION': 'v2.40.3'}
                if mode == 'home':
                    env['HOME'] = str(root / 'home')
                if mode == 'config':
                    env['DOCKER_CONFIG'] = str(root / 'custom')
                if mode == 'empty':
                    env.update(HOME='', DOCKER_CONFIG='')
                subprocess.run(['/bin/bash', '-c', script], env=env, check=True)
                config = root / {'home': 'home/.docker', 'temp': 'docker-config',
                                 'config': 'custom', 'empty': 'docker-config'}[mode]
                self.assertTrue((config / 'cli-plugins/docker-compose').exists())
                if mode != 'config':
                    self.assertEqual((root / 'env').read_text(), f'DOCKER_CONFIG={config}\n')
                    env['DOCKER_CONFIG'] = str(config)
                else:
                    self.assertFalse((root / 'env').exists())
                subprocess.run(['docker', 'compose', 'version'], env=env, check=True)

if __name__ == '__main__':
    unittest.main()
