package notify

import (
	"context"
	"os/exec"
)

const pgrepPath = "/usr/bin/pgrep"

type contextCommandRunner func(ctx context.Context, name string, args ...string) error

type Terminal struct {
	Path string
	apps map[string]string
	run  contextCommandRunner
}

func (t Terminal) Notify(ctx context.Context, msg Notifacation) error {
	run := t.run
	if run == nil {
		run = runContextCommand
	}

	var processName string
	switch msg.AppKey {
	case "qq":
		processName = "QQ"
	case "wechat":
		processName = "WeChat"
	case "mail":
		processName = "Mail"
	}
	if processName != "" && run(ctx, pgrepPath, "-x", processName) == nil {
		return nil
	}

	path, args := t.command(msg)
	return run(ctx, path, args...)
}

func (t Terminal) command(msg Notifacation) (string, []string) {
	if path, ok := t.apps[msg.AppKey]; ok {
		return path, []string{
			"--title", msg.Title,
			"--subtitle", msg.Subtitle,
			"--message", msg.Body,
		}
	}
	return t.Path, []string{
		"-title", msg.Title,
		"-subtitle", msg.Subtitle,
		"-message", msg.Body,
	}
}

func runContextCommand(ctx context.Context, name string, args ...string) error {
	return exec.CommandContext(ctx, name, args...).Run()
}
