// Package client provides Telegram MTProto message listening.
package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"github.com/lainniay/Notier/internal/config"
)

type Message struct {
	PeerID int64
	Text   string
}

type Handler func(context.Context, Message) error

// Run starts a Telegram client, authenticates when needed, and forwards matching messages to handler.
func Run(ctx context.Context, cfg config.Config, sessionFile string, handler Handler) error {
	dispatcher := tg.NewUpdateDispatcher()

	client := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionFile},
		UpdateHandler:  dispatcher,
	})

	dispatcher.OnNewMessage(func(ctx context.Context, _ tg.Entities, u *tg.UpdateNewMessage) error {
		msg, ok := u.Message.(*tg.Message)
		if !ok || msg.Message == "" {
			return nil
		}

		id := peerID(msg.PeerID)
		if id == 0 {
			return nil
		}
		if cfg.TargetPeerID != 0 && id != cfg.TargetPeerID {
			return nil
		}

		return handler(ctx, Message{
			PeerID: id,
			Text:   msg.Message,
		})
	})

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, authFlow(cfg)); err != nil {
			return err
		}

		<-ctx.Done()
		return nil
	})
}

// authFlow builds the user login flow from config.
func authFlow(cfg config.Config) auth.Flow {
	code := auth.CodeAuthenticatorFunc(func(_ context.Context, _ *tg.AuthSentCode) (string, error) {
		fmt.Print("telegram code: ")
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(text), nil
	})

	if cfg.Password != "" {
		return auth.NewFlow(
			auth.Constant(cfg.Phone, cfg.Password, code),
			auth.SendCodeOptions{},
		)
	}

	return auth.NewFlow(
		auth.CodeOnly(cfg.Phone, code),
		auth.SendCodeOptions{},
	)
}

// peerID converts gotd peer variants to Telegram-style peer IDs.
func peerID(peer tg.PeerClass) int64 {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return p.UserID
	case *tg.PeerChat:
		return -p.ChatID
	case *tg.PeerChannel:
		return -1000000000000 - p.ChannelID
	default:
		return 0
	}
}
