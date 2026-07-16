// Package notify encapsulation terminal-notifier(MVP)
package notify

import (
	"context"
	"fmt"
)

const TerminalNotifier = "terminal-notifier"

type Notifacation struct {
	Title    string
	Subtitle string
	Body     string
	AppKey   string
}

type Notifier interface {
	Notify(ctx context.Context, msg Notifacation) error
}

func New(name, appsDir string) (Notifier, error) {
	if name == "" {
		name = TerminalNotifier
	}
	switch name {
	case TerminalNotifier:
		apps, err := loadApps(appsDir, runCommand)
		if err != nil {
			return nil, err
		}
		return Terminal{Path: name, apps: apps, run: runContextCommand}, nil
	default:
		return nil, fmt.Errorf("unsupported notifier: %s", name)
	}
}
