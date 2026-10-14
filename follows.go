package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"time"
)

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: gator follow <url>")
	}
	user, err := s.db.GetUser(s.ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("login required: %w", err)
	}
	feed, err := s.db.GetFeedByURL(s.ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("find feed: %w", err)
	}
	now := time.Now().UTC()
	follow, err := s.db.CreateFeedFollow(s.ctx, database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return fmt.Errorf("follow feed: %w", err)
	}
	fmt.Printf("%s follows %s\n", follow.UserName, follow.FeedName)
	return nil
}
func handlerFollowing(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator following")
	}
	user, err := s.db.GetUser(s.ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("login required: %w", err)
	}
	follows, err := s.db.GetFeedFollowsForUser(s.ctx, user.ID)
	if err != nil {
		return fmt.Errorf("list follows: %w", err)
	}
	for _, f := range follows {
		fmt.Printf("* %s\n", f.FeedName)
	}
	return nil
}
