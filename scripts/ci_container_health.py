#!/usr/bin/env python3
"""Bounded, allowlisted container health metadata for failed disposable CI targets.

Never collect container environment, health-check output, names, or raw errors.
"""
import json
import re
import shutil
import subprocess
import sys
import time

BUDGET_SECONDS = 20
MAX_CONTAINERS = 16
STATE_FORMAT = ('{"status":{{json .State.Status}},"exit_code":{{.State.ExitCode}},'
                '"health":{{if .State.Health}}{{json .State.Health.Status}}{{else}}"none"{{end}}}')
STATUSES = {'created', 'running', 'paused', 'restarting', 'removing', 'exited', 'dead'}
HEALTH = {'none', 'starting', 'healthy', 'unhealthy'}


def collect(compose_file):
    deadline = time.monotonic() + BUDGET_SECONDS
    result = {'containers': [], 'warnings': []}

    def run(args):
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise subprocess.TimeoutExpired(args, BUDGET_SECONDS)
        return subprocess.run(args, capture_output=True, text=True,
                              timeout=remaining, check=True).stdout

    try:
        compose = ['docker-compose'] if shutil.which('docker-compose') else ['docker', 'compose']
        ids = run(compose + ['-f', compose_file, 'ps', '--all', '--quiet']).split()
        if len(ids) > MAX_CONTAINERS:
            result['warnings'].append('container_limit')
        for container_id in ids[:MAX_CONTAINERS]:
            if not re.fullmatch(r'[a-fA-F0-9]{12,64}', container_id):
                result['warnings'].append('invalid_container_id')
                continue
            try:
                state = json.loads(run(['docker', 'inspect', '--format', STATE_FORMAT, container_id]))
                if (not isinstance(state, dict) or state.get('status') not in STATUSES
                        or state.get('health') not in HEALTH
                        or type(state.get('exit_code')) is not int):
                    result['warnings'].append('invalid_state')
                    continue
                result['containers'].append({key: state[key] for key in ('status', 'exit_code', 'health')})
            except (json.JSONDecodeError, TypeError):
                result['warnings'].append('invalid_state')
            except subprocess.CalledProcessError:
                result['warnings'].append('inspect_failed')
    except subprocess.TimeoutExpired:
        result['warnings'].append('diagnostic_timeout')
    except (OSError, subprocess.CalledProcessError):
        result['warnings'].append('diagnostic_unavailable')
    return result


if __name__ == '__main__':
    if len(sys.argv) != 2:
        sys.exit('Usage: ci_container_health.py COMPOSE_FILE')
    print(json.dumps(collect(sys.argv[1]), sort_keys=True))
