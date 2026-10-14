package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/yaqyn/blog-aggregator/internal/config"
	"github.com/yaqyn/blog-aggregator/internal/database"
	"os"
	"time"
)

type state struct {
	cfg  *config.Config
	db   *database.Queries
	conn *sql.DB
	ctx  context.Context
}
type command struct {
	name string
	args []string
}
type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) { c.handlers[name] = f }
func (c *commands) run(s *state, cmd command) error {
	f, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command %q", cmd.name)
	}
	return f(s, cmd)
}
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 || cmd.args[0] == "" {
		return fmt.Errorf("usage: gator login <username>")
	}
	if _, err := s.db.GetUser(s.ctx, cmd.args[0]); err != nil {
		return fmt.Errorf("user does not exist: %w", err)
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User set to %s\n", cmd.args[0])
	return nil
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("a command is required; use gator help")
	}
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer conn.Close()
	if err = conn.PingContext(ctx); err != nil {
		return fmt.Errorf("database unavailable (check configuration and migrations)")
	}
	s := &state{cfg: &cfg, db: database.New(conn), conn: conn, ctx: ctx}
	c := commands{handlers: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", handlerAddFeed)
	c.register("feeds", handlerFeeds)
	c.register("follow", handlerFollow)
	c.register("following", handlerFollowing)
	return c.run(s, command{name: os.Args[1], args: os.Args[2:]})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gator:", err)
		os.Exit(1)
	}
}
