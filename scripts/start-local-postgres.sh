#!/usr/bin/env bash
# An opt-in PostgreSQL cluster owned by the current OS user, with no TCP listener.
set -euo pipefail
cluster_root="${GATOR_POSTGRES_ROOT:-${XDG_DATA_HOME:-$HOME/.local/share}/blog-aggregator}"
cluster_port="${GATOR_POSTGRES_PORT:-5433}"
for tool in initdb pg_ctl createdb psql; do
  command -v "$tool" >/dev/null || { echo "Missing PostgreSQL tool: $tool" >&2; exit 1; }
done
umask 077
mkdir -p "$cluster_root"
chmod 700 "$cluster_root"
if [[ ! -f "$cluster_root/postgres/PG_VERSION" ]]; then
  initdb -D "$cluster_root/postgres" --auth-local=peer --auth-host=reject
fi
python3 - "$cluster_root" "$cluster_port" <<'PYCONF'
import pathlib, sys
root, raw_port = sys.argv[1:]
port = int(raw_port)
if not 1024 <= port <= 65535 or any(ord(c) < 32 for c in root):
    raise SystemExit('Choose a port from 1024 to 65535 and a path without control characters')
value = root.replace('\\', '\\\\').replace("'", "''")
config = pathlib.Path(root) / 'postgres/postgresql.conf'
include = "include_if_exists = 'gator-local.conf'"
if include not in config.read_text():
    with config.open('a') as f:
        f.write('\n' + include + '\n')
(config.parent / 'gator-local.conf').write_text("listen_addresses = ''\nunix_socket_directories = '" + value + "'\nport = " + str(port) + '\n')
PYCONF
if ! pg_ctl -D "$cluster_root/postgres" status >/dev/null; then
  pg_ctl -D "$cluster_root/postgres" -l "$cluster_root/postgres.log" start
fi
if [[ "$(psql -h "$cluster_root" -p "$cluster_port" -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname='gator'")" != 1 ]]; then
  createdb -h "$cluster_root" -p "$cluster_port" gator
fi
# JSON encoding correctly escapes paths and usernames without printing secrets.
python3 - "$cluster_root" "$cluster_port" <<'PY'
import json, getpass, sys, urllib.parse
query = urllib.parse.urlencode({'sslmode': 'disable', 'host': sys.argv[1], 'port': sys.argv[2]})
print(json.dumps({'db_url': 'postgres://' + urllib.parse.quote(getpass.getuser(), safe='') + '@/gator?' + query, 'current_user_name': ''}, indent=2))
PY
printf '\nCopy the JSON above to your private ~/.gatorconfig.json, then run gator migrate.\n'
