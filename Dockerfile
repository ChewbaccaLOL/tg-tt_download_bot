FROM golang:1.22-bookworm AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /out/tg-video-bot ./cmd/bot

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg python3 python3-pip \
    && pip3 install --break-system-packages --no-cache-dir yt-dlp \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/tg-video-bot /usr/local/bin/tg-video-bot
COPY config.example.json /app/config.example.json

ENV CONFIG_PATH=/app/config.json

VOLUME ["/app/data"]
CMD ["tg-video-bot"]

