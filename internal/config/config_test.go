package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSupportedURL(t *testing.T) {
	allowed := []string{"tiktok.com", "vm.tiktok.com", "vt.tiktok.com"}

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "tiktok canonical", raw: "https://www.tiktok.com/@user/video/123", want: true},
		{name: "tiktok short", raw: "https://vm.tiktok.com/ZMabc/", want: true},
		{name: "nested supported host", raw: "https://m.tiktok.com/v/123", want: true},
		{name: "unsupported host", raw: "https://example.com/video/123", want: false},
		{name: "not a url", raw: "hello", want: false},
		{name: "unsupported scheme", raw: "ftp://tiktok.com/video/123", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedURL(tt.raw, allowed); got != tt.want {
				t.Fatalf("IsSupportedURL(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestLoadConfigFromFileAndEnv(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	data := `{
  "download_dir": "` + filepath.ToSlash(filepath.Join(dir, "downloads")) + `",
  "settings_path": "` + filepath.ToSlash(filepath.Join(dir, "settings.json")) + `",
  "default_quality": "highest",
  "max_upload_size_mb": 42,
  "cleanup_after_minutes": 7,
  "yt_dlp_bin": "/usr/local/bin/yt-dlp",
  "ffmpeg_bin": "/usr/local/bin/ffmpeg",
  "allowed_domains": ["example.com"],
  "access": {
    "mode": "whitelist_or_paid",
    "whitelist_user_ids": [42, 99],
    "paid_download_stars": 5
  },
  "compact": {
    "max_height": 480,
    "crf": 30,
    "audio_bitrate": "96k",
    "preset": "fast"
  }
}`
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("CONFIG_PATH", configPath)
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.TelegramToken != "token" {
		t.Fatalf("TelegramToken = %q, want token", cfg.TelegramToken)
	}
	if cfg.DefaultQuality != QualityHighest {
		t.Fatalf("DefaultQuality = %q, want highest", cfg.DefaultQuality)
	}
	if cfg.MaxUploadSizeMB != 42 {
		t.Fatalf("MaxUploadSizeMB = %d, want 42", cfg.MaxUploadSizeMB)
	}
	if cfg.CleanupAfterMinutes != 7 || cfg.CleanupAfter.String() != "7m0s" {
		t.Fatalf("cleanup = %d/%s, want 7/7m0s", cfg.CleanupAfterMinutes, cfg.CleanupAfter)
	}
	if got := cfg.Compact.AudioBitrate; got != "96k" {
		t.Fatalf("Compact.AudioBitrate = %q, want 96k", got)
	}
	if cfg.Access.Mode != AccessModeWhitelistOrPaid {
		t.Fatalf("Access.Mode = %q, want whitelist_or_paid", cfg.Access.Mode)
	}
	if !IsWhitelisted(99, cfg.Access.WhitelistUserIDs) {
		t.Fatal("expected user 99 to be whitelisted")
	}
	if cfg.Access.PaidDownloadStars != 5 {
		t.Fatalf("PaidDownloadStars = %d, want 5", cfg.Access.PaidDownloadStars)
	}
}

func TestLoadRequiresTelegramToken(t *testing.T) {
	t.Setenv("CONFIG_PATH", filepath.Join(t.TempDir(), "missing.json"))
	t.Setenv("TELEGRAM_BOT_TOKEN", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load returned nil error without TELEGRAM_BOT_TOKEN")
	}
}

func TestLoadRejectsUnknownDefaultQuality(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"default_quality":"tiny"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("CONFIG_PATH", configPath)
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")

	if _, err := Load(); err == nil {
		t.Fatal("Load returned nil error for unknown quality")
	}
}

func TestLoadRejectsUnknownAccessMode(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"access":{"mode":"friends_only"}}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("CONFIG_PATH", configPath)
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")

	if _, err := Load(); err == nil {
		t.Fatal("Load returned nil error for unknown access mode")
	}
}
