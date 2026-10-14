package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"time"
)

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: gator follow <url>")
	}
	feed, err := s.db.GetFeedByURL(s.ctx, cmd.args[0])
	if err != nil {
		return databaseError("find feed", err)
	}
	now := time.Now().UTC()
	follow, err := s.db.CreateFeedFollow(s.ctx, database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return databaseError("follow feed", err)
	}
	fmt.Printf("%s follows %s\n", cleanText(follow.UserName, 100), cleanText(follow.FeedName, 200))
	return nil
}
func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator following")
	}
	follows, err := s.db.GetFeedFollowsForUser(s.ctx, user.ID)
	if err != nil {
		return databaseError("list follows", err)
	}
	for _, f := range follows {
		fmt.Printf("* %s\n", cleanText(f.FeedName, 200))
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: gator unfollow <url>")
	}
	feed, err := s.db.GetFeedByURL(s.ctx, cmd.args[0])
	if err != nil {
		return databaseError("find feed", err)
	}
	count, err := s.db.DeleteFeedFollow(s.ctx, database.DeleteFeedFollowParams{UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return databaseError("unfollow feed", err)
	}
	if count == 0 {
		return fmt.Errorf("you are not following this feed")
	}
	fmt.Printf("Unfollowed %s\n", cleanText(feed.Name, 200))
	return nil
}
