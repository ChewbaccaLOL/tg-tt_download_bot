package downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type YTDLP struct {
	bin string
}

func NewYTDLP(bin string) *YTDLP {
	return &YTDLP{bin: bin}
}

func (d *YTDLP) DownloadBest(ctx context.Context, rawURL, dir, id string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	prefix := filepath.Join(dir, id)
	outputTemplate := prefix + ".%(ext)s"
	args := []string{
		"--no-playlist",
		"--restrict-filenames",
		"-f", "bv*+ba/best",
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
