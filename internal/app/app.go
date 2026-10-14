package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/yaqyn/blog-aggregator/internal/config"
	"github.com/yaqyn/blog-aggregator/internal/database"
)

var Version = "dev"

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
		return fmt.Errorf("unknown command %q; use gator help", cleanText(cmd.name, 100))
	}
	return f(s, cmd)
}
func registeredCommands() commands {
	c := commands{handlers: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("fetch", handlerFetch)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
	c.register("migrate", handlerMigrate)
	c.register("doctor", handlerDoctor)
	return c
}
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 || !validName(cmd.args[0], 100) {
		return fmt.Errorf("usage: gator login <username>")
	}
	if _, err := s.db.GetUser(s.ctx, cmd.args[0]); err != nil {
		return databaseError("find user", err)
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User set to %s\n", cleanText(cmd.args[0], 100))
	return nil
}

const usage = `gator — collect and browse RSS and Atom feeds

Usage: gator <command> [arguments]

  migrate                         Apply embedded database migrations
  doctor                          Check database connectivity and schema
  register <username>             Create and select a local user
  login <username>                Select an existing local user
  users                           List local users
  addfeed <name> <url>            Add a feed and follow it atomically
  feeds                           List feeds and their creators
  follow <url>                    Follow an existing feed
  following                       List your followed feeds
  unfollow <url>                  Unfollow a feed
  fetch <url>                     Preview a feed without saving it
  agg <interval> [--once]         Collect one feed per cycle (minimum 5s)
  browse [limit] [offset]         Read followed posts (default limit: 2)
  reset --confirm                 DELETE all users, feeds, follows and posts
  version                         Print the build version
  help                            Show this help

Config: ~/.gatorconfig.json (override with GATOR_CONFIG)
Database URL override: GATOR_DATABASE_URL
Local user selection is not authentication.
`

func runCLI(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a command is required; use gator help")
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(usage)
		return nil
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Println("gator", Version)
		return nil
	}
	c := registeredCommands()
	if _, ok := c.handlers[args[0]]; !ok {
		return fmt.Errorf("unknown command %q; use gator help", cleanText(args[0], 100))
	}
	cmd := command{name: args[0], args: args[1:]}
	if cmd.name != "agg" {
		timeout := 15 * time.Second
		if cmd.name == "fetch" {
			timeout = 35 * time.Second
		}
		if cmd.name == "migrate" {
			timeout = 2 * time.Minute
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	// Feed previews and help do not require an available database or config file.
	if cmd.name == "fetch" {
		return c.run(&state{ctx: ctx}, cmd)
	}
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	dbURL := cfg.DBURL
	if override := os.Getenv("GATOR_DATABASE_URL"); override != "" {
		dbURL = override
	}
	u, err := url.Parse(dbURL)
	if err != nil || u == nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Path == "" {
		return fmt.Errorf("db_url must be a PostgreSQL connection URL")
	}
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("invalid database connection configuration")
	}
	defer conn.Close()
	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(2)
	conn.SetConnMaxLifetime(30 * time.Minute)
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err = conn.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database unavailable; check configuration, credentials and PostgreSQL service")
	}
	return c.run(&state{cfg: &cfg, db: database.New(conn), conn: conn, ctx: ctx}, cmd)
}
func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if os.Getenv("GATOR_LOG_FORMAT") == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
	return runCLI(ctx, os.Args[1:])
}
func Main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gator:", cleanText(err.Error(), 1000))
		os.Exit(1)
	}
}
