# The optional web companion

Readlane is terminal-first. The companion provides a wider reading surface
for posts already collected by the terminal worker. It does not manage feeds,
switch the shared config's reader, start collection, or change stored data.

```sh
readlane serve
# another local port or an explicitly selected reader:
readlane serve --addr 127.0.0.1:8091 --user alice
```

Open the printed URL. The selected reader stays fixed for that server process.
Stop it with Ctrl+C. The server includes local fonts and all assets in the Go
binary; it makes no third-party browser requests for fonts or analytics.

## Read and find stories

The reading list shows 12 posts per page. Search matches collected titles and
feed content, and a source choice limits results to that subscription. Search,
source, and page are stored in the URL. Opening a feed summary preserves that
context; its Back link returns to the filtered reading list. Read original
opens the source article in a separate browser tab.

Search has an explicit clear action, 300ms debounce, composition-safe typing,
and cancellation/stale-response protection. Empty data, no matches, unavailable
posts, and server failures have distinct recovery paths. A failed refresh
retains currently displayed stories and offers Retry. The native source select
is intentionally platform-owned; up to 200 followed sources appear in the
companion, with the full registry available through `readlane sources`.

The layout stacks on narrow screens, provides a skip link and visible focus,
respects reduced motion and forced colors, and never injects feed HTML. Only
feed-supplied summaries are shown, bounded to 12,000 characters in detail.

## Local boundary

The listener accepts only loopback IPs or localhost, never `0.0.0.0` or a public
address. Requests with an unexpected Host, foreign Origin, or cross-site browser
context are rejected. Only GET and HEAD are accepted; the companion exposes no
write endpoint. Queries, direct post IDs, and source filters are scoped to the
selected reader's subscriptions.

This is not remote authentication. Do not expose the listener through a public
reverse proxy, tunnel, or forwarded port. A trusted local operator still owns
the database credentials and can access all data through other clients.

## Verification scope

The HTTP handlers and database scoping are integration-tested. DOM interaction
tests cover search, stale responses, composition, empty/error recovery, safe
text rendering, and post navigation. Axe checks semantic accessibility in that
DOM environment; color contrast is checked separately against design tokens.

The current browser controller reports `net::ERR_BLOCKED_BY_CLIENT` for both
loopback URL forms, despite successful local HTTP requests. Full browser visual,
narrow-viewport, and native-select popup checks are therefore not claimed as
passed. The terminal reader was exercised in a real PTY.
