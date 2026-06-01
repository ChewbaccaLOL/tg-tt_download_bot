package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chewbaccalol/tg-tt-download-bot/internal/bot"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/config"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/downloader"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/settings"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/telegram"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/video"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := os.MkdirAll(cfg.DownloadDir, 0o755); err != nil {
		log.Fatalf("create download dir: %v", err)
	}

	store, err := settings.NewFileStore(cfg.SettingsPath, cfg.DefaultQuality)
	if err != nil {
		log.Fatalf("open settings store: %v", err)
	}

	tg := telegram.NewClient(cfg.TelegramToken)
	dl := downloader.NewYTDLP(cfg.YTDLPBin)
	optimizer := video.NewOptimizer(cfg.FFmpegBin, cfg.Compact)

	app := bot.New(bot.Dependencies{
		Config:     cfg,
		Telegram:   tg,
		Downloader: dl,
		Optimizer:  optimizer,
		Settings:   store,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("bot started")
	if err := app.Run(ctx); err != nil {
		log.Fatalf("bot stopped: %v", err)
	}
}
