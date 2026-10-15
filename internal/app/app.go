package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/yaqyn/readlane-feed-reader/internal/config"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
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
		return fmt.Errorf("unknown command %q; use readlane help", cleanText(cmd.name, 100))
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
	c.register("collect", handlerAgg)
	c.register("fetch", handlerFetch)
	c.register("preview", handlerFetch)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("sources", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
	c.register("read", middlewareLoggedIn(handlerBrowse))
	c.register("migrate", handlerMigrate)
	c.register("doctor", handlerDoctor)
	c.register("serve", handlerServe)
	c.register("tui", middlewareLoggedIn(handlerTerminal))
	return c
}
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 || !validName(cmd.args[0], 100) {
		return fmt.Errorf("usage: readlane login <username>")
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

const usage = `Readlane
Your feeds. Your pace.

Usage: readlane <command> [arguments]

Start here
  tui                             Open the interactive terminal reader
  migrate                         Prepare or update the database
  register <name>                 Create and select a local reader
  addfeed <name> <url>            Add a source and follow it
  collect <interval> [--once]     Collect due posts (try: 1m)
  read [limit] [offset]           Read your posts (default: 5)
  serve [--addr address]          Open the local web reader

Your sources
  sources                         List sources and retry status
  follow <url>                    Follow a source already added
  following                       Show the sources you follow
  unfollow <url>                  Stop following a source
  preview <url>                   Preview a feed without saving it

Your workspace
  login <name>                    Switch the current local reader
  users                           List readers
  doctor                          Check database and migrations
  reset --confirm                 Delete all readers, sources and posts
  version                         Show the installed version
  help                            Show this guide

Config: ~/.readlane.json (READLANE_CONFIG overrides it)
Database: READLANE_DATABASE_URL overrides the configured connection.
Local reader selection is not authentication.

Existing command names agg, browse, feeds and fetch are also supported.
`

func runCLI(ctx context.Context, args []string) error {
	if len(args) == 0 {
		if terminalAvailable() {
			args = []string{"tui"}
		} else {
			fmt.Print(usage)
			return nil
		}
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(cliStyle("Readlane", "1;38;5;99") + strings.TrimPrefix(usage, "Readlane"))
		return nil
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Println("readlane", Version)
		return nil
	}
	c := registeredCommands()
	if _, ok := c.handlers[args[0]]; !ok {
		return fmt.Errorf("unknown command %q; use readlane help", cleanText(args[0], 100))
	}
	cmd := command{name: args[0], args: args[1:]}
	if cmd.name != "agg" && cmd.name != "collect" && cmd.name != "serve" && cmd.name != "tui" {
		timeout := 15 * time.Second
		if cmd.name == "fetch" || cmd.name == "preview" {
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
	if cmd.name == "fetch" || cmd.name == "preview" {
		return c.run(&state{ctx: ctx}, cmd)
	}
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	dbURL := cfg.DBURL
	if override := config.Env("DATABASE_URL"); override != "" {
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
	if config.Env("LOG_FORMAT") == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
	return runCLI(ctx, os.Args[1:])
}
func Main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "readlane:", cleanText(err.Error(), 1000))
		os.Exit(1)
	}
}
