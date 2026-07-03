package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	QualityHighest = "highest"
	QualityCompact = "compact"

	AccessModePublic          = "public"
	AccessModeWhitelist       = "whitelist"
	AccessModeWhitelistOrPaid = "whitelist_or_paid"
)

type CompactConfig struct {
	MaxHeight    int    `json:"max_height"`
	CRF          int    `json:"crf"`
	AudioBitrate string `json:"audio_bitrate"`
	Preset       string `json:"preset"`
}

type AccessConfig struct {
	Mode                       string  `json:"mode"`
	WhitelistUserIDs           []int64 `json:"whitelist_user_ids"`
	PaidDownloadStars          int     `json:"paid_download_stars"`
	PaidDownloadStarsPerMinute int     `json:"paid_download_stars_per_minute"`
	MaxPaidDurationMinutes     int     `json:"max_paid_duration_minutes"`
}

type Config struct {
	TelegramToken       string        `json:"-"`
	DownloadDir         string        `json:"download_dir"`
	SettingsPath        string        `json:"settings_path"`
	DefaultQuality      string        `json:"default_quality"`
	MaxUploadSizeMB     int64         `json:"max_upload_size_mb"`
	CleanupAfter        time.Duration `json:"-"`
	CleanupAfterMinutes int           `json:"cleanup_after_minutes"`
	YTDLPBin            string        `json:"yt_dlp_bin"`
	FFmpegBin           string        `json:"ffmpeg_bin"`
	AllowedDomains      []string      `json:"allowed_domains"`
	Access              AccessConfig  `json:"access"`
	Compact             CompactConfig `json:"compact"`
}

func Load() (Config, error) {
	cfg := defaultConfig()

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.json"
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	cfg.TelegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.TelegramToken == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}

	applyDefaults(&cfg)
	if err := validateQuality(cfg.DefaultQuality); err != nil {
		return Config{}, err
	}
	if err := validateAccessMode(cfg.Access.Mode); err != nil {
		return Config{}, err
	}

	cfg.CleanupAfter = time.Duration(cfg.CleanupAfterMinutes) * time.Minute
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		DownloadDir:         "./data/downloads",
		SettingsPath:        "./data/settings.json",
		DefaultQuality:      QualityCompact,
		MaxUploadSizeMB:     50,
		CleanupAfterMinutes: 30,
		YTDLPBin:            "yt-dlp",
		FFmpegBin:           "ffmpeg",
		AllowedDomains:      []string{"tiktok.com", "vm.tiktok.com", "vt.tiktok.com", "youtube.com", "youtu.be"},
		Access: AccessConfig{
			Mode:                       AccessModePublic,
			WhitelistUserIDs:           []int64{},
			PaidDownloadStars:          3,
			PaidDownloadStarsPerMinute: 1,
			MaxPaidDurationMinutes:     30,
		},
		Compact: CompactConfig{
			MaxHeight:    720,
			CRF:          28,
			AudioBitrate: "128k",
			Preset:       "veryfast",
		},
	}
}

func applyDefaults(cfg *Config) {
	if cfg.DownloadDir == "" {
		cfg.DownloadDir = "./data/downloads"
	}
	if cfg.SettingsPath == "" {
		cfg.SettingsPath = "./data/settings.json"
	}
	if cfg.DefaultQuality == "" {
		cfg.DefaultQuality = QualityCompact
	}
	if cfg.MaxUploadSizeMB <= 0 {
		cfg.MaxUploadSizeMB = 50
	}
	if cfg.CleanupAfterMinutes <= 0 {
		cfg.CleanupAfterMinutes = 30
	}
	if cfg.YTDLPBin == "" {
		cfg.YTDLPBin = "yt-dlp"
	}
	if cfg.FFmpegBin == "" {
		cfg.FFmpegBin = "ffmpeg"
	}
	if len(cfg.AllowedDomains) == 0 {
		cfg.AllowedDomains = []string{"tiktok.com", "vm.tiktok.com", "vt.tiktok.com", "youtube.com", "youtu.be"}
	}
	if cfg.Access.Mode == "" {
		cfg.Access.Mode = AccessModePublic
	}
	if cfg.Access.PaidDownloadStars <= 0 {
		cfg.Access.PaidDownloadStars = 3
	}
	if cfg.Access.PaidDownloadStarsPerMinute <= 0 {
		cfg.Access.PaidDownloadStarsPerMinute = 1
	}
	if cfg.Access.MaxPaidDurationMinutes <= 0 {
		cfg.Access.MaxPaidDurationMinutes = 30
	}
	if cfg.Compact.MaxHeight <= 0 {
		cfg.Compact.MaxHeight = 720
	}
	if cfg.Compact.CRF <= 0 {
		cfg.Compact.CRF = 28
	}
	if cfg.Compact.AudioBitrate == "" {
		cfg.Compact.AudioBitrate = "128k"
	}
	if cfg.Compact.Preset == "" {
		cfg.Compact.Preset = "veryfast"
	}
}

func validateQuality(quality string) error {
	switch quality {
	case QualityHighest, QualityCompact:
		return nil
	default:
		return fmt.Errorf("unknown quality %q", quality)
	}
}

func validateAccessMode(mode string) error {
	switch mode {
	case AccessModePublic, AccessModeWhitelist, AccessModeWhitelistOrPaid:
		return nil
	default:
		return fmt.Errorf("unknown access mode %q", mode)
	}
}

func IsSupportedURL(raw string, allowedDomains []string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	for _, domain := range allowedDomains {
		domain = strings.TrimPrefix(strings.ToLower(domain), "www.")
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func IsWhitelisted(userID int64, allowed []int64) bool {
	for _, allowedID := range allowed {
		if userID == allowedID {
			return true
		}
	}
	return false
}
