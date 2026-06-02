package bot

import (
	"testing"

	"github.com/chewbaccalol/tg-tt-download-bot/internal/config"
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
