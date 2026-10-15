# Operating Readlane

Run the CLI and collector under a trusted OS account. User selection inside Readlane is a convenience, not authentication. Protect the config, backups, and database credentials with that account's filesystem permissions.

## Install and supervise the worker

Build from a reviewed checkout and run migrations before starting the worker:

```sh
make install
readlane migrate
readlane doctor
mkdir -p ~/.config/systemd/user
install -m 644 deploy/readlane.service ~/.config/systemd/user/readlane.service
systemctl --user daemon-reload
systemctl --user enable --now readlane.service
systemctl --user status readlane.service
journalctl --user -u readlane.service -f
```

The unit uses `~/.local/bin/readlane` and `~/.readlane.json`, collects one due feed per minute, emits JSON logs, and restarts after failures with a 15-second delay. A 256 MiB memory limit bounds process growth. It applies a read-only filesystem, no new privileges, private temporary files, address-family restrictions, and kernel/control-group protections. The worker reads configuration and writes to PostgreSQL; it does not need a writable home directory.

Linux user namespaces are required for the filesystem restrictions. If systemd reports a namespace failure, check the host's namespace support and policy; do not silently remove the hardening. The included unit was exercised on the development host with its restrictions enabled.

Normal user services stop when the user session ends. An operator can deliberately enable systemd user lingering if the worker must survive logout:

```sh
loginctl enable-linger "$USER"
```

Stop and disable collection:

```sh
systemctl --user disable --now readlane.service
```

A manual worker and the service cannot run together against the same database. Readlane uses a dedicated connection and a PostgreSQL advisory lock for the worker lifetime. A crash or lost connection releases that lock through PostgreSQL. If PostgreSQL becomes unavailable, the worker exits; systemd retries. Configuration failures therefore also appear as restart attempts—inspect the logs and fix the cause.

## Health and failure signals

- `readlane doctor` returns zero only when PostgreSQL is reachable and there are no pending embedded migrations.
- `readlane sources` lists retry timestamps and consecutive failure counts.
- `feed collected` logs identify the feed and report item and insertion counts.
- `feed unchanged` indicates HTTP 304 and a successful cache check.
- `feed request failed` includes the scheduled retry delay.
- `no feeds due` is normal when all feeds are cooling down or no feeds exist.

HTTP failures and malformed feeds are retried with persisted exponential backoff. `Retry-After` is respected up to 24 hours. Database failures stop collection. A process interrupted during storage may leave a partial batch; URL uniqueness makes the next attempt safe to repeat.

## Database roles and remote connections

The optional local helper uses OS-user peer authentication over a private Unix socket; it is intended for a single trusted operator. Its database role owns the local cluster. For a remotely hosted deployment, use a dedicated database and role, restrict network access, and require `sslmode=verify-full` with a trusted root certificate.

Run migrations using a schema owner. After applying them, a separate worker role can receive only the required access:

```sql
GRANT CONNECT ON DATABASE readlane TO readlane_worker;
GRANT USAGE ON SCHEMA public TO readlane_worker;
GRANT SELECT, UPDATE ON feeds TO readlane_worker;
GRANT INSERT, SELECT (url) ON posts TO readlane_worker;
```

The collector reads and updates feeds and inserts posts with application-generated UUIDs; no sequence grants are needed. The `SELECT (url)` grant is required by the explicit conflict target in `ON CONFLICT (url) DO NOTHING`. The worker role should not be used for registration, subscriptions, reset, migrations, or `doctor`. Use a separate private config/environment for the service, and an administrative CLI config for those operations. Never grant credentials to an untrusted user as a substitute for authentication.

To customize the service's config file, use a systemd override with `Environment=READLANE_CONFIG=/absolute/path/to/worker-config.json`. Protect the file with mode `600`. The worker's read-only filesystem requires that config to already exist and be readable by the service account.

## Back up and restore

