package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"log/slog"
	"time"
)

const workerLockID int64 = 718421

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 || len(cmd.args) > 2 || (len(cmd.args) == 2 && cmd.args[1] != "--once") {
		return fmt.Errorf("usage: gator agg <interval> [--once]")
	}
	interval, err := time.ParseDuration(cmd.args[0])
	if err != nil || interval < 5*time.Second {
		return fmt.Errorf("interval must be at least 5s (recommended: 1m)")
	}
	return aggregate(s, interval, len(cmd.args) == 2)
}
func aggregate(s *state, interval time.Duration, once bool) error {
	conn, err := s.conn.Conn(s.ctx)
	if err != nil {
		return fmt.Errorf("open worker connection: %w", err)
	}
	defer conn.Close()
	lockCtx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	var locked bool
	if err = conn.QueryRowContext(lockCtx, "SELECT pg_try_advisory_lock($1)", workerLockID).Scan(&locked); err != nil {
		return fmt.Errorf("acquire worker lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("an aggregation worker is already running for this database")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", workerLockID)
	}()
	worker := *s
	worker.db = database.New(conn)
	slog.Info("collecting feeds", "interval", interval.String(), "once", once)
	// A timer restarted after each scrape guarantees an interval between requests,
	// even when a slow request would have left ticker events waiting in a queue.
	for {
		if s.ctx.Err() != nil {
			return nil
		}
		ctx, cancel := context.WithTimeout(s.ctx, 45*time.Second)
		worker.ctx = ctx
		err := scrapeFeeds(&worker)
		cancel()
		if s.ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return err
		}
		if once {
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-s.ctx.Done():
			timer.Stop()
			slog.Info("aggregation stopped")
			return nil
		case <-timer.C:
		}
	}
}
func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(s.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Info("no feeds to collect")
		return nil
	}
	if err != nil {
		return databaseError("select next feed", err)
	}
	// Record an attempt before the request so a broken feed cannot starve others.
	if err = s.db.MarkFeedFetched(s.ctx, feed.ID); err != nil {
		return databaseError("mark feed fetched", err)
	}
	slog.Info("fetching feed", "feed_id", feed.ID, "name", feed.Name)
	rss, err := fetchFeed(s.ctx, feed.Url)
	if err != nil {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}
		slog.Warn("feed request failed", "feed_id", feed.ID, "error", err)
		return nil
	}
	inserted := int64(0)
	for _, item := range rss.Channel.Item {
		if err := validateFeedURL(item.Link); err != nil {
			slog.Warn("skipping item without a valid URL", "feed_id", feed.ID)
			continue
		}
		n, err := savePost(s, feed.ID, item)
		if err != nil {
			return databaseError("save post", err)
		}
		inserted += n
	}
	slog.Info("feed collected", "feed_id", feed.ID, "items", len(rss.Channel.Item), "inserted", inserted)
	return nil
}
