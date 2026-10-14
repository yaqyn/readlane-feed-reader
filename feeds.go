package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"strings"
	"time"
)

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 || strings.TrimSpace(cmd.args[0]) == "" || len(cmd.args[0]) > 200 {
		return fmt.Errorf("usage: gator addfeed <name> <url>")
	}
	if err := validateFeedURL(cmd.args[1]); err != nil {
		return err
	}
	user, err := s.db.GetUser(s.ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("login required: %w", err)
	}
	now := time.Now().UTC()
	feed, err := s.db.CreateFeed(s.ctx, database.CreateFeedParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID})
	if err != nil {
		return fmt.Errorf("create feed: %w", err)
	}
	fmt.Printf("Added %s\nURL: %s\nID: %s\nUser: %s\n", feed.Name, feed.Url, feed.ID, user.Name)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator feeds")
	}
	feeds, err := s.db.GetFeeds(s.ctx)
	if err != nil {
		return fmt.Errorf("list feeds: %w", err)
	}
	for _, feed := range feeds {
		fmt.Printf("* %s\n  %s\n  Added by %s\n", feed.Name, feed.Url, feed.UserName)
	}
	return nil
}
