package main

import (
	"fmt"
	"github.com/yaqyn/blog-aggregator/internal/migrations"
)

func handlerMigrate(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator migrate")
	}
	if err := migrations.Up(s.ctx, s.conn); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	fmt.Println("Database migrations applied")
	return nil
}
func handlerDoctor(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: gator doctor")
	}
	p, err := migrations.Provider(s.conn)
	if err != nil {
		return err
	}
	pending, err := p.HasPending(s.ctx)
	if err != nil {
		return fmt.Errorf("check migrations: %w", err)
	}
	if pending {
		return fmt.Errorf("pending database migrations; run gator migrate")
	}
	fmt.Println("Database reachable; schema up to date")
	return nil
}
