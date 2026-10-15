# Moving from Gator to Readlane

The product and Go module are now **Readlane** (`github.com/yaqyn/readlane`).
The executable is `readlane`, its command package is `cmd/readlane`, and the
collector service is `deploy/readlane.service`. This checkout directory remains
`blog-aggregator`; no remote repository has been created or renamed.

## Keep your data

There is no database rename or destructive data migration in this rebrand.
Your existing database URL continues to work, including a database called
`gator`. Apply the new, non-destructive search index with `readlane migrate`.

Config lookup prefers `READLANE_CONFIG`, then `GATOR_CONFIG`. Without either,
`~/.readlane.json` takes priority; if it is absent, the existing
`~/.gatorconfig.json` is used in place. Reader changes update whichever file was
selected. The fallback does not copy credentials or create competing configs.

If you want the new filename, copy the private config deliberately:

```sh
cp -p ~/.gatorconfig.json ~/.readlane.json
chmod 600 ~/.readlane.json
readlane doctor
```

Do not overwrite an existing Readlane config that points to another database.
`READLANE_DATABASE_URL` and `READLANE_LOG_FORMAT` take precedence over their
`GATOR_` equivalents. The new test variable is `READLANE_TEST_DATABASE_URL`.

## Collector and database helper

Stop an old collector before enabling the new service. Both versions keep the
same PostgreSQL advisory-lock key, so they cannot accidentally collect in
parallel against one database. Install the new binary and service using the
[operations guide](operations.md); the new service is not enabled automatically.

The local database helper defaults to `~/.local/share/readlane`, but reuses the
existing `~/.local/share/blog-aggregator` cluster and `gator` database when it
finds that prior setup. Custom root and port variables accept the old names as
fallbacks. The helper prints the actual connection URL it selected.

## Commands

Prefer `collect`, `read`, `sources`, and `preview`. The prior `agg`, `browse`,
`feeds`, and `fetch` names still work as aliases. Source records, subscriptions,
post IDs, and stored articles retain their existing values. `reset --confirm`
remains explicitly destructive and is not part of migration.
