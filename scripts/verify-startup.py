#!/usr/bin/env python3
"""Verify first launch and stopped-cluster recovery in a private disposable cluster."""
import fcntl
import json
import os
import pathlib
import pty
import select
import shutil
import struct
import subprocess
import sys
import tempfile
import termios
import time

binaries = [pathlib.Path(p).resolve() for p in (sys.argv[1:] or ['./readlane'])]
for tool in ('initdb', 'pg_ctl'):
    if not shutil.which(tool):
        raise SystemExit(f'Missing PostgreSQL tool: {tool}')

with tempfile.TemporaryDirectory(prefix='readlane-startup-') as directory:
    root = pathlib.Path(directory)
    env = {k: v for k, v in os.environ.items()
           if not k.startswith(('READLANE_', 'GATOR_', 'PG'))}
    env.update(READLANE_CONFIG=str(root / 'config.json'),
               XDG_DATA_HOME=str(root / 'data'), XDG_CACHE_HOME=str(root / 'cache'), TERM='xterm-256color',
               READLANE_REDUCED_MOTION='1')
    if any(binary.name == 'launch' for binary in binaries) and shutil.which('go'):
        env['GOCACHE'] = subprocess.check_output(['go', 'env', 'GOCACHE'], text=True).strip()
    data = root / 'data/readlane/postgres'

    def stop():
        if (data / 'PG_VERSION').exists():
            subprocess.run(['pg_ctl', '-D', str(data), '-m', 'fast', '-w', 'stop'],
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    try:
        help_result = subprocess.run([str(binaries[0])], env=env, capture_output=True, timeout=5)
        assert help_result.returncode == 0 and b'Usage:' in help_result.stdout
        assert not (root / 'config.json').exists(), 'piped help created configuration'
        for index, binary in enumerate(binaries):
            stop()  # Each invocation must recover a stopped cluster.
            master, slave = pty.openpty()
            fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack('HHHH', 24, 80, 0, 0))
            args = [str(binary)] + ([] if index == 0 or binary.name == 'launch' else ['tui'])
            process = subprocess.Popen(args, env=env, cwd=root, stdin=slave, stdout=slave, stderr=slave)
            os.close(slave)
            output = b''
            try:
                deadline = time.monotonic() + 45
                while time.monotonic() < deadline:
                    if select.select([master], [], [], .2)[0]:
                        try:
                            output += os.read(master, 65536)
                        except OSError:
                            break
                    if b'0 posts' in output:
                        break
                    if process.poll() is not None:
                        break
                assert b'0 posts' in output, output.decode(errors='replace')
                config = json.loads((root / 'config.json').read_text())
                assert config['current_user_name'], 'no reader selected'
                assert (root / 'config.json').stat().st_mode & 0o777 == 0o600
                doctor = subprocess.run([str(binary), 'doctor'], env=env, capture_output=True, timeout=10)
                assert doctor.returncode == 0, doctor.stderr.decode()
                os.write(master, b'q')
                assert process.wait(timeout=5) == 0
                if binary.name == 'launch':
                    log = root / 'cache/readlane/collector.log'
                    assert 'collecting feeds' in log.read_text(), 'collector did not start'
                    assert log.stat().st_mode & 0o777 == 0o600
                print(f'PASS: {binary.name} {"first launch" if index == 0 else "launcher + collector" if binary.name == "launch" else "tui restart"}, private config, reader, migrations, clean exit')
            finally:
                if process.poll() is None:
                    process.terminate()
                    process.wait(timeout=5)
                os.close(master)
    finally:
        stop()
