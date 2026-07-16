package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lainniay/Notier/internal/client"
	"github.com/lainniay/Notier/internal/config"
	"github.com/lainniay/Notier/internal/notify"
	"github.com/lainniay/Notier/internal/paths"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	xdg, err := paths.Default()
	if err != nil {
		return err
	}

	configFile := flag.String("config", xdg.ConfigFile, "config file path")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := config.WriteTemplate(*configFile); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "created config file %s; fill it and run again\n", *configFile)
			return nil
		}
		return err
	}

	notifier, err := notify.New(cfg.Notifier, xdg.AppsDir)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return client.Run(ctx, cfg, xdg.SessionFile, func(ctx context.Context, msg client.Message) error {
		notification, ok := notify.Parse(msg.Text)
		if !ok {
			return nil
		}
		return notifier.Notify(ctx, notification)
	})
}
