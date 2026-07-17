# Notier OpenCode Notes

## Scope
- Small single-module Go daemon for macOS Telegram-to-notification forwarding.
- Go version is `1.26.4` from `go.mod`.
- There is no README, CI, Makefile, Taskfile, linter config, or OpenCode config, so derive commands from Go tooling and current tests.

## Commands
- Full tests: `go test ./...`
- Focus notify package: `go test ./internal/notify`
- Focus paths package: `go test ./internal/paths`
- Focus message parser cases: `go test ./internal/notify -run TestParse`
- Focus sender app validation: `go test ./internal/notify -run TestLoadApps`
- Focus sender app path: `go test ./internal/paths -run TestDefault_sets_apps_dir_under_user_applications`
- Race and order check: `go test -race -shuffle=on -count=1 ./...`
- Vet: `go vet ./...`
- Build without polluting repo: `go build -o /tmp/notier ./cmd/notier`

## Runtime Flow
- Entry point: `cmd/notier/main.go`.
- Actual path: `paths.Default()` -> `config.Load()` or `config.WriteTemplate()` -> `notify.New()` -> `client.Run()` -> `notify.Parse()` -> `Terminal.Notify()`.
- First run writes the config template with mode `0600` and exits with `created config file ...; fill it and run again`.
- Gotd login can prompt on stdin for `telegram code:`.
- Telegram session persists at `xdg.SessionFile`.
- `target_peer_id = 0` disables peer filtering; a nonzero value accepts only that Telegram peer.

## Paths
- Config: `<XDG_CONFIG_HOME>/notier/config.toml`, or `~/.config/notier/config.toml` when unset or relative.
- Sender apps: `~/Applications/Notier Senders`.
- Session: `~/.local/share/notier/session.json` by default.
- `paths.Default` exposes state log paths, but `cmd/notier/main.go` does not redirect stdout or stderr itself.
- Current quirk: session path uses `XDG_STATE_HOME` when it is set because `xdg.go` passes that env name for `dataDir`; don't silently fix it in unrelated work.

## Message Contract
- Telegram text format is exactly `APP\n\nTITLE\n\nMSG`.
- `notify.Parse` uses `strings.SplitN(text, "\n\n", 3)`, so the body may contain further blank lines.
- Malformed messages, missing sections, or blank trimmed sections are ignored.
- Known lowercase app keys are `qq`, `wechat`, `wecom`, `mail`, `sms`.
- Unknown app names are allowed; they use plain `terminal-notifier` because they have no native sender app key.
- `Notifacation` is the current exported type spelling; don't opportunistically rename it.

## Notifications And Sender Apps
- The configured notifier value remains `terminal-notifier`; `notify.New` also loads optional native sender apps. The binary is used only as fallback for unknown app names or known apps whose native bundle is missing.
- Runtime requires macOS. Present native bundles use `/usr/bin/plutil` and LaunchServices `lsregister` during startup; fallback delivery requires `terminal-notifier` on `PATH`.
- Loaded app keys run the bundle's `Contents/MacOS/NotierSender`: `--title` receives `msg.Subtitle`, `--subtitle` receives an empty string, and `--message` receives `msg.Body`; `msg.Title` is discarded on this native path.
- Recognized bundles are `NotierQQ.app` (`com.notier.sender.qq`), `NotierWeChat.app` (`com.notier.sender.wechat`), `NotierWeCom.app` (`com.notier.sender.wecom`), `NotierMail.app` (`com.notier.sender.mail`), and `NotierSMS.app` (`com.notier.sender.sms`).
- Clicking QQ, WeChat, or Mail notifications launches the corresponding application; WeCom and SMS clicks intentionally do nothing. These native sender implementations live outside this repository.
- QQ, WeChat, and Mail notifications are suppressed while their corresponding application process is running; WeCom and SMS have no process check.
- Startup validates present bundle IDs with `/usr/bin/plutil` and requires an executable `Contents/MacOS/NotierSender` before any registration.
- Missing bundles are skipped without startup validation or registration commands and fall back to plain `terminal-notifier` unless the running-app suppression above applies.
- Bundle ID mismatch aborts startup before registration.
- Registration failure from `lsregister -f <app>` aborts startup with the app path in the error.
