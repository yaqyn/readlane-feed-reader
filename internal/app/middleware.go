package app

import (
	"fmt"
	"github.com/yaqyn/readlane/internal/database"
)

func middlewareLoggedIn(handler func(*state, command, database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("no current user; use readlane login <username>")
		}
		user, err := s.db.GetUser(s.ctx, s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("current user not found; use readlane login <username>")
		}
		return handler(s, cmd, user)
	}
}
