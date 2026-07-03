# Setup

## Create a Telegram Bot

1. Open Telegram and message `@BotFather`.
2. Run `/newbot`.
3. Pick a display name and username.
4. Copy the token into `.env`.

```env
TELEGRAM_BOT_TOKEN=123456:your-token-here
```

## Prepare Config

```bash
cp .env.example .env
cp config.example.json config.json
```

The default config supports TikTok and YouTube links and stores runtime data under `data/`.

## Start

```bash
docker compose up -d --build
```

Then send `/start` to your bot.
