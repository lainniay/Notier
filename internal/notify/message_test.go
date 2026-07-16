package notify

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Notifacation
		ok   bool
	}{
		{
			name: "valid message",
			text: "GitHub\n\nBuild failed\n\nmain branch",
			want: Notifacation{Title: "GitHub", Subtitle: "Build failed", Body: "main branch"},
			ok:   true,
		},
		{name: "qq title", text: "qq\n\nTitle\n\nBody", want: Notifacation{Title: "QQ", Subtitle: "Title", Body: "Body", AppKey: "qq"}, ok: true},
		{name: "wechat title", text: "wechat\n\nTitle\n\nBody", want: Notifacation{Title: "WeChat", Subtitle: "Title", Body: "Body", AppKey: "wechat"}, ok: true},
		{name: "wecom title", text: "wecom\n\nTitle\n\nBody", want: Notifacation{Title: "WeCom", Subtitle: "Title", Body: "Body", AppKey: "wecom"}, ok: true},
		{name: "mail title", text: "mail\n\nTitle\n\nBody", want: Notifacation{Title: "Mail", Subtitle: "Title", Body: "Body", AppKey: "mail"}, ok: true},
		{name: "sms title", text: "sms\n\nTitle\n\nBody", want: Notifacation{Title: "SMS", Subtitle: "Title", Body: "Body", AppKey: "sms"}, ok: true},
		{
			name: "multiline body",
			text: "GitHub\n\nBuild failed\n\nline one\n\nline two\nline three",
			want: Notifacation{Title: "GitHub", Subtitle: "Build failed", Body: "line one\n\nline two\nline three"},
			ok:   true,
		},
		{name: "missing fields", text: "GitHub\n\nBuild failed"},
		{name: "empty app", text: "\n\nBuild failed\n\nmain branch"},
		{name: "empty title", text: "GitHub\n\n \n\nmain branch"},
		{name: "empty body", text: "GitHub\n\nBuild failed\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Parse(tt.text)
			if ok != tt.ok {
				t.Fatalf("Parse() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
