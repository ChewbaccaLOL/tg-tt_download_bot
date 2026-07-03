# Access Control and Telegram Stars

The bot can run in three access modes:

- `public`: anyone who can message the bot can download.
- `whitelist`: only configured Telegram user IDs can download.
- `whitelist_or_paid`: whitelisted users download for free; everyone else receives a Telegram Stars invoice based on video length.

Configure this in `config.json`:

```json
{
  "access": {
    "mode": "whitelist_or_paid",
    "whitelist_user_ids": [123456789],
    "paid_download_stars_per_minute": 1,
    "max_paid_duration_minutes": 30
  }
}
```

If a non-whitelisted user messages a `whitelist` bot, the bot replies with their Telegram user ID so you can add them to the config.

## Stars Pricing

Telegram Stars payments use the `XTR` currency. Telegram requires digital goods and services sold inside bots to use Stars.

There is no stable fixed USD value for one Star. The value depends on how users buy Stars and how developers withdraw rewards. Treat `paid_download_stars_per_minute` as a throttle and convenience fee, not a precise metered cloud-cost number.

For a small VPS, the marginal cost of one short-video request is usually dominated by bandwidth and CPU time:

- 20-80 MB transferred per request is common for short videos.
- At $0.01/GB overage bandwidth, 80 MB costs roughly $0.0008.
- On a $5/month VPS, 5,000 downloads/month adds roughly $0.001 of fixed server cost per download.

That makes 1 Star per minute likely enough for raw marginal infra in many cases, while still scaling naturally for longer YouTube links.

Before sending an invoice, the bot asks `yt-dlp` for clean video metadata. Paid downloads are rejected before payment when the video has no clear duration, looks live or upcoming, or exceeds `max_paid_duration_minutes`.

## Payment Flow

For paid users:

1. User sends a supported URL.
2. Bot checks the video duration.
3. Bot sends a Telegram Stars invoice.
4. Bot answers Telegram's pre-checkout query.
5. Bot waits for `successful_payment`.
6. Bot downloads and sends the video.

Pending payment requests are kept in memory for 15 minutes. If the bot restarts before payment completes, the user should send the link again.
