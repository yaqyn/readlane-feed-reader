package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func configPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("READLANE_CONFIG", p)
	return p
}
func TestSetUserPersistsPrivateConfig(t *testing.T) {
	p := configPath(t)
	c := Config{DBURL: "postgres://user@localhost/readlane", CurrentUserName: "old"}
	if err := c.SetUser("new"); err != nil {
		t.Fatal(err)
	}
	got, err := Read()
	if err != nil {
		t.Fatal(err)
	}
	if got != c || got.CurrentUserName != "new" {
		t.Fatalf("unexpected config: %+v", got)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions: %o", info.Mode().Perm())
	}
}
func TestFailedWriteDoesNotChangeMemory(t *testing.T) {
	t.Setenv("READLANE_CONFIG", filepath.Join(t.TempDir(), "missing", "config.json"))
	c := Config{DBURL: "postgres://example", CurrentUserName: "old"}
	if err := c.SetUser("new"); err == nil {
		t.Fatal("expected write failure")
	}
	if c.CurrentUserName != "old" {
		t.Fatal("failed write changed current user")
	}
}
func TestRejectInvalidConfig(t *testing.T) {
	for _, raw := range []string{`{"db_url":`, `{"db_urll":"typo"}`, `{"db_url":"x"} {}`, strings.Repeat("x", 65537)} {
		t.Run(raw[:min(16, len(raw))], func(t *testing.T) {
			p := configPath(t)
			if err := os.WriteFile(p, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Read(); err == nil {
				t.Fatal("expected invalid config error")
			}
		})
	}
}
func TestRejectPublicConfig(t *testing.T) {
	p := configPath(t)
	if err := os.WriteFile(p, []byte(`{"db_url":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(); err == nil {
		t.Fatal("accepted world-readable config")
	}
}

func TestReadlaneEnvironmentPrecedesLegacy(t *testing.T) {
	t.Setenv("GATOR_CONFIG", "legacy.json")
	t.Setenv("READLANE_CONFIG", "current.json")
	if got := Env("CONFIG"); got != "current.json" {
		t.Fatalf("new environment lost precedence: %s", got)
	}
	t.Setenv("READLANE_CONFIG", "")
	if got := Env("CONFIG"); got != "legacy.json" {
		t.Fatalf("legacy environment not accepted: %s", got)
	}
}
func TestDefaultPathPreservesExistingConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("READLANE_CONFIG", "")
	t.Setenv("GATOR_CONFIG", "")
	legacy := filepath.Join(home, ".gatorconfig.json")
	if err := os.WriteFile(legacy, []byte(`{"db_url":"postgres://existing"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Read()
	if err != nil || cfg.DBURL != "postgres://existing" {
		t.Fatalf("legacy config failed: %+v %v", cfg, err)
	}
	current := filepath.Join(home, ".readlane.json")
	if err := os.WriteFile(current, []byte(`{"db_url":"postgres://current"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err = Read()
	if err != nil || cfg.DBURL != "postgres://current" {
		t.Fatalf("new config failed: %+v %v", cfg, err)
	}
}
