# Telegram Video Download Bot Plan

## Goal

Build a self-hosted Telegram bot that downloads short-form videos for users. The initial provider is TikTok, with room to add YouTube Shorts, Instagram Reels, and similar providers later.

The project is open source and intended for people who want to run their own instance with their own Telegram bot token and infrastructure.

## Product Scope

- Accept a supported video URL in Telegram.
- Download the video and send it back to the user.
- Keep one per-user quality setting:
  - `highest`: best available video and best available audio.
  - `compact`: optimized output with smaller file size.
- Start with TikTok-only support.
- Keep provider-specific logic isolated so more sites can be added later.
- Prefer simple self-hosting over complex platform assumptions.

## Technical Direction

- Language: Go.
- Telegram API: direct HTTP calls from Go's standard library.
- Download engine: `yt-dlp` external binary.
- Video optimization: `ffmpeg` external binary.
- Config: local JSON config file plus environment variables for secrets.
- User settings: local JSON state file for the first version.
- Deployment: Docker and Docker Compose.
- License: MIT.

## Configuration Model

Secrets should stay out of source control:

- `TELEGRAM_BOT_TOKEN` is read from the environment.

Local non-secret config is read from `config.json` by default, or from `CONFIG_PATH` if set.

Examples:

- `config.example.json`
- `.env.example`

## Deployment Model

Default deployment:

1. Copy `.env.example` to `.env`.
2. Copy `config.example.json` to `config.json`.
3. Put a Telegram bot token in `.env`.
4. Run `docker compose up -d --build`.

Update path:

1. Pull new code.
2. Rebuild/restart Docker Compose.

Later release path:

- Build and publish Docker images from GitHub Actions on tags.
- Users can pin a tagged image instead of building locally.
- Optional docs for a self-hosted GitHub runner that deploys on pushes/tags.

## Provider Roadmap

### Phase 1

- TikTok URLs:
  - `tiktok.com`
  - `www.tiktok.com`
  - `vm.tiktok.com`
  - `vt.tiktok.com`

### Later

- YouTube Shorts.
- Instagram Reels.
- Provider-specific validation and error messages.

## UX Roadmap

### Phase 1

- `/start` and `/help`.
- `/settings` with one inline toggle.
- Paste URL to download.

### Later

- Better progress messages.
- Admin-only configuration reload.
- Optional allowlist mode.
- Optional default quality override.

## Legal/Policy Note

The project should clearly state that users are responsible for complying with copyright law, platform terms, and local rules. The bot is a self-hosted tool and should only be used for content the user has rights to access and download.

