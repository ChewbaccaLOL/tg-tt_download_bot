package video

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chewbaccalol/tg-tt-download-bot/internal/config"
)

func TestCompactInvokesFFmpegWithExpectedOptions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script test")
	}

	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args.txt")
	scriptPath := filepath.Join(dir, "fake-ffmpeg")
	script := `#!/bin/sh
printf '%s\n' "$@" > "` + argsPath + `"
last=""
for arg in "$@"; do
  last="$arg"
done
printf 'compact' > "$last"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}

	inputPath := filepath.Join(dir, "source.mp4")
	outputPath := filepath.Join(dir, "compact.mp4")
	if err := os.WriteFile(inputPath, []byte("source"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	optimizer := NewOptimizer(scriptPath, config.CompactConfig{
		MaxHeight:    480,
		CRF:          29,
		AudioBitrate: "96k",
		Preset:       "fast",
	})
	if err := optimizer.Compact(context.Background(), inputPath, outputPath); err != nil {
		t.Fatalf("Compact: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output was not created: %v", err)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	args := string(argsData)
	for _, want := range []string{"scale=-2:min(480\\,ih)", "libx264", "fast", "29", "aac", "96k", "+faststart"} {
		if !strings.Contains(args, want) {
			t.Fatalf("ffmpeg args missing %q in:\n%s", want, args)
		}
	}
}
