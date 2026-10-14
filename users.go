package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"strings"
	"time"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 || strings.TrimSpace(cmd.args[0]) == "" || len(cmd.args[0]) > 100 {
		return fmt.Errorf("usage: gator register <username> (1–100 characters)")
	}
	now := time.Now().UTC()
	user, err := s.db.CreateUser(s.ctx, database.CreateUserParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: cmd.args[0]})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err = s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("user created, but configuration was not updated: %w", err)
	}
	fmt.Printf("Created user %s (%s)\n", user.Name, user.ID)
	return nil
}
