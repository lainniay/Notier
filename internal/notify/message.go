package notify

import "strings"

// Parse converts an APP, TITLE, MSG Telegram message into a notification.
func Parse(text string) (Notifacation, bool) {
	parts := strings.SplitN(text, "\n\n", 3)
	if len(parts) != 3 {
		return Notifacation{}, false
	}
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return Notifacation{}, false
		}
	}
	title, appKey := app(parts[0])
	return Notifacation{
		Title:    title,
		Subtitle: parts[1],
		Body:     parts[2],
		AppKey:   appKey,
	}, true
}

func app(name string) (string, string) {
	switch name {
	case "qq":
		return "QQ", name
	case "wechat":
		return "WeChat", name
	case "wecom":
		return "WeCom", name
	case "mail":
		return "Mail", name
	case "sms":
		return "SMS", name
	default:
		return name, ""
	}
}
