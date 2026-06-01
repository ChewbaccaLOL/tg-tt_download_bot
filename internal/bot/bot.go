package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chewbaccalol/tg-tt-download-bot/internal/config"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/downloader"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/settings"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/telegram"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/video"
)

type Dependencies struct {
	Config     config.Config
	Telegram   *telegram.Client
	Downloader *downloader.YTDLP
	Optimizer  *video.Optimizer
	Settings   *settings.FileStore
}

type Bot struct {
	cfg        config.Config
	tg         *telegram.Client
	downloader *downloader.YTDLP
	optimizer  *video.Optimizer
	settings   *settings.FileStore
}

func New(deps Dependencies) *Bot {
	return &Bot{
		cfg:        deps.Config,
		tg:         deps.Telegram,
		downloader: deps.Downloader,
		optimizer:  deps.Optimizer,
		settings:   deps.Settings,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		updates, err := b.tg.GetUpdates(ctx, offset, 25)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("get updates: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update telegram.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}
	if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
		return
	}

	msg := update.Message
	text := strings.TrimSpace(msg.Text)

	switch {
	case strings.HasPrefix(text, "/start"), strings.HasPrefix(text, "/help"):
		_ = b.tg.SendMessage(ctx, msg.Chat.ID, "Send me a TikTok URL and I will download it. Use /settings to switch between highest quality and compact mode.")
	case strings.HasPrefix(text, "/settings"):
		b.sendSettings(ctx, msg.Chat.ID, userID(msg))
	default:
		if config.IsSupportedURL(text, b.cfg.AllowedDomains) {
			go b.downloadAndSend(ctx, msg.Chat.ID, userID(msg), text)
			return
		}
		_ = b.tg.SendMessage(ctx, msg.Chat.ID, "I can download supported TikTok links for now. Send /settings to change quality mode.")
	}
}

func (b *Bot) handleCallback(ctx context.Context, query *telegram.CallbackQuery) {
	if query.Data != "toggle_quality" {
		_ = b.tg.AnswerCallbackQuery(ctx, query.ID, "")
		return
	}

	quality, err := b.settings.ToggleQuality(query.From.ID)
	if err != nil {
		_ = b.tg.AnswerCallbackQuery(ctx, query.ID, "Could not update settings")
		return
	}

	_ = b.tg.AnswerCallbackQuery(ctx, query.ID, "Quality set to "+quality)
	if query.Message != nil {
		_ = b.tg.EditMessageText(ctx, query.Message.Chat.ID, query.Message.MessageID, settingsText(quality), settingsKeyboard(quality))
	}
}

func (b *Bot) sendSettings(ctx context.Context, chatID int64, userID int64) {
	quality := b.settings.Quality(userID)
	_ = b.tg.SendMessageWithKeyboard(ctx, chatID, settingsText(quality), settingsKeyboard(quality))
}

func (b *Bot) downloadAndSend(ctx context.Context, chatID int64, userID int64, rawURL string) {
	quality := b.settings.Quality(userID)
	_ = b.tg.SendMessage(ctx, chatID, "Downloading in "+quality+" mode...")

	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	workDir := filepath.Join(b.cfg.DownloadDir, requestID)
	defer os.RemoveAll(workDir)

	sourcePath, err := b.downloader.DownloadBest(ctx, rawURL, workDir, "source")
	if err != nil {
		log.Printf("download failed: %v", err)
		_ = b.tg.SendMessage(ctx, chatID, "Download failed. The link may be unavailable or unsupported.")
		return
	}

	outputPath := sourcePath
	if quality == config.QualityCompact {
		outputPath = filepath.Join(workDir, "compact.mp4")
		if err := b.optimizer.Compact(ctx, sourcePath, outputPath); err != nil {
			log.Printf("optimize failed: %v", err)
			_ = b.tg.SendMessage(ctx, chatID, "Optimization failed after download.")
			return
		}
	}

	if err := b.ensureUploadable(outputPath); err != nil {
		log.Printf("upload check failed: %v", err)
		_ = b.tg.SendMessage(ctx, chatID, "The video is larger than this bot is configured to upload.")
		return
	}

	if err := b.tg.SendDocument(ctx, chatID, outputPath, ""); err != nil {
		log.Printf("send document failed: %v", err)
		_ = b.tg.SendMessage(ctx, chatID, "Telegram upload failed.")
	}
}

func (b *Bot) ensureUploadable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	limit := b.cfg.MaxUploadSizeMB * 1024 * 1024
	if info.Size() > limit {
		return fmt.Errorf("file size %d exceeds limit %d", info.Size(), limit)
	}
	return nil
}

func settingsText(quality string) string {
	if quality == config.QualityHighest {
		return "Quality mode: highest\nBest available video and best available audio."
	}
	return "Quality mode: compact\nSmaller optimized MP4 output."
}

func settingsKeyboard(quality string) telegram.InlineKeyboardMarkup {
	next := "Switch to highest"
	if quality == config.QualityHighest {
		next = "Switch to compact"
	}
	return telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{{Text: next, CallbackData: "toggle_quality"}},
		},
	}
}

func userID(msg *telegram.Message) int64 {
	if msg.From != nil {
		return msg.From.ID
	}
	return msg.Chat.ID
}
