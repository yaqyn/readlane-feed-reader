package main

import (
	"fmt"
	"github.com/yaqyn/blog-aggregator/internal/config"
	"os"
)

type state struct{ cfg *config.Config }
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
	s := &state{cfg: &cfg}
	c := commands{handlers: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	return c.run(s, command{name: os.Args[1], args: os.Args[2:]})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gator:", err)
		os.Exit(1)
	}
}
