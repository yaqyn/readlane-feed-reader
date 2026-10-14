package app

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/readlane/internal/database"
	"time"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 || !validName(cmd.args[0], 200) {
		return fmt.Errorf("usage: readlane addfeed <name> <url>")
	}
	if err := validateFeedURL(cmd.args[1]); err != nil {
		return err
	}
	now := time.Now().UTC()
	tx, err := s.conn.BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := s.db.WithTx(tx)
	feed, err := q.CreateFeed(s.ctx, database.CreateFeedParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID})
	if err != nil {
		return databaseError("create feed", err)
	}
	_, err = q.CreateFeedFollow(s.ctx, database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return databaseError("follow new feed", err)
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("Added %s\nURL: %s\nID: %s\nUser: %s\n", cleanText(feed.Name, 200), cleanText(feed.Url, 4096), feed.ID, cleanText(user.Name, 100))
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: readlane feeds")
	}
	feeds, err := s.db.GetFeeds(s.ctx)
	if err != nil {
		return databaseError("list feeds", err)
	}
	for _, feed := range feeds {
		fmt.Printf("* %s\n  %s\n  Added by %s\n", cleanText(feed.Name, 200), cleanText(feed.Url, 4096), cleanText(feed.UserName, 100))
		if feed.FailureCount > 0 {
			fmt.Printf("  Retry scheduled: %s (%d failures)\n", feed.NextFetchAt.UTC().Format(time.RFC3339), feed.FailureCount)
		}
	}
	return nil
}
