// Package config persists the local user's Readlane settings.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const configFileName = ".readlane.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	if p := Env("CONFIG"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	current := filepath.Join(home, configFileName)
	if _, err := os.Stat(current); os.IsNotExist(err) {
		legacy := filepath.Join(home, ".gatorconfig.json")
		if _, err := os.Stat(legacy); err == nil {
			return legacy, nil
		}
	}
	return current, nil
}
func Read() (Config, error) {
	var cfg Config
	p, err := getConfigFilePath()
	if err != nil {
		return cfg, err
	}
	f, err := os.Open(p)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return cfg, err
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return cfg, fmt.Errorf("config must be private; run chmod 600 on your config file")
	}
	b, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if len(b) > 65536 {
		return cfg, fmt.Errorf("config exceeds 64 KiB")
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return Config{}, fmt.Errorf("config must contain one JSON object")
	}
	return cfg, nil
}
func (c *Config) SetUser(name string) error {
	next := *c
	next.CurrentUserName = name
	if err := write(next); err != nil {
		return err
	}
	*c = next
	return nil
}

// Save persists a private configuration using an atomic rename.
func Save(cfg Config) error { return write(cfg) }

func write(cfg Config) error {
	p, err := getConfigFilePath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(p), ".readlaneconfig-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), p)
}

// Env prefers the Readlane name and accepts the prior Gator setting during migration.
func Env(key string) string {
	if value := os.Getenv("READLANE_" + key); value != "" {
		return value
	}
	return os.Getenv("GATOR_" + key)
}
