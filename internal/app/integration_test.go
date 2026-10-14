package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yaqyn/readlane/internal/config"
	"github.com/yaqyn/readlane/internal/database"
	"github.com/yaqyn/readlane/internal/migrations"
)

// Every integration test gets a separate schema in an explicitly named test DB.
// The database-name check prevents accidentally running fixtures on a real DB.
func integrationState(t *testing.T) *state {
	t.Helper()
	dsn := os.Getenv("READLANE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("READLANE_TEST_DATABASE_URL is unset")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	var name string
	if err = admin.QueryRowContext(ctx, "SELECT current_database()").Scan(&name); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	if !strings.HasSuffix(name, "_test") {
		admin.Close()
		t.Fatal("integration tests require a database whose name ends in _test")
	}
	schema := "readlane_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(c, "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE"); err != nil {
			t.Errorf("clean test schema: %v", err)
		}
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	conn, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(5)
	t.Cleanup(func() { conn.Close() })
	if err = migrations.Up(ctx, conn); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("READLANE_CONFIG", p)
	cfg := &config.Config{DBURL: u.String()}
	if err = cfg.SetUser(""); err != nil {
		t.Fatal(err)
	}
	return &state{cfg: cfg, conn: conn, db: database.New(conn), ctx: ctx}
}
func testUser(t *testing.T, s *state, name string) database.User {
	t.Helper()
	if err := handlerRegister(s, command{args: []string{name}}); err != nil {
		t.Fatal(err)
	}
	u, err := s.db.GetUser(s.ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
func TestIntegrationSubscriptionsAndUserIsolation(t *testing.T) {
	s := integrationState(t)
	alice := testUser(t, s, "alice")
	add := middlewareLoggedIn(handlerAddFeed)
	if err := add(s, command{args: []string{"Example", "https://example.com/feed"}}); err != nil {
		t.Fatal(err)
	}
	feed, err := s.db.GetFeedByURL(s.ctx, "https://example.com/feed")
	if err != nil {
		t.Fatal(err)
	}
	bob := testUser(t, s, "bob")
	n, err := savePost(s, feed.ID, RSSItem{Title: "Article", Link: "https://example.com/post", PubDate: "2026-10-14T08:00:00Z"})
	if err != nil || n != 1 {
		t.Fatalf("insert: %d %v", n, err)
	}
	n, err = savePost(s, feed.ID, RSSItem{Title: "Duplicate", Link: "https://example.com/post"})
	if err != nil || n != 0 {
		t.Fatalf("deduplicate: %d %v", n, err)
	}
	a, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: alice.ID, Limit: 10})
	if err != nil || len(a) != 1 {
		t.Fatalf("alice posts: %d %v", len(a), err)
	}
	b, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: bob.ID, Limit: 10})
	if err != nil || len(b) != 0 {
		t.Fatalf("bob received alice posts: %d %v", len(b), err)
	}
	follow := middlewareLoggedIn(handlerFollow)
	cmd := command{args: []string{feed.Url}}
	if err = follow(s, cmd); err != nil {
		t.Fatal(err)
	}
	if err = follow(s, cmd); err == nil {
		t.Fatal("duplicate subscription succeeded")
	}
	if err = middlewareLoggedIn(handlerUnfollow)(s, cmd); err != nil {
		t.Fatal(err)
	}
	b, err = s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: bob.ID, Limit: 10})
	if err != nil || len(b) != 0 {
		t.Fatalf("unfollow failed: %d %v", len(b), err)
	}
	if err = handlerRegister(s, command{args: []string{"alice"}}); err == nil {
		t.Fatal("duplicate user succeeded")
	}
	if err = handlerLogin(s, command{args: []string{"missing"}}); err == nil {
		t.Fatal("nonexistent login succeeded")
	}
	if s.cfg.CurrentUserName != "bob" {
		t.Fatal("failed commands changed current user")
	}
	if err = handlerReset(s, command{}); err == nil {
		t.Fatal("unguarded reset succeeded")
	}
	var count int
	if err = s.conn.QueryRowContext(s.ctx, "SELECT count(*) FROM users").Scan(&count); err != nil || count != 2 {
		t.Fatalf("unguarded reset deleted users: %d %v", count, err)
	}
	if err = handlerReset(s, command{args: []string{"--confirm"}}); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"users", "feeds", "feed_follows", "posts"} {
		if err = s.conn.QueryRowContext(s.ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("cascade %s: %d %v", table, count, err)
		}
	}
	if s.cfg.CurrentUserName != "" {
		t.Fatal("reset retained current user")
	}
}
func TestIntegrationAddFeedRollsBackOnFollowFailure(t *testing.T) {
	s := integrationState(t)
	testUser(t, s, "alice")
	// Simulate the dependent write failing after the feed INSERT.
	_, err := s.conn.ExecContext(s.ctx, "ALTER TABLE feed_follows ADD CONSTRAINT reject_fixture CHECK (false)")
	if err != nil {
		t.Fatal(err)
	}
	err = middlewareLoggedIn(handlerAddFeed)(s, command{args: []string{"Example", "https://example.com/feed"}})
	if err == nil {
		t.Fatal("follow failure was ignored")
	}
	_, err = s.db.GetFeedByURL(s.ctx, "https://example.com/feed")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("partial feed survived transaction: %v", err)
	}
}
func TestIntegrationWorkerStoresPostsAndCancels(t *testing.T) {
	s := integrationState(t)
	testUser(t, s, "alice")
	requested := make(chan struct{}, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested <- struct{}{}
		fmt.Fprint(w, `<rss><channel><item><title>First</title><link>/first</link><pubDate>Wed, 14 Oct 2026 08:00:00 GMT</pubDate></item><item><title>Unknown date</title><link>/second</link><pubDate>invalid</pubDate></item><item><title>Invalid URL</title><link>javascript:bad()</link></item></channel></rss>`)
	}))
	defer server.Close()
	if err := middlewareLoggedIn(handlerAddFeed)(s, command{args: []string{"Fixture", server.URL + "/feed"}}); err != nil {
		t.Fatal(err)
	}
	if err := aggregate(s, 5*time.Second, true); err != nil {
		t.Fatal(err)
	}
	makeFeedsDue(t, s)
	if err := aggregate(s, 5*time.Second, true); err != nil {
		t.Fatal(err)
	}
	posts, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: mustCurrentUser(t, s).ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 || posts[0].Title != "First" || !posts[0].PublishedAt.Valid || posts[1].PublishedAt.Valid {
		t.Fatalf("unexpected posts: %+v", posts)
	}
	if err := handlerBrowse(s, command{args: []string{"0"}}, mustCurrentUser(t, s)); err == nil {
		t.Fatal("invalid browse limit accepted")
	}
	page, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: mustCurrentUser(t, s).ID, Limit: 1, Offset: 1})
	if err != nil || len(page) != 1 || page[0].Title != "Unknown date" {
		t.Fatalf("pagination: %+v %v", page, err)
	}
	makeFeedsDue(t, s)
	// Drain the previous once requests before observing the continuous worker.
	for len(requested) > 0 {
		<-requested
	}
	ctx, cancel := context.WithCancel(s.ctx)
	worker := *s
	worker.ctx = ctx
	done := make(chan error, 1)
	go func() { done <- aggregate(&worker, time.Hour, false) }()
	select {
	case <-requested:
	case <-time.After(3 * time.Second):
		cancel()
		t.Fatal("worker did not fetch immediately")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop promptly")
	}
	// Cancellation must release the singleton advisory lock.
	if err := aggregate(s, 5*time.Second, true); err != nil {
		t.Fatal(err)
	}
}
func mustCurrentUser(t *testing.T, s *state) database.User {
	t.Helper()
	u, err := s.db.GetUser(s.ctx, s.cfg.CurrentUserName)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
func TestIntegrationWorkerSingletonAndFailedFeedFairness(t *testing.T) {
	s := integrationState(t)
	testUser(t, s, "alice")
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer bad.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<rss><channel><item><title>Good</title><link>/post</link></item></channel></rss>`)
	}))
	defer good.Close()
	for i, u := range []string{bad.URL, good.URL} {
		if err := middlewareLoggedIn(handlerAddFeed)(s, command{args: []string{fmt.Sprintf("Feed%d", i), u}}); err != nil {
			t.Fatal(err)
		}
	}
	conn, err := s.conn.Conn(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.ExecContext(s.ctx, "SELECT pg_advisory_lock($1)", workerLockID); err != nil {
		conn.Close()
		t.Fatal(err)
	}
	if err = aggregate(s, 5*time.Second, true); err == nil {
		t.Fatal("second worker acquired singleton lock")
	}
	if _, err = conn.ExecContext(s.ctx, "SELECT pg_advisory_unlock($1)", workerLockID); err != nil {
		t.Fatal(err)
	}
	conn.Close()
	// Two attempts must visit both feeds, including after a failed HTTP response.
	for range 2 {
		if err := aggregate(s, 5*time.Second, true); err != nil {
			t.Fatal(err)
		}
	}
	posts, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: mustCurrentUser(t, s).ID, Limit: 10})
	if err != nil || len(posts) != 1 {
		t.Fatalf("bad feed starved good feed: %+v %v", posts, err)
	}
	failed, err := s.db.GetFeedByURL(s.ctx, bad.URL)
	if err != nil {
		t.Fatal(err)
	}
	if failed.FailureCount != 1 || failed.LastError == "" || !failed.NextFetchAt.After(time.Now().Add(30*time.Second)) {
		t.Fatalf("retry state was not persisted: %+v", failed)
	}
	if _, err = s.db.GetNextFeedToFetch(s.ctx); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cooldown not enforced: %v", err)
	}

}
func TestIntegrationMigrationsReversibleAndIdempotent(t *testing.T) {
	s := integrationState(t)
	testUser(t, s, "retained")
	if err := migrations.Up(s.ctx, s.conn); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.GetUser(s.ctx, "retained"); err != nil {
		t.Fatal("idempotent migration lost data", err)
	}
	p, err := migrations.Provider(s.conn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.DownTo(s.ctx, 0); err != nil {
		t.Fatal(err)
	}
	if err = migrations.Up(s.ctx, s.conn); err != nil {
		t.Fatal(err)
	}
	if err = handlerDoctor(s, command{}); err != nil {
		t.Fatal(err)
	}
}

func makeFeedsDue(t *testing.T, s *state) {
	t.Helper()
	if _, err := s.conn.ExecContext(s.ctx, "UPDATE feeds SET next_fetch_at = NOW()"); err != nil {
		t.Fatal(err)
	}
}
