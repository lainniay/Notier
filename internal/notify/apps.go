package notify

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	plutilPath     = "/usr/bin/plutil"
	lsregisterPath = "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
)

type commandRunner func(name string, args ...string) ([]byte, error)

type senderApp struct {
	key        string
	path       string
	executable string
}

var senderApps = [...]struct {
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

func loadApps(appsDir string, run commandRunner) (map[string]string, error) {
	present := make([]senderApp, 0, len(senderApps))
	for _, app := range senderApps {
		appPath := filepath.Join(appsDir, app.filename)
		if _, err := os.Stat(appPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("%s: inspect bundle: %w", appPath, err)
		}

		plistPath := filepath.Join(appPath, "Contents", "Info.plist")
		output, err := run(plutilPath, "-extract", "CFBundleIdentifier", "raw", "-o", "-", plistPath)
		if err != nil {
			return nil, fmt.Errorf("%s: read bundle identifier: %w", appPath, err)
		}
		if bundleID := strings.TrimSpace(string(output)); bundleID != app.bundleID {
			return nil, fmt.Errorf("%s: bundle identifier %q, want %q", appPath, bundleID, app.bundleID)
		}

		executablePath := filepath.Join(appPath, "Contents", "MacOS", "NotierSender")
		info, err := os.Stat(executablePath)
		if err != nil {
			return nil, fmt.Errorf("%s: inspect executable: %w", executablePath, err)
		}
		if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
			return nil, fmt.Errorf("%s: not an executable file", executablePath)
		}
		present = append(present, senderApp{key: app.key, path: appPath, executable: executablePath})
	}

	executables := make(map[string]string, len(present))
	for _, app := range present {
		if _, err := run(lsregisterPath, "-f", app.path); err != nil {
			return nil, fmt.Errorf("%s: register bundle: %w", app.path, err)
		}
		executables[app.key] = app.executable
	}
	return executables, nil
}

func runCommand(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}
