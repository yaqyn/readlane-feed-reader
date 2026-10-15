# Readlane in the terminal

The terminal is Readlane's primary interface. Use the interactive reader for
browsing; use commands for setup, source management, collection, and scripts.

## Open your reading list

`readlane` opens the reader when stdin and stdout are terminals. With redirected
output it prints help. `readlane tui` explicitly requests the interactive view
and reports a useful error when no terminal is attached.

The top bar identifies the selected local reader. The Reading list view shows
collected stories, their source, and publication date. The selected story has a
pointer and, when color is enabled, a violet highlight. Color is supplemental;
`NO_COLOR` leaves the layout and selection marker usable.

Use ↑/↓ or j/k to move. Enter opens a scrollable feed summary, with the original
URL shown below it. Esc returns to the same list position. q closes Readlane;
Ctrl+C also restores the previous terminal screen.

Press / to search titles and stored feed content. Results update after a short
pause, and superseded results cannot replace a newer query. Enter applies the
query; Esc ends editing without discarding it. Ctrl+U clears it. While typing,
letters remain text rather than navigation shortcuts.

Tab opens the Sources view. Choose a source with ↑/↓ and Enter. The current
search stays active. a returns to all followed sources. List pages contain 20
stories; n/p or arrow/page keys move between pages. r refreshes after a collector
has stored new posts. The reader is intentionally not an automatic collector.

## Command reference

| Command | Result |
| --- | --- |
| `readlane help` / `version` | Help or build version; no database required |
| `readlane migrate` | Apply embedded migrations |
| `readlane doctor` | Check the database and pending migrations |
| `readlane register <name>` | Create and select a local reader |
| `readlane login <name>` | Select an existing reader; not authentication |
| `readlane users` | List readers and mark the current one |
| `readlane addfeed <name> <url>` | Create a source and follow it in one transaction |
| `readlane sources` | List sources, creators, and failed-feed retry schedules |
| `readlane follow <url>` | Follow an existing source |
| `readlane following` | List the current reader's subscriptions |
| `readlane unfollow <url>` | Remove a subscription |
| `readlane preview <url>` | Inspect RSS or Atom without a database |
| `readlane collect <interval> [--once]` | Attempt one due source per cycle; interval must be at least 5s |
| `readlane read [limit] [offset]` | Plain posts; default limit 5, maximum 100, maximum offset 1,000,000 |
| `readlane tui` | Open the keyboard-driven reader |
| `readlane serve [--addr address] [--user name]` | Start the optional loopback-only web companion |
| `readlane reset --confirm` | Delete every reader, source, subscription, and post |

The collector waits for the interval after each attempt, and respects persisted
cooldowns and retries. `collect 1m --once` can report no due feeds; that is not an
error. It attempts one eligible feed, not every feed in the database.

`agg`, `browse`, `feeds`, and `fetch` alias `collect`, `read`, `sources`, and
`preview`. Commands return nonzero on failure. Human-readable output remains
plain when redirected; no JSON-output contract is currently provided.

## Recover from common problems

- **No config:** create a private `~/.readlane.json` using the database helper's
  output or `config.example.json`, then run `readlane migrate`.
- **Database unavailable:** start PostgreSQL and check your URL, role, and socket.
- **Missing schema:** run `readlane migrate` before opening either reader.
- **Empty reading list:** check `readlane following`, collect a due feed, then
  refresh. Following a source does not download it immediately.
- **A collector is already running:** inspect its terminal or service before
  starting another. Both the old and new app use the same database lock.
- **Small terminal:** use at least 40 columns and 14 rows, or `readlane read`.

Use `reset --confirm` only when intentionally discarding this entire database's
application data. The operation has no undo; backups belong in your operating
routine rather than in an interactive confirmation dialog.
