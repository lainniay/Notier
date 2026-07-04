// Package paths provides XDG file paths.
package paths

import (
	"os"
	"path/filepath"
)

type Path struct {
	ConfigFile  string
	SessionFile string
	StdoutLog   string
	StderrLog   string
}

func DefaultDir() (Path, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return Path{}, err
	}

	stateDir, err := getStateDir()
	if err != nil {
		return Path{}, err
	}

	return Path{
		ConfigFile:  filepath.Join(configDir, "config.toml"),
		SessionFile: filepath.Join(stateDir, "session.json"),
		StdoutLog:   filepath.Join(stateDir, "stdout.log"),
		StderrLog:   filepath.Join(stateDir, "stderr.log"),
	}, nil
}

func getConfigDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" || !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "notier"), nil
}

func getStateDir() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" || !filepath.IsAbs(base) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "notier"), nil
}
