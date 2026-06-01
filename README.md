# Telegram TikTok Download Bot

A self-hosted Telegram bot for downloading TikTok videos.

The bot has one user-facing quality setting:

- `highest`: best available video and best available audio.
- `compact`: optimized MP4 output with smaller file size.

The first supported provider is TikTok. The code is structured so YouTube Shorts, Instagram Reels, and other providers can be added later through the same downloader flow.

## Requirements

For Docker deployment:

- Docker
- Docker Compose
- A Telegram bot token from BotFather

For local development:

- Go 1.22+
- `yt-dlp`
- `ffmpeg`

## Quick Start

Create a Telegram bot with BotFather first. See [docs/setup.md](docs/setup.md) for the full setup flow.

```bash
cp .env.example .env
cp config.example.json config.json
```

Edit `.env` and set:

```env
TELEGRAM_BOT_TOKEN=123456:your-token-here
```

Start the bot:

```bash
docker compose up -d --build
```

Open Telegram, send `/start` to your bot, then paste a TikTok URL.

## Settings

Users can send `/settings` to switch between:

- compact optimized downloads
- highest quality downloads

Settings are stored in `data/settings.json`.

## Configuration

The bot reads config from `config.json` by default. You can override the path with `CONFIG_PATH`.

See [config.example.json](config.example.json).

Secrets are read from environment variables:

- `TELEGRAM_BOT_TOKEN`

## Updating

For a source-based deployment:

```bash
git pull
docker compose up -d --build
```

Keep your local `.env`, `config.json`, and `data/` directory outside git.

## CI/CD

The included GitHub Actions workflow runs Go tests on `main` and pull requests. Tags matching `v*` publish a Docker image to GitHub Container Registry.

See [docs/github-runner.md](docs/github-runner.md) for notes on self-hosted runner deployment.

## Legal Note

This project is a self-hosted tool. You are responsible for using it only with content you have rights to access and download, and for complying with platform terms and local law.

## Maintenance Stance

This source is published for people who want to run their own instance. It is not currently intended to be a collaborative project, so issues and pull requests may not be reviewed.

## License

MIT. See [LICENSE](LICENSE).
