// Package config provides Config File Manage
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const template = `app_id = 0
app_hash = ""
phone = ""
password = ""
target_peer_id = 0
notifier = "terminal-notifier"
`

type Config struct {
	AppID        int    `toml:"app_id"`
	AppHash      string `toml:"app_hash"`
	Phone        string `toml:"phone"`
	Password     string `toml:"password"`
	TargetPeerID int64  `toml:"target_peer_id"`
	Notifier     string `toml:"notifier"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Notifier == "" {
		cfg.Notifier = "terminal-notifier"
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func WriteTemplate(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(template), 0o600)
}

func (c Config) Validate() error {
	if c.AppID == 0 {
		return errors.New("app_id is required")
	}
	if c.AppHash == "" {
		return errors.New("app_hash is required")
	}
	if c.Phone == "" {
		return errors.New("phone is required")
	}
	return nil
}
