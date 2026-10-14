package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"strconv"
	"time"
)

func parsePublishedAt(raw string) sql.NullTime {
	for _, format := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RFC3339Nano, time.RFC850, time.ANSIC, "Mon, 2 Jan 2006 15:04:05 -0700", "2006-01-02T15:04:05-0700", "2006-01-02"} {
		if t, err := time.Parse(format, raw); err == nil {
			return sql.NullTime{Time: t.UTC(), Valid: true}
		}
	}
	return sql.NullTime{}
}
func savePost(s *state, feedID uuid.UUID, item RSSItem) (int64, error) {
	now := time.Now().UTC()
	return s.db.CreatePost(s.ctx, database.CreatePostParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Title: item.Title, Url: item.Link, Description: item.Description, PublishedAt: parsePublishedAt(item.PubDate), FeedID: feedID})
}
func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 2 {
		return fmt.Errorf("usage: gator browse [limit] [offset]")
	}
	limit, offset := 2, 0
	var err error
	if len(cmd.args) > 0 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil || limit < 1 || limit > 100 {
			return fmt.Errorf("limit must be between 1 and 100")
		}
	}
	if len(cmd.args) > 1 {
		offset, err = strconv.Atoi(cmd.args[1])
		if err != nil || offset < 0 || offset > 1000000 {
			return fmt.Errorf("offset must be between 0 and 1000000")
		}
	}
	posts, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return fmt.Errorf("browse posts: %w", err)
	}
	if len(posts) == 0 {
		fmt.Println("No posts found. Follow a feed and run gator agg 1m.")
		return nil
	}
	for _, p := range posts {
		date := "unknown date"
		if p.PublishedAt.Valid {
			date = p.PublishedAt.Time.UTC().Format(time.RFC3339)
		}
		fmt.Printf("\n%s\n%s · %s\n%s\n%s\n", p.Title, p.FeedName, date, p.Url, p.Description)
	}
	return nil
}
