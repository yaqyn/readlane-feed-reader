package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/yaqyn/readlane-feed-reader/internal/database"
)

// Launch keeps collection and reading in one process. Collector diagnostics go
// to a private file so background output cannot corrupt the terminal screen.
func handlerLaunch(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: readlane launch")
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(cache, "readlane")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(dir, "collector.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open collector log: %w", err)
	}
	defer logFile.Close()
	if err := logFile.Chmod(0600); err != nil {
		return err
	}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(logFile, nil)))
	defer slog.SetDefault(previous)
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	session := *s
	session.ctx = ctx
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := aggregate(&session, time.Minute, false); err != nil && ctx.Err() == nil {
			slog.Warn("collection stopped; another worker may already be running", "error", err)
		}
	}()
	err = handlerTerminal(&session, command{name: "tui"}, user)
	cancel()
	<-done
	return err
}
