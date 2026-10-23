package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
)

func readerRequest(handler http.Handler, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", target, nil)
	req.Host = "127.0.0.1:8090"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}
func TestLoopbackReaderAddress(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:8090", "[::1]:8090", "localhost:0"} {
		if _, err := readerAddress(addr); err != nil {
			t.Errorf("rejected local address %s: %v", addr, err)
		}
	}
	for _, addr := range []string{":8090", "0.0.0.0:8090", "example.com:8090", "192.168.1.1:8090", "127.0.0.1:99999"} {
		if _, err := readerAddress(addr); err == nil {
			t.Errorf("accepted public/invalid address %s", addr)
		}
	}
}
func TestReaderRejectsCrossSiteAndMutations(t *testing.T) {
	// Boundary checks must run before any database access.
	handler := newReaderHandler(nil, database.User{})
	cases := []struct {
		method, host, origin, site string
		want                       int
	}{
		{"GET", "evil.example", "", "", 403}, {"GET", "127.0.0.1:8090", "https://evil.example", "", 403},
		{"GET", "127.0.0.1:8090", "", "cross-site", 403}, {"POST", "localhost:8090", "", "", 405},
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, "/api/reader", nil)
		r.Host = c.host
		r.Header.Set("Origin", c.origin)
		r.Header.Set("Sec-Fetch-Site", c.site)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Errorf("boundary status %d want %d", w.Code, c.want)
		}
	}
	for _, path := range []string{"/", "/assets/reader.css", "/assets/reader.js", "/assets/fonts/manrope.woff2"} {
		w := readerRequest(handler, path)
		if w.Code != 200 {
			t.Fatalf("asset %s returned %d", path, w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
			t.Fatal("missing CSP")
		}
	}
	if w := readerRequest(handler, "/missing"); w.Code != 404 {
		t.Fatal("unknown page did not return 404")
	}
}
func TestIntegrationReaderScopeSearchAndErrors(t *testing.T) {
	s := integrationState(t)
	alice := testUser(t, s, "alice")
	if err := middlewareLoggedIn(handlerAddFeed)(s, command{args: []string{"Example", "https://example.com/feed"}}); err != nil {
		t.Fatal(err)
	}
	feed, err := s.db.GetFeedByURL(s.ctx, "https://example.com/feed")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = savePost(s, feed.ID, RSSItem{Title: "PostgreSQL patterns", Link: "https://example.com/post", Description: "<p>Useful reading</p><script>bad()</script>"}); err != nil {
		t.Fatal(err)
	}
	bob := testUser(t, s, "bob")
	aliceHandler := newReaderHandler(s, alice)
	bobHandler := newReaderHandler(s, bob)
	w := readerRequest(aliceHandler, "/api/reader?q=patterns&page=999")
	if w.Code != 200 {
		t.Fatalf("reader: %d %s", w.Code, w.Body.String())
	}
	var result struct {
		Posts    []readerPost `json:"posts"`
		Total    int          `json:"total"`
		Page     int          `json:"page"`
		PageSize int          `json:"page_size"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || result.Page != 1 || len(result.Posts) != 1 {
		t.Fatalf("search/pagination: %+v", result)
	}
	if w = readerRequest(aliceHandler, "/api/reader?page_size=1"); w.Code != 200 || !strings.Contains(w.Body.String(), `"page_size":1`) {
		t.Fatalf("custom page size: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(result.Posts[0].Summary, "bad") {
		t.Fatal("script text survived summary")
	}
	id := result.Posts[0].ID
	if w = readerRequest(aliceHandler, "/api/posts/"+id); w.Code != 200 {
		t.Fatal("could not read owned post")
	}
	if w = readerRequest(bobHandler, "/api/posts/"+id); w.Code != 404 {
		t.Fatal("post ID exposed another user's post")
	}
	if w = readerRequest(bobHandler, "/api/reader?source="+feed.ID.String()); w.Code != 200 || strings.Contains(w.Body.String(), "PostgreSQL") {
		t.Fatal("source filter leaked other user's post")
	}
	for _, path := range []string{"/api/reader?q=%ZZ", "/api/reader?page=-1", "/api/reader?page_size=0", "/api/reader?page_size=101", "/api/reader?page_size=bad", "/api/reader?source=invalid", "/api/reader?q=" + strings.Repeat("x", 201), "/api/reader?q=one&q=two", "/api/reader?unknown=value"} {
		if w = readerRequest(aliceHandler, path); w.Code != 400 {
			t.Errorf("accepted invalid query: %s", path)
		}
	}
	if w = readerRequest(aliceHandler, "/api/posts/"+uuid.NewString()); w.Code != 404 {
		t.Fatal("missing post did not return 404")
	}
	s.conn.Close()
	if w = readerRequest(aliceHandler, "/api/reader"); w.Code != 503 {
		t.Fatalf("database failure did not return 503: %d", w.Code)
	}
}
