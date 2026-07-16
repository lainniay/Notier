package notify

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestTerminalNotify_skips_notification_when_corresponding_app_is_running(t *testing.T) {
	tests := []struct {
		appKey      string
		processName string
	}{
		{appKey: "qq", processName: "QQ"},
		{appKey: "wechat", processName: "WeChat"},
		{appKey: "mail", processName: "Mail"},
	}

	for _, tt := range tests {
		t.Run(tt.appKey, func(t *testing.T) {
			// Given: the corresponding macOS app process is running.
			var calls [][]string
			terminal := Terminal{
				Path: TerminalNotifier,
				run: func(_ context.Context, name string, args ...string) error {
					calls = append(calls, append([]string{name}, args...))
					return nil
				},
			}

			// When: a matching Telegram notification is forwarded.
			err := terminal.Notify(context.Background(), Notifacation{AppKey: tt.appKey})
			// Then: only the process check runs; no notification command runs.
			if err != nil {
				t.Fatalf("Notify() error = %v", err)
			}
			want := [][]string{{pgrepPath, "-x", tt.processName}}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("commands = %#v, want %#v", calls, want)
			}
		})
	}
}

func TestTerminalNotify_forwards_notification_when_corresponding_app_is_not_running(t *testing.T) {
	// Given: the Mail process is not running.
	var calls [][]string
	terminal := Terminal{
		Path: TerminalNotifier,
		run: func(_ context.Context, name string, args ...string) error {
			calls = append(calls, append([]string{name}, args...))
			if name == pgrepPath {
				return errors.New("process not found")
			}
			return nil
		},
	}
	msg := Notifacation{Title: "Mail", Subtitle: "Subject", Body: "Body", AppKey: "mail"}

	// When: the notification is forwarded.
	err := terminal.Notify(context.Background(), msg)
	// Then: terminal-notifier runs after the process check.
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	want := [][]string{
		{pgrepPath, "-x", "Mail"},
		{TerminalNotifier, "-title", "Mail", "-subtitle", "Subject", "-message", "Body"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("commands = %#v, want %#v", calls, want)
	}
}

func TestTerminalCommand_routes_loaded_apps_to_native_executable(t *testing.T) {
	tests := []struct {
		name     string
		appKey   string
		wantPath string
	}{
		{name: "qq", appKey: "qq", wantPath: "/senders/NotierQQ.app/Contents/MacOS/NotierSender"},
		{name: "wechat", appKey: "wechat", wantPath: "/senders/NotierWeChat.app/Contents/MacOS/NotierSender"},
		{name: "wecom", appKey: "wecom", wantPath: "/senders/NotierWeCom.app/Contents/MacOS/NotierSender"},
		{name: "mail", appKey: "mail", wantPath: "/senders/NotierMail.app/Contents/MacOS/NotierSender"},
		{name: "sms", appKey: "sms", wantPath: "/senders/NotierSMS.app/Contents/MacOS/NotierSender"},
	}

	apps := map[string]string{
		"qq":     "/senders/NotierQQ.app/Contents/MacOS/NotierSender",
		"wechat": "/senders/NotierWeChat.app/Contents/MacOS/NotierSender",
		"wecom":  "/senders/NotierWeCom.app/Contents/MacOS/NotierSender",
		"mail":   "/senders/NotierMail.app/Contents/MacOS/NotierSender",
		"sms":    "/senders/NotierSMS.app/Contents/MacOS/NotierSender",
	}
	terminal := Terminal{Path: TerminalNotifier, apps: apps}
	wantArgs := []string{
		"--title", "Title",
		"--subtitle", "Subtitle",
		"--message", "Body",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Notifacation{Title: "Title", Subtitle: "Subtitle", Body: "Body", AppKey: tt.appKey}
			gotPath, gotArgs := terminal.command(msg)
			if gotPath != tt.wantPath {
				t.Fatalf("command() path = %q, want %q", gotPath, tt.wantPath)
			}
			if !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Fatalf("command() args = %#v, want %#v", gotArgs, wantArgs)
			}
		})
	}
}

func TestTerminalCommand_falls_back_to_terminal_notifier(t *testing.T) {
	tests := []struct {
		name string
		msg  Notifacation
	}{
		{name: "unknown app", msg: Notifacation{Title: "GitHub", Subtitle: "Title", Body: "Body", AppKey: "github"}},
		{name: "missing known app", msg: Notifacation{Title: "Mail", Subtitle: "Title", Body: "Body", AppKey: "mail"}},
	}
	terminal := Terminal{Path: TerminalNotifier, apps: map[string]string{
		"qq": "/senders/NotierQQ.app/Contents/MacOS/NotierSender",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotArgs := terminal.command(tt.msg)
			if gotPath != TerminalNotifier {
				t.Fatalf("command() path = %q, want %q", gotPath, TerminalNotifier)
			}
			wantArgs := []string{"-title", tt.msg.Title, "-subtitle", tt.msg.Subtitle, "-message", tt.msg.Body}
			if !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Fatalf("command() args = %#v, want %#v", gotArgs, wantArgs)
			}
		})
	}
}
