package app

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"time"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 || !validName(cmd.args[0], 100) {
		return fmt.Errorf("usage: gator register <username> (1–100 characters)")
	}
	now := time.Now().UTC()
	user, err := s.db.CreateUser(s.ctx, database.CreateUserParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: cmd.args[0]})
	if err != nil {
		return databaseError("create user", err)
	}
	if err = s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("user created, but configuration was not updated: %w", err)
	}
	fmt.Printf("Created user %s (%s)\n", cleanText(user.Name, 100), user.ID)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 1 || cmd.args[0] != "--confirm" {
		return fmt.Errorf("reset deletes every user, feed and post; use gator reset --confirm")
	}
	if err := s.db.ResetUsers(s.ctx); err != nil {
		return databaseError("reset users", err)
	}
	if err := s.cfg.SetUser(""); err != nil {
		return fmt.Errorf("database reset, but config update failed: %w", err)
	}
	fmt.Println("Database reset")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator users")
	}
	users, err := s.db.GetUsers(s.ctx)
	if err != nil {
		return databaseError("list users", err)
	}
	for _, user := range users {
		suffix := ""
		if user.Name == s.cfg.CurrentUserName {
			suffix = " (current)"
		}
		fmt.Printf("* %s%s\n", cleanText(user.Name, 100), suffix)
	}
	return nil
}
