# Verification record

Local verification date: 2026-10-14. This record describes the checkout and development host checks, not a remotely deployed service.

## Course work

The 14 CLI lessons were run locally and submitted with explicit UUIDs using `bootdev run -s <UUID>`. All submissions reported **All tests passed!**. Aggregation, post storage, and documentation from the non-CLI lessons were also implemented. Their website completion questions and repository submissions remain the owner's responsibility.

| Lesson | Title | UUID | Outcome |
| --- | --- | --- | --- |
| 1.1 | Welcome to the Blog Aggregator | `14b7179b-ced3-4141-9fa5-e67dbc3e5242` | Passed locally and submitted with `bootdev run -s` |
| 1.2 | Config | `5e53c561-f059-49aa-b06c-5c5535fa6df0` | Passed locally and submitted with `bootdev run -s` |
| 1.3 | Commands | `dca1352a-7600-4d1d-bfdf-f9d741282e55` | Passed locally and submitted with `bootdev run -s` |
| 2.1 | Postgres | `74bea1f2-19cd-4ea9-966e-e2ca9dd1dfa9` | Passed locally and submitted with `bootdev run -s` |
| 2.2 | Goose Migrations | `d049abb2-c733-4986-a33a-73f05e2ea06c` | Passed locally and submitted with `bootdev run -s` |
| 2.3 | SQLC | `8279802c-a867-4ba6-9d06-25625bc42544` | Passed locally and submitted with `bootdev run -s` |
| 2.4 | Reset | `6619ebf8-44ab-4a2b-a536-0b17d116c15e` | Passed locally and submitted with `bootdev run -s` |
| 2.5 | List Users | `371be77c-711d-4072-8392-81732ed87512` | Passed locally and submitted with `bootdev run -s` |
| 3.1 | RSS | `7347666d-7967-4c77-84c5-a0306bee8d05` | Passed locally and submitted with `bootdev run -s` |
| 3.2 | Feeds | `f0126e90-414e-4a45-b6b6-758d59af012c` | Passed locally and submitted with `bootdev run -s` |
| 3.3 | List Feeds | `3c66635a-cf05-471e-8ad8-ff3a80a6b177` | Passed locally and submitted with `bootdev run -s` |
| 4.1 | Follow | `a5f72e6a-6af3-4568-9eb7-079a3809a46c` | Passed locally and submitted with `bootdev run -s` |
| 4.2 | Middleware | `dbc877bf-a777-416e-ac07-f6ca9559f48c` | Passed locally and submitted with `bootdev run -s` |
| 4.3 | Unfollow | `b1eb06af-f46e-40c1-a64f-836248122bb0` | Passed locally and submitted with `bootdev run -s` |
| 5.1 | Aggregate | `096ad14b-a863-4dcf-861d-9085bfc64cf9` | Implemented; website submission left to owner |
| 5.2 | Posts | `d391e27f-fbc9-4ca0-bc4c-d1a4f912bf16` | Implemented; website submission left to owner |
| 5.3 | Docs | `55d273f2-5d8f-4acc-9493-bf8be8aea700` | Implemented; website submission left to owner |
| 5.4 | Submit Your Repo! | `0851e2b5-0e12-4e09-a4b3-ce30c6dad33e` | Implemented; website submission left to owner |

The final app intentionally changes two intermediate lesson interfaces: `agg` now requires an interval and runs the real collector; `reset` requires `--confirm`. The intermediate `go run .` config-printing lesson becomes the proper CLI. Earlier tests describe incremental course snapshots and are not the final regression suite.

## Production verification

- Go 1.27.2, PostgreSQL 18.6, sqlc 1.31.1, and Goose library 3.28.0.
- Race-enabled unit and PostgreSQL integration tests passed, including isolated migrations, cascade deletion, subscription isolation, deduplication, transactional rollback, worker locking, cancellation, retry scheduling, and HTTP cache checks.
- `go vet ./...`, generated-query consistency, private configuration behavior, static builds, and `govulncheck ./...` checked locally. The vulnerability scan reported no known vulnerabilities in reachable code at the time of verification.
- Installed `gator` applied embedded migrations, passed `doctor`, collected a real feed, and displayed stored summaries.
- The systemd unit's sandbox and resource restrictions were exercised in a transient, single-collection service on the local host. The persistent worker was not enabled automatically.
- `make smoke` verified the static executable against fresh disposable databases: repeat migrations, registration, feed creation, collection, browsing, guarded reset, backup/transactional restore, and a worker with the documented restricted grants. Fixtures were removed afterward.
- `actionlint` validated the GitHub workflow; `bash -n` and an actual run checked the local PostgreSQL helper. The CI PostgreSQL image tag was verified against the registry. Local SVGs parsed successfully and Markdown file links resolved.
- `systemd-analyze --user verify` accepted the service after the executable was installed. Its transient smoke run exited successfully, with a measured memory peak of 21.1 MiB for one live-feed collection (not a capacity guarantee).

The GitHub workflow is prepared with pinned action revisions and PostgreSQL integration testing. The repository has not been pushed, so remote CI and remote deployment are unverified.
