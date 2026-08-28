package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Metadata struct {
	DurationSeconds float64
	IsLive          bool
	LiveStatus      string
}

type YTDLP struct {
	bin string
}

func NewYTDLP(bin string) *YTDLP {
	return &YTDLP{bin: bin}
}

func (d *YTDLP) Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, d.bin, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp version failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (d *YTDLP) ProbeMetadata(ctx context.Context, rawURL string) (Metadata, error) {
	args := []string{
		"--no-cache-dir",
		"--no-playlist",
		"--skip-download",
		"--dump-single-json",
		rawURL,
	}

	cmd := exec.CommandContext(ctx, d.bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return Metadata{}, fmt.Errorf("yt-dlp metadata probe failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	var result struct {
		Duration   float64 `json:"duration"`
		IsLive     bool    `json:"is_live"`
		LiveStatus string  `json:"live_status"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return Metadata{}, fmt.Errorf("parse yt-dlp metadata: %w", err)
	}

	return Metadata{
		DurationSeconds: result.Duration,
		IsLive:          result.IsLive,
		LiveStatus:      result.LiveStatus,
	}, nil
}

func (d *YTDLP) DownloadBest(ctx context.Context, rawURL, dir, id string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	prefix := filepath.Join(dir, id)
	outputTemplate := prefix + ".%(ext)s"
	args := []string{
		"--no-cache-dir",
		"--no-playlist",
		"--restrict-filenames",
		"-f", "bv*[ext=mp4]+ba[ext=m4a]/best[ext=mp4]/bv*+ba/best",
		"--merge-output-format", "mp4",
		"-o", outputTemplate,
		rawURL,
	}

	cmd := exec.CommandContext(ctx, d.bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	path, err := findDownloadedFile(dir, id)
	if err != nil {
		return "", fmt.Errorf("find downloaded file: %w", err)
	}
	return path, nil
}

func findDownloadedFile(dir, id string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	prefix := id + "."
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(dir, entry.Name()), nil
		}
	}
	return "", os.ErrNotExist
}
