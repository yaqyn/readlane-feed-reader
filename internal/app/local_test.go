package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagedLocalBoundary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("READLANE_POSTGRES_ROOT", "")
	t.Setenv("GATOR_POSTGRES_ROOT", "")
	t.Setenv("READLANE_POSTGRES_PORT", "")
	t.Setenv("GATOR_POSTGRES_PORT", "")
	raw, err := defaultLocalURL()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(home, ".local/share/readlane")
	if got, ok := managedLocalRoot(raw); !ok || got != root {
		t.Fatalf("default not managed: %s %v", got, ok)
	}
	for _, raw := range []string{
		"postgres://user@localhost/readlane",
		"postgres://user@remote.example/readlane?host=" + root,
		"postgres:///other?host=" + root,
		"postgres:///readlane?host=" + root + "&hostaddr=192.0.2.1",
		"postgres:///readlane?host=/tmp/unrelated",
		"invalid",
	} {
		if _, ok := managedLocalRoot(raw); ok {
			t.Fatalf("custom database considered managed: %s", raw)
		}
	}
	legacy := filepath.Join(home, ".local/share/blog-aggregator/postgres")
	if err := os.MkdirAll(legacy, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "PG_VERSION"), []byte("18"), 0600); err != nil {
		t.Fatal(err)
	}
	raw, err = defaultLocalURL()
	if err != nil || !strings.Contains(raw, "/gator?") {
		t.Fatalf("legacy data not reused: %s %v", raw, err)
	}
	if _, ok := managedLocalRoot(raw); !ok {
		t.Fatal("legacy cluster not managed")
	}
}
