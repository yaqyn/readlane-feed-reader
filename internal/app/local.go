package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yaqyn/readlane-feed-reader/internal/config"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
	"github.com/yaqyn/readlane-feed-reader/internal/migrations"
)

// Only Readlane's private cluster is managed automatically. A custom server is
// never started, reconfigured, or replaced after a failed connection.
func localRoots() (current, legacy string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		data = filepath.Join(home, ".local", "share")
	}
	current = config.Env("POSTGRES_ROOT")
	if current == "" {
		current = filepath.Join(data, "readlane")
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return "", "", err
	}
	return current, filepath.Join(data, "blog-aggregator"), nil
}

func defaultLocalURL() (string, error) {
	root, legacy, err := localRoots()
	if err != nil {
		return "", err
	}
	name := "readlane"
	if config.Env("POSTGRES_ROOT") == "" {
		if _, err := os.Stat(filepath.Join(root, "postgres", "PG_VERSION")); errors.Is(err, os.ErrNotExist) {
			if _, err := os.Stat(filepath.Join(legacy, "postgres", "PG_VERSION")); err == nil {
				root, name = legacy, "gator"
			}
		}
	}
	account, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("find local PostgreSQL user: %w", err)
	}
	port := config.Env("POSTGRES_PORT")
	if port == "" {
		port = "5433"
	}
	u := &url.URL{Scheme: "postgres", User: url.User(account.Username), Path: "/" + name}
	q := url.Values{"host": {root}, "port": {port}, "sslmode": {"disable"}}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func managedLocalRoot(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host != "" || (u.Path != "/readlane" && u.Path != "/gator") {
		return "", false
	}
	// Extra libpq routing options (hostaddr/service) make this a custom
	// connection even if its display host resembles the private socket.
	for key := range u.Query() {
		switch key {
		case "host", "port", "sslmode":
		default:
			return "", false
		}
	}
	if u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			return "", false
		}
	}
	root, legacy, err := localRoots()
	if err != nil {
		return "", false
	}
	host := u.Query().Get("host")
	if host == "" || !filepath.IsAbs(host) || (host != root && host != legacy) {
		return "", false
	}
	return host, true
}

func postgresTool(ctx context.Context, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("local reading needs PostgreSQL tools (%s missing); install PostgreSQL, then run readlane again", name)
	}
	cmd := exec.CommandContext(ctx, name, args...)
	// Startup diagnostics go to stderr; never print a configured connection URL.
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed; inspect the local postgres.log before retrying: %w", name, err)
	}
	return nil
}

func startLocalPostgres(ctx context.Context, raw string) error {
	root, ok := managedLocalRoot(raw)
	if !ok {
		return fmt.Errorf("database unavailable; check configuration, credentials and PostgreSQL service")
	}
	u, _ := url.Parse(raw)
	port := u.Query().Get("port")
	if port == "" {
		port = "5433"
	}
	// pg_ctl passes server options through a shell; keep the path single-quoted
	// and require a numeric port so neither value can become shell syntax.
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1024 || portNumber > 65535 || strings.ContainsAny(root, "'\\\n\r") {
		return fmt.Errorf("local PostgreSQL needs a port from 1024 to 65535 and a path without quotes, backslashes or line breaks")
	}
	port = strconv.Itoa(portNumber)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	data := filepath.Join(root, "postgres")
	if _, err := os.Stat(filepath.Join(data, "PG_VERSION")); errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "Preparing your private reading database…")
		if err := postgresTool(ctx, "initdb", "-D", data, "--auth-local=peer", "--auth-host=reject"); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := exec.LookPath("pg_ctl"); err != nil {
		return fmt.Errorf("local reading needs pg_ctl; install PostgreSQL, then run readlane again")
	}
	if exec.CommandContext(ctx, "pg_ctl", "-D", data, "status").Run() == nil {
		return nil
	}
	fmt.Fprintln(os.Stderr, "Starting your reading database…")
	options := fmt.Sprintf("-c listen_addresses='' -c unix_socket_directories='%s' -p %s", root, port)
	return postgresTool(ctx, "pg_ctl", "-D", data, "-l", filepath.Join(root, "postgres.log"), "-w", "-t", "15", "-o", options, "start")
}

func ensureLocalDatabase(ctx context.Context, raw string) error {
	if _, ok := managedLocalRoot(raw); !ok {
		return fmt.Errorf("refusing to create a database outside the managed local cluster")
	}
	u, _ := url.Parse(raw)
	name := strings.TrimPrefix(u.Path, "/") // managedLocalRoot restricts this to two fixed names.
	u.Path = "/postgres"
	admin, err := sql.Open("postgres", u.String())
	if err != nil {
		return err
	}
	defer admin.Close()
	var exists bool
	if err := admin.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return fmt.Errorf("connect to local PostgreSQL: %w", err)
	}
	if !exists {
		if _, err := admin.ExecContext(ctx, `CREATE DATABASE "`+name+`"`); err != nil {
			return fmt.Errorf("create local reading database: %w", err)
		}
	}
	return nil
}

func prepareLocalReader(s *state) error {
	if err := migrations.Up(s.ctx, s.conn); err != nil {
		return fmt.Errorf("prepare reading database: %w", err)
	}
	if s.cfg.CurrentUserName != "" {
		return nil
	}
	users, err := s.db.GetUsers(s.ctx)
	if err != nil {
		return databaseError("find readers", err)
	}
	if len(users) > 1 {
		return fmt.Errorf("multiple local readers exist; select one with readlane login <name> (readlane users lists them)")
	}
	if len(users) == 1 {
		return s.cfg.SetUser(users[0].Name)
	}
	name := "reader"
	if account, err := user.Current(); err == nil && validName(account.Username, 100) {
		name = account.Username
	}
	now := time.Now().UTC()
	reader, err := s.db.CreateUser(s.ctx, database.CreateUserParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: name})
	if err != nil {
		return databaseError("create local reader", err)
	}
	return s.cfg.SetUser(reader.Name)
}
