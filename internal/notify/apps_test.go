package notify

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var testSenderApps = [...]struct {
	key      string
	filename string
	bundleID string
}{
	{key: "qq", filename: "NotierQQ.app", bundleID: "com.notier.sender.qq"},
	{key: "wechat", filename: "NotierWeChat.app", bundleID: "com.notier.sender.wechat"},
	{key: "wecom", filename: "NotierWeCom.app", bundleID: "com.notier.sender.wecom"},
	{key: "mail", filename: "NotierMail.app", bundleID: "com.notier.sender.mail"},
	{key: "sms", filename: "NotierSMS.app", bundleID: "com.notier.sender.sms"},
}

type commandCall struct {
	name string
	args []string
}

func createTestSenderApp(t *testing.T, appsDir, filename string, mode os.FileMode) string {
	t.Helper()
	appPath := filepath.Join(appsDir, filename)
	executablePath := filepath.Join(appPath, "Contents", "MacOS", "NotierSender")
	if err := os.MkdirAll(filepath.Dir(executablePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executablePath, nil, mode); err != nil {
		t.Fatal(err)
	}
	return executablePath
}

func TestLoadApps_validates_all_present_bundles_before_registration(t *testing.T) {
	appsDir := t.TempDir()
	executables := make(map[string]string, len(testSenderApps))
	bundleIDs := make(map[string]string, len(testSenderApps))
	for _, app := range testSenderApps {
		executables[app.key] = createTestSenderApp(t, appsDir, app.filename, 0o755)
		bundleIDs[app.filename] = app.bundleID
	}

	var calls []commandCall
	run := func(name string, args ...string) ([]byte, error) {
		calls = append(calls, commandCall{name: name, args: append([]string(nil), args...)})
		if name == plutilPath {
			filename := filepath.Base(filepath.Dir(filepath.Dir(args[len(args)-1])))
			return []byte(bundleIDs[filename] + "\n"), nil
		}
		return nil, nil
	}

	got, err := loadApps(appsDir, run)
	if err != nil {
		t.Fatalf("loadApps() error = %v", err)
	}

	if !reflect.DeepEqual(got, executables) {
		t.Fatalf("loadApps() = %#v, want %#v", got, executables)
	}
	if len(calls) != len(testSenderApps)*2 {
		t.Fatalf("command calls = %d, want %d", len(calls), len(testSenderApps)*2)
	}
	for index, app := range testSenderApps {
		appPath := filepath.Join(appsDir, app.filename)
		plistPath := filepath.Join(appPath, "Contents", "Info.plist")
		wantValidation := commandCall{
			name: plutilPath,
			args: []string{"-extract", "CFBundleIdentifier", "raw", "-o", "-", plistPath},
		}
		if !reflect.DeepEqual(calls[index], wantValidation) {
			t.Fatalf("validation call %d = %#v, want %#v", index, calls[index], wantValidation)
		}
		wantRegistration := commandCall{name: lsregisterPath, args: []string{"-f", appPath}}
		if !reflect.DeepEqual(calls[len(testSenderApps)+index], wantRegistration) {
			t.Fatalf("registration call %d = %#v, want %#v", index, calls[len(testSenderApps)+index], wantRegistration)
		}
	}
}

func TestLoadApps_missing_bundles_fall_back_without_commands(t *testing.T) {
	called := false
	run := func(string, ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	got, err := loadApps(t.TempDir(), run)
	if err != nil {
		t.Fatalf("loadApps() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("loadApps() = %#v, want empty map", got)
	}
	if called {
		t.Fatal("loadApps() ran a command for missing bundles")
	}
}

func TestLoadApps_mismatched_bundle_prevents_all_registration(t *testing.T) {
	appsDir := t.TempDir()
	createTestSenderApp(t, appsDir, "NotierQQ.app", 0o755)
	createTestSenderApp(t, appsDir, "NotierWeChat.app", 0o755)

	registrations := 0
	run := func(name string, args ...string) ([]byte, error) {
		if name == lsregisterPath {
			registrations++
			return nil, nil
		}
		if strings.Contains(args[len(args)-1], "NotierWeChat.app") {
			return []byte("com.tencent.xinWeChat"), nil
		}
		return []byte("com.notier.sender.qq"), nil
	}

	_, err := loadApps(appsDir, run)
	if err == nil {
		t.Fatal("loadApps() error = nil, want mismatch error")
	}
	if !strings.Contains(err.Error(), filepath.Join(appsDir, "NotierWeChat.app")) {
		t.Fatalf("loadApps() error = %q, want app path", err)
	}
	if registrations != 0 {
		t.Fatalf("registrations = %d, want 0", registrations)
	}
}

func TestLoadApps_plutil_failure_is_path_qualified(t *testing.T) {
	appsDir := t.TempDir()
	createTestSenderApp(t, appsDir, "NotierMail.app", 0o755)
	appPath := filepath.Join(appsDir, "NotierMail.app")
	run := func(string, ...string) ([]byte, error) {
		return nil, errors.New("malformed plist")
	}

	_, err := loadApps(appsDir, run)
	if err == nil || !strings.Contains(err.Error(), appPath) {
		t.Fatalf("loadApps() error = %v, want error containing %q", err, appPath)
	}
}

func TestLoadApps_registration_failure_is_path_qualified(t *testing.T) {
	appsDir := t.TempDir()
	createTestSenderApp(t, appsDir, "NotierSMS.app", 0o755)
	appPath := filepath.Join(appsDir, "NotierSMS.app")
	run := func(name string, _ ...string) ([]byte, error) {
		if name == plutilPath {
			return []byte("com.notier.sender.sms"), nil
		}
		return nil, errors.New("registration failed")
	}

	_, err := loadApps(appsDir, run)
	if err == nil || !strings.Contains(err.Error(), appPath) {
		t.Fatalf("loadApps() error = %v, want error containing %q", err, appPath)
	}
}

func TestLoadApps_invalid_executable_prevents_all_registration(t *testing.T) {
	tests := []struct {
		name string
		make func(t *testing.T, appsDir string)
	}{
		{
			name: "missing",
			make: func(t *testing.T, appsDir string) {
				t.Helper()
				if err := os.Mkdir(filepath.Join(appsDir, "NotierQQ.app"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "not executable",
			make: func(t *testing.T, appsDir string) {
				t.Helper()
				createTestSenderApp(t, appsDir, "NotierQQ.app", 0o644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appsDir := t.TempDir()
			tt.make(t, appsDir)
			registrations := 0
			run := func(name string, _ ...string) ([]byte, error) {
				if name == lsregisterPath {
					registrations++
				}
				return []byte("com.notier.sender.qq"), nil
			}

			_, err := loadApps(appsDir, run)
			if err == nil || !strings.Contains(err.Error(), "NotierSender") {
				t.Fatalf("loadApps() error = %v, want executable path error", err)
			}
			if registrations != 0 {
				t.Fatalf("registrations = %d, want 0", registrations)
			}
		})
	}
}

func TestNew_rejects_unsupported_notifier_before_loading_apps(t *testing.T) {
	appsDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(appsDir, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := New("unsupported", appsDir)
	if err == nil || err.Error() != "unsupported notifier: unsupported" {
		t.Fatalf("New() error = %v, want unsupported notifier error", err)
	}
}
