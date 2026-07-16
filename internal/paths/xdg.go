// Package paths provides XDG file paths.
package paths

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ConfigFile  string
	AppsDir     string
	SessionFile string
	StdoutLog   string
	StderrLog   string
}

func Default() (Paths, error) {
	configDir, err := getXdgDir("XDG_CONFIG_HOME", ".config")
	if err != nil {
		return Paths{}, err
	}

	stateDir, err := getXdgDir("XDG_STATE_HOME", filepath.Join(".local", "state"))
	if err != nil {
		return Paths{}, err
	}

	dataDir, err := getXdgDir("XDG_STATE_HOME", filepath.Join(".local", "share"))
	if err != nil {
		return Paths{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ConfigFile:  filepath.Join(configDir, "config.toml"),
		AppsDir:     filepath.Join(home, "Applications", "Notier Senders"),
		SessionFile: filepath.Join(dataDir, "session.json"),
		StdoutLog:   filepath.Join(stateDir, "stdout.log"),
		StderrLog:   filepath.Join(stateDir, "stderr.log"),
	}, nil
}

func getXdgDir(env, fallback string) (string, error) {
	base := os.Getenv(env)
	if base == "" || !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, fallback)
	}
	dir := filepath.Join(base, "notier")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
