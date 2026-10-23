package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
	"github.com/yaqyn/readlane-feed-reader/internal/web"
)

const readerPageSize = 12

type readerPost struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Source      string     `json:"source"`
	SourceID    string     `json:"source_id"`
	URL         string     `json:"url"`
	Summary     string     `json:"summary"`
	PublishedAt *time.Time `json:"published_at"`
}
type readerSource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func readerAddress(raw string) (string, error) {
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		return "", fmt.Errorf("use a loopback address and port, for example 127.0.0.1:8090")
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	n, err := strconv.Atoi(port)
	if ip == nil || !ip.IsLoopback() || err != nil || n < 0 || n > 65535 {
		return "", fmt.Errorf("the companion reader only accepts loopback addresses")
	}
	return net.JoinHostPort(host, port), nil
}
func handlerServe(s *state, cmd command) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	addr := flags.String("addr", "127.0.0.1:8090", "loopback address")
	name := flags.String("user", s.cfg.CurrentUserName, "local reader")
	if err := flags.Parse(cmd.args); err != nil || flags.NArg() != 0 {
		return fmt.Errorf("usage: readlane serve [--addr 127.0.0.1:8090] [--user name]")
	}
	bind, err := readerAddress(*addr)
	if err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("select a local reader with readlane login <name>, or use serve --user <name>")
	}
	user, err := s.db.GetUser(s.ctx, *name)
	if err != nil {
		return databaseError("find reader", err)
	}
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return fmt.Errorf("cannot start companion reader; check the address and whether its port is in use")
	}
	server := &http.Server{Handler: newReaderHandler(s, user), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 64 << 10}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-s.ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		case <-done:
		}
	}()
	fmt.Printf("%s\nhttp://%s\n\nReading as %s. Keep readlane collect running separately.\nPress Ctrl+C to close the reader.\n", cliStyle("Readlane is ready", "1;38;5;99"), listener.Addr(), cleanText(user.Name, 100))
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func newReaderHandler(s *state, user database.User) http.Handler {
	mux := http.NewServeMux()
	files, _ := fs.Sub(web.Assets, "assets")
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(files))))
	shell := func(w http.ResponseWriter, r *http.Request) {
		body, err := fs.ReadFile(files, "reader.html")
		if err != nil {
			http.Error(w, "Reader assets unavailable", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/posts/") {
			w.WriteHeader(http.StatusNotFound)
		}
		_, _ = w.Write(body)
	}
	mux.HandleFunc("GET /", shell)
	mux.HandleFunc("GET /api/reader", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			readerJSON(w, 400, map[string]string{"error": "Reading-list filters are not valid URL encoding."})
			return
		}
		allowed := map[string]bool{"q": true, "source": true, "page": true, "page_size": true}
		for name, entries := range values {
			if !allowed[name] || len(entries) > 1 {
				readerJSON(w, 400, map[string]string{"error": "Unknown or duplicate reading-list filter."})
				return
			}
		}
		query := strings.TrimSpace(values.Get("q"))
		if len([]rune(query)) > 200 {
			readerJSON(w, 400, map[string]string{"error": "Search must contain 200 characters or fewer."})
			return
		}
		source := uuid.NullUUID{}
		if raw := values.Get("source"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				readerJSON(w, 400, map[string]string{"error": "Choose a valid source."})
				return
			}
			source = uuid.NullUUID{UUID: id, Valid: true}
		}
		page := 1
		if raw := values.Get("page"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 83334 {
				readerJSON(w, 400, map[string]string{"error": "Choose a valid reading-list page."})
				return
			}
			page = n
		}
		pageSize := readerPageSize
		if raw := values.Get("page_size"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 100 {
				readerJSON(w, 400, map[string]string{"error": "Choose a page size from 1 to 100."})
				return
			}
			pageSize = n
		}
		// Repeatable read keeps counts and page contents in the same snapshot.
		tx, err := s.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
		if err != nil {
			readerFailure(w, err)
			return
		}
		defer tx.Rollback()
		q := s.db.WithTx(tx)
		count, err := q.CountReaderPosts(ctx, database.CountReaderPostsParams{UserID: user.ID, SourceID: source, Search: query})
		if err != nil {
			readerFailure(w, err)
			return
		}
		maxPage := max(1, int((count+int64(pageSize)-1)/int64(pageSize)))
		page = min(page, maxPage)
		rows, err := q.GetReaderPosts(ctx, database.GetReaderPostsParams{UserID: user.ID, SourceID: source, Search: query, PageOffset: int32((page - 1) * pageSize), PageLimit: int32(pageSize)})
		if err != nil {
			readerFailure(w, err)
			return
		}
		sources, err := q.GetReaderSources(ctx, user.ID)
		if err != nil {
			readerFailure(w, err)
			return
		}
		if err = tx.Commit(); err != nil {
			readerFailure(w, err)
			return
		}
		posts := make([]readerPost, 0, len(rows))
		for _, row := range rows {
			posts = append(posts, toReaderPost(row.ID, row.Title, row.FeedName, row.FeedID, row.Url, row.Description, row.PublishedAt, 500))
		}
		sourceList := make([]readerSource, 0, min(len(sources), 200))
		for _, row := range sources[:min(len(sources), 200)] {
			sourceList = append(sourceList, readerSource{ID: row.ID.String(), Name: cleanText(row.Name, 200)})
		}
		readerJSON(w, 200, map[string]any{"posts": posts, "sources": sourceList, "sources_truncated": len(sources) > 200, "total": count, "page": page, "page_size": pageSize, "reader": cleanText(user.Name, 100)})
	})
	mux.HandleFunc("GET /api/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			readerJSON(w, 404, map[string]string{"error": "This post is no longer in your reading list."})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		row, err := s.db.GetReaderPost(ctx, database.GetReaderPostParams{UserID: user.ID, ID: id})
		if errors.Is(err, sql.ErrNoRows) {
			readerJSON(w, 404, map[string]string{"error": "This post is no longer in your reading list."})
			return
		}
		if err != nil {
			readerFailure(w, err)
			return
		}
		readerJSON(w, 200, toReaderPost(row.ID, row.Title, row.FeedName, row.FeedID, row.Url, row.Description, row.PublishedAt, 12000))
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; font-src 'self'; img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			readerJSON(w, 403, map[string]string{"error": "Open this reader using its localhost address."})
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Scheme != "http" || u.Host != r.Host {
				readerJSON(w, 403, map[string]string{"error": "Cross-origin requests are not supported."})
				return
			}
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			readerJSON(w, 403, map[string]string{"error": "Cross-site requests are not supported."})
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			readerJSON(w, 405, map[string]string{"error": "The companion reader is read-only."})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/reader" && !strings.HasPrefix(r.URL.Path, "/api/posts/") {
			readerJSON(w, 404, map[string]string{"error": "This reader endpoint does not exist."})
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func toReaderPost(id uuid.UUID, title, source string, sourceID uuid.UUID, link, description string, date sql.NullTime, limit int) readerPost {
	if validateFeedURL(link) != nil {
		link = ""
	}
	post := readerPost{ID: id.String(), Title: cleanText(title, 300), Source: cleanText(source, 200), SourceID: sourceID.String(), URL: link, Summary: plainText(description, limit)}
	if post.Title == "" {
		post.Title = "Untitled post"
	}
	if date.Valid {
		t := date.Time.UTC()
		post.PublishedAt = &t
	}
	return post
}
func readerJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func readerFailure(w http.ResponseWriter, err error) {
	slog.Error("reader database request failed", "error", databaseError("read posts", err))
	readerJSON(w, 503, map[string]string{"error": "Your reading list is temporarily unavailable. Check PostgreSQL, then try again."})
}
