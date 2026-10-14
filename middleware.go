package main

import (
	"fmt"
	"github.com/yaqyn/blog-aggregator/internal/database"
)

func middlewareLoggedIn(handler func(*state, command, database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("no current user; use gator login <username>")
		}
		user, err := s.db.GetUser(s.ctx, s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("current user not found; use gator login <username>: %w", err)
		}
		return handler(s, cmd, user)
	}
}
