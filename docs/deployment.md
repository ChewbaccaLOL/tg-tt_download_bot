# Deployment

## Docker Compose

The recommended deployment path is Docker Compose.

```bash
cp .env.example .env
cp config.example.json config.json
```

Set `TELEGRAM_BOT_TOKEN` in `.env`, then run:

```bash
docker compose up -d --build
```

Logs:

```bash
docker compose logs -f bot
```

Stop:

```bash
docker compose down
```

## Updating

When running from source:

```bash
git pull
docker compose up -d --build
```

The local files below are intentionally ignored by git:

- `.env`
- `config.json`
- `data/`

