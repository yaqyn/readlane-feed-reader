package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFetchRSS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "readlane" {
			t.Error("missing user agent")
		}
		fmt.Fprint(w, `<rss><channel><title>A &amp;amp; B</title><item><title>&amp;ldquo;Hello&amp;rdquo;</title><link>/post#part</link><description>&lt;p&gt;Summary&lt;/p&gt;</description></item></channel></rss>`)
	}))
	defer server.Close()
	f, err := fetchFeed(context.Background(), server.URL+"/feed")
	if err != nil {
		t.Fatal(err)
	}
	if f.Channel.Title != "A & B" || f.Channel.Item[0].Title != "“Hello”" || f.Channel.Item[0].Link != server.URL+"/post" {
		t.Fatalf("unexpected feed: %+v", f)
	}
}
func TestParseAtom(t *testing.T) {
	b := []byte(`<feed xmlns="http://www.w3.org/2005/Atom"><title>Atom</title><entry><title>New post</title><link rel="self" href="/entry.xml"/><link rel="alternate" href="/post"/><content>&lt;p&gt;Content&lt;/p&gt;</content><updated>2026-10-14T10:00:00Z</updated></entry></feed>`)
	base, _ := url.Parse("https://example.com/feed.xml")
	f, err := parseFeed(b, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Channel.Item) != 1 || f.Channel.Item[0].Link != "https://example.com/post" || f.Channel.Item[0].PubDate != "2026-10-14T10:00:00Z" {
		t.Fatalf("unexpected Atom: %+v", f)
	}
}
func TestFetchRejectsBadResponses(t *testing.T) {
	for _, kind := range []string{"status", "malformed", "html", "oversized", "redirect-loop"} {
		t.Run(kind, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch kind {
				case "status":
					w.WriteHeader(503)
				case "malformed":
					fmt.Fprint(w, "<rss><")
				case "html":
					fmt.Fprint(w, "<html>not a feed</html>")
				case "oversized":
					fmt.Fprint(w, strings.Repeat("x", maxFeedBytes+1))
				case "redirect-loop":
					http.Redirect(w, r, "/loop", http.StatusFound)
				}
			}))
			defer s.Close()
			if _, err := fetchFeed(context.Background(), s.URL); err == nil {
				t.Fatal("accepted bad response")
			}
		})
	}
}
func TestFetchCancellationAndCredentialRedaction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := fetchFeed(ctx, "http://localhost:9/feed?token=private-value")
	if err == nil {
		t.Fatal("expected cancellation")
	}
	if strings.Contains(err.Error(), "private-value") {
		t.Fatal("error exposed URL query")
	}
}
func TestRejectUnsafeFeedURLs(t *testing.T) {
	for _, raw := range []string{"file:///etc/passwd", "javascript:alert(1)", "https://user:secret@example.com/feed", "/relative", "http://", "https://example.com/\x1b"} {
		if validateFeedURL(raw) == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}
func TestHTTPSRedirectCannotDowngrade(t *testing.T) {
	first, _ := http.NewRequest("GET", "https://example.com/feed", nil)
	next, _ := http.NewRequest("GET", "http://example.com/feed", nil)
	if feedClient.CheckRedirect(next, []*http.Request{first}) == nil {
		t.Fatal("accepted TLS downgrade")
	}
}
func TestPublishedDates(t *testing.T) {
	for _, raw := range []string{"Wed, 14 Oct 2026 10:00:00 +0200", "Wed, 14 Oct 2026 08:00:00 GMT", "2026-10-14T08:00:00Z", "Wed, 14 Oct 2026 03:00:00 EST"} {
		d := parsePublishedAt(raw)
		if !d.Valid || !d.Time.Equal(time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC)) {
			t.Errorf("bad date %q: %+v", raw, d)
		}
	}
	if parsePublishedAt("bad date").Valid {
		t.Fatal("invalid date became valid")
	}
}
func TestSafeTerminalSummary(t *testing.T) {
	raw := `<p>Hello <b>world</b></p><script>bad()</script><style>bad</style>` + "\x1b[2J\u202esecret"
	got := summary(raw)
	if strings.ContainsAny(got, "\x1b\u202e") || strings.Contains(got, "bad") || strings.Contains(got, "<") {
		t.Fatalf("unsafe summary %q", got)
	}
	if !strings.Contains(got, "Hello world") {
		t.Fatalf("missing summary text: %q", got)
	}
	if len([]rune(cleanText(strings.Repeat("a", 600), 10))) != 11 {
		t.Fatal("output not capped")
	}
}
func TestHelpAndVersionNeedNoConfig(t *testing.T) {
	t.Setenv("READLANE_CONFIG", "/nonexistent/config")
	for _, cmd := range []string{"help", "version"} {
		if err := runCLI(context.Background(), []string{cmd}); err != nil {
			t.Fatal(err)
		}
	}
}
func TestBadWorkerArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"0s"}, {"-1s"}, {"1s"}, {"bad"}, {"1m", "unknown"}} {
		if handlerAgg(nil, command{args: args}) == nil {
			t.Fatalf("accepted args %v", args)
		}
	}
}

func TestConditionalFetch(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"version-1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"version-1"`)
		w.Header().Set("Last-Modified", "Wed, 14 Oct 2026 08:00:00 GMT")
		fmt.Fprint(w, `<rss><channel><title>Cached</title></channel></rss>`)
	}))
	defer s.Close()
	feed, meta, err := fetchFeedConditional(context.Background(), s.URL, "", "")
	if err != nil || feed == nil {
		t.Fatalf("initial fetch: %v", err)
	}
	feed, updated, err := fetchFeedConditional(context.Background(), s.URL, meta.ETag, meta.LastModified)
	if err != nil || feed != nil || updated != meta {
		t.Fatalf("unchanged fetch: %+v %+v %v", feed, updated, err)
	}
}
func TestRetryAfterAndBackoff(t *testing.T) {
	now := time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC)
	cases := map[string]time.Duration{"120": 2 * time.Minute, "bad": 0, "-1": 0, "99999999": 24 * time.Hour, now.Add(3 * time.Minute).Format(http.TimeFormat): 3 * time.Minute}
	for raw, want := range cases {
		if got := retryAfter(raw, now); got != want {
			t.Errorf("retry after %q: got %s want %s", raw, got, want)
		}
	}
	if retryDelay(0) != time.Minute || retryDelay(1) != 2*time.Minute || retryDelay(30) != 256*time.Minute {
		t.Fatal("unexpected failure backoff")
	}
}