Install the PostgreSQL client tools. Take backups before every schema upgrade and regularly while collecting. Keep backup files private and copy them to storage outside the database host.

Add PostgreSQL client service definitions to a private `~/.pg_service.conf` (`chmod 600`). For example, adapt these values to your database and restore target:

```ini
[readlane]
host=/var/run/postgresql
port=5432
dbname=readlane
user=your-user
sslmode=disable

[readlane_restore]
host=/var/run/postgresql
port=5432
dbname=readlane_restore_test
user=your-user
sslmode=disable
```

For the optional local helper, use its printed socket directory and port 5433. For remote databases, use certificate-verified TLS and supply credentials through a private `.pgpass`, service configuration, or secret store. A libpq service name avoids including the connection password in process arguments:


```sh
umask 077
backup_dir="$HOME/.local/share/readlane-backups"
mkdir -p "$backup_dir"
pg_dump --dbname=service=readlane --format=custom \
  --file "$backup_dir/readlane-$(date -u +%Y%m%dT%H%M%SZ).dump"
```

`pg_dump` provides a consistent snapshot while the worker runs. Test restores into a **separate empty database**, configured as the `readlane_restore` service:

```sh
set -o pipefail
pg_restore --no-owner --no-acl --file=- /absolute/path/to/readlane-backup.dump | \
  psql service=readlane_restore -X --set ON_ERROR_STOP=on --single-transaction
```

This pipes the restore SQL into `psql` using the private client service configuration. The single transaction avoids leaving a partially restored database on a SQL error. Do not test a restore on the live database. Check counts, subscriptions, and `readlane doctor` against the restored database before relying on a backup. Restore into a compatible PostgreSQL version; `pg_dump` must be at least as new as the source server.

A local backup/restore smoke check is recorded in [verification](verification.md). That check is not a replacement for your own scheduled backups and recovery exercises.

## Upgrade and recovery

1. Stop the worker and take a verified backup.
2. Build the intended Git revision with `make build`; retain the prior binary.
3. Install the new binary and run `readlane migrate` with schema-owner credentials.
4. Run `readlane doctor`, a single due-feed collection, and `readlane read`.
5. Restart the worker and inspect its logs.

The installed binary contains the migrations for its revision. Migrations never delete posts on `up`. Down migrations are developer recovery tools and can drop data; they are exercised only in isolated test schemas. For a failed production upgrade, prefer restoring a tested backup or applying a corrective migration rather than running destructive down migrations.

## Local PostgreSQL lifecycle

Opening `readlane` or `readlane tui` automatically initializes or restarts the
private local cluster, applies embedded migrations, and prepares a local reader.
This does not install or enable a PostgreSQL service. Closing the reader leaves
PostgreSQL running for other commands and collectors. Custom database servers
remain operator-managed. The optional helper remains available for manual setup.

The helper does not install or enable a PostgreSQL service. If it reused a
legacy cluster, substitute the printed `blog-aggregator` data-directory path
for `readlane` in the commands below. Its dedicated local cluster can be inspected and stopped with:

```sh
pg_ctl -D "$HOME/.local/share/readlane/postgres" status
pg_ctl -D "$HOME/.local/share/readlane/postgres" stop -m fast
```

Run `./scripts/start-local-postgres.sh` to start it again. Keep the data directory; removing it removes the database. PostgreSQL major upgrades require a planned `pg_upgrade` or dump/restore; do not point a new major version's server binary at an older data directory.

## Capacity and scope

The worker is intentionally sequential. More feeds increase the time between visits; monitor collection logs if that becomes unacceptable. Successful feeds wait at least one minute, and failures up to 256 minutes unless `Retry-After` asks for a longer delay. An interval below five seconds is rejected.

There is no automatic post retention, distributed queue, hosted API, or remote authentication. Feed URLs can access the worker host's local network; only trusted operators should add them. Offset browsing can shift while new posts arrive. These are explicit scope boundaries, not guarantees about an arbitrary hosting environment.
