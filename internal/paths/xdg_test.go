package paths

import (
	"path/filepath"
	"testing"
)

func TestDefault_sets_apps_dir_under_user_applications(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	got, err := Default()
	if err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	want := filepath.Join(home, "Applications", "Notier Senders")
	if got.AppsDir != want {
		t.Fatalf("Default().AppsDir = %q, want %q", got.AppsDir, want)
	}
}
