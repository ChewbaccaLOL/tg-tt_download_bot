# Telegram TikTok and YouTube Download Bot

A self-hosted Telegram bot for downloading TikTok and YouTube videos.

The bot has one user-facing quality setting:

- `highest`: best available video and best available audio.
- `compact`: optimized MP4 output with smaller file size.

The first supported providers are TikTok and YouTube. The code is structured so Instagram Reels and other providers can be added later through the same downloader flow.

Videos are downloaded to a temporary per-request directory, sent back as Telegram media, then removed locally.

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

Open Telegram, send `/start` to your bot, then paste a TikTok or YouTube URL.

## Settings

Users can send `/settings` to switch between:

- compact optimized downloads
- highest quality downloads

Settings are stored in `data/settings.json`.

## Configuration

The bot reads config from `config.json` by default. You can override the path with `CONFIG_PATH`.

See [config.example.json](config.example.json).

Access can be public, whitelist-only, or whitelist plus Telegram Stars paid downloads. See [docs/access.md](docs/access.md).

Secrets are read from environment variables:

- `TELEGRAM_BOT_TOKEN`

## Updating

For a source-based deployment:

```bash
git pull
docker compose up -d --build
```

When TikTok downloads start failing after previously working, refresh the
bundled `yt-dlp` extractor by rebuilding without Docker's layer cache:

```bash
docker compose build --no-cache bot
docker compose up -d
```

You can confirm the bundled extractor version with:

```bash
docker compose exec bot yt-dlp --version
```

TikTok may also require browser impersonation support. Confirm that the image
has available impersonation targets with:

```bash
docker compose exec bot yt-dlp --list-impersonate-targets
```

To inspect recent failures, check the bot logs:

```bash
docker compose logs --tail=200 bot
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
