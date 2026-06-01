# Development

Run locally:

```bash
cp .env.example .env
cp config.example.json config.json
go run ./cmd/bot
```

Required local tools:

- Go 1.22+
- `yt-dlp`
- `ffmpeg`

Format and test:

```bash
gofmt -w ./cmd ./internal
go test ./...
```

The project currently uses Go's standard library for Telegram API calls and local JSON files for settings. This keeps the initial setup small and easy to audit.

