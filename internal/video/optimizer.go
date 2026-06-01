package video

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/chewbaccalol/tg-tt-download-bot/internal/config"
)

type Optimizer struct {
	bin string
	cfg config.CompactConfig
}

func NewOptimizer(bin string, cfg config.CompactConfig) *Optimizer {
	return &Optimizer{bin: bin, cfg: cfg}
}

func (o *Optimizer) Compact(ctx context.Context, inputPath, outputPath string) error {
	filter := fmt.Sprintf("scale=-2:min(%d\\,ih)", o.cfg.MaxHeight)
	args := []string{
		"-y",
		"-i", inputPath,
		"-map_metadata", "-1",
		"-vf", filter,
		"-c:v", "libx264",
		"-preset", o.cfg.Preset,
		"-crf", strconv.Itoa(o.cfg.CRF),
		"-c:a", "aac",
		"-b:a", o.cfg.AudioBitrate,
		"-movflags", "+faststart",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, o.bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
