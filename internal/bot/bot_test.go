package bot

import (
	"strings"
	"testing"

	"github.com/chewbaccalol/tg-tt-download-bot/internal/config"
	"github.com/chewbaccalol/tg-tt-download-bot/internal/downloader"
)

func TestSettingsText(t *testing.T) {
	tests := []struct {
		quality string
		want    string
	}{
		{quality: config.QualityHighest, want: "Quality mode: highest\nBest available video and best available audio."},
		{quality: config.QualityCompact, want: "Quality mode: compact\nSmaller optimized MP4 output."},
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			if got := settingsText(tt.quality); got != tt.want {
				t.Fatalf("settingsText(%q) = %q, want %q", tt.quality, got, tt.want)
			}
		})
	}
}

func TestSettingsKeyboardShowsOppositeAction(t *testing.T) {
	tests := []struct {
		quality string
		want    string
	}{
		{quality: config.QualityHighest, want: "Switch to compact"},
		{quality: config.QualityCompact, want: "Switch to highest"},
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			keyboard := settingsKeyboard(tt.quality)
			if len(keyboard.InlineKeyboard) != 1 || len(keyboard.InlineKeyboard[0]) != 1 {
				t.Fatalf("keyboard shape = %#v, want one button", keyboard.InlineKeyboard)
			}

			button := keyboard.InlineKeyboard[0][0]
			if button.Text != tt.want {
				t.Fatalf("button text = %q, want %q", button.Text, tt.want)
			}
			if button.CallbackData != "toggle_quality" {
				t.Fatalf("callback data = %q, want toggle_quality", button.CallbackData)
			}
		})
	}
}

func TestAccessFor(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		user int64
		want accessDecision
	}{
		{
			name: "public allows everyone",
			cfg:  config.Config{Access: config.AccessConfig{Mode: config.AccessModePublic}},
			user: 42,
			want: accessAllowed,
		},
		{
			name: "whitelist allows listed user",
			cfg:  config.Config{Access: config.AccessConfig{Mode: config.AccessModeWhitelist, WhitelistUserIDs: []int64{42}}},
			user: 42,
			want: accessAllowed,
		},
		{
			name: "whitelist denies unlisted user",
			cfg:  config.Config{Access: config.AccessConfig{Mode: config.AccessModeWhitelist, WhitelistUserIDs: []int64{42}}},
			user: 7,
			want: accessDenied,
		},
		{
			name: "whitelist or paid charges unlisted user",
			cfg:  config.Config{Access: config.AccessConfig{Mode: config.AccessModeWhitelistOrPaid, WhitelistUserIDs: []int64{42}}},
			user: 7,
			want: accessPaid,
		},
		{
			name: "whitelist or paid allows listed user",
			cfg:  config.Config{Access: config.AccessConfig{Mode: config.AccessModeWhitelistOrPaid, WhitelistUserIDs: []int64{42}}},
			user: 42,
			want: accessAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := New(Dependencies{Config: tt.cfg})
			if got := bot.accessFor(tt.user); got != tt.want {
				t.Fatalf("accessFor(%d) = %v, want %v", tt.user, got, tt.want)
			}
		})
	}
}

func TestCalculatePaidDownloadPrice(t *testing.T) {
	access := config.AccessConfig{
		PaidDownloadStarsPerMinute: 2,
		MaxPaidDurationMinutes:     10,
	}

	tests := []struct {
		name     string
		metadata downloader.Metadata
		want     paidDownloadPrice
	}{
		{
			name:     "rounds partial minute up",
			metadata: downloader.Metadata{DurationSeconds: 61},
			want:     paidDownloadPrice{Stars: 4, BillableMinutes: 2},
		},
		{
			name:     "minimum one billable minute",
			metadata: downloader.Metadata{DurationSeconds: 12},
			want:     paidDownloadPrice{Stars: 2, BillableMinutes: 1},
		},
		{
			name:     "exact minute",
			metadata: downloader.Metadata{DurationSeconds: 600},
			want:     paidDownloadPrice{Stars: 20, BillableMinutes: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculatePaidDownloadPrice(access, tt.metadata)
			if err != nil {
				t.Fatalf("calculatePaidDownloadPrice: %v", err)
			}
			if got != tt.want {
				t.Fatalf("price = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCalculatePaidDownloadPriceRejectsUncleanLengths(t *testing.T) {
	access := config.AccessConfig{
		PaidDownloadStarsPerMinute: 1,
		MaxPaidDurationMinutes:     10,
	}

	tests := []struct {
		name     string
		metadata downloader.Metadata
		want     string
	}{
		{
			name:     "missing duration",
			metadata: downloader.Metadata{},
			want:     "clean video length",
		},
		{
			name:     "is live",
			metadata: downloader.Metadata{DurationSeconds: 60, IsLive: true},
			want:     "Live or upcoming",
		},
		{
			name:     "live status",
			metadata: downloader.Metadata{DurationSeconds: 60, LiveStatus: "is_upcoming"},
			want:     "Live or upcoming",
		},
		{
			name:     "over max duration",
			metadata: downloader.Metadata{DurationSeconds: 601},
			want:     "over the 10 minute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calculatePaidDownloadPrice(access, tt.metadata)
			if err == nil {
				t.Fatal("calculatePaidDownloadPrice returned nil error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}
