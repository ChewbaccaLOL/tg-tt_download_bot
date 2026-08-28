package downloader

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDownloadBestInvokesYTDLPAndFindsOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script test")
	}

	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args.txt")
	scriptPath := filepath.Join(dir, "fake-yt-dlp")
	script := `#!/bin/sh
printf '%s\n' "$@" > "` + argsPath + `"
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
  fi
  shift
done
out="${out%\.\%\(ext\)s}.mp4"
printf 'video' > "$out"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}

	path, err := NewYTDLP(scriptPath).DownloadBest(context.Background(), "https://www.tiktok.com/@u/video/1", dir, "source")
	if err != nil {
		t.Fatalf("DownloadBest: %v", err)
	}
	if filepath.Base(path) != "source.mp4" {
		t.Fatalf("downloaded path = %q, want source.mp4", path)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	args := string(argsData)
	for _, want := range []string{"--no-cache-dir", "--no-playlist", "bv*[ext=mp4]+ba[ext=m4a]/best[ext=mp4]/bv*+ba/best", "--merge-output-format", "mp4"} {
		if !strings.Contains(args, want) {
			t.Fatalf("yt-dlp args missing %q in:\n%s", want, args)
		}
	}
}

func TestVersionInvokesYTDLPVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script test")
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "fake-yt-dlp")
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then
  printf '2026.08.19\n'
  exit 0
fi
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}

	version, err := NewYTDLP(scriptPath).Version(context.Background())
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if version != "2026.08.19" {
		t.Fatalf("version = %q, want 2026.08.19", version)
	}
}

func TestProbeMetadataInvokesYTDLPAndParsesDuration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script test")
	}

	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args.txt")
	scriptPath := filepath.Join(dir, "fake-yt-dlp")
	script := `#!/bin/sh
printf '%s\n' "$@" > "` + argsPath + `"
printf '{"duration":125.4,"is_live":false,"live_status":"not_live"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}

	metadata, err := NewYTDLP(scriptPath).ProbeMetadata(context.Background(), "https://youtu.be/abc")
	if err != nil {
		t.Fatalf("ProbeMetadata: %v", err)
	}
	if metadata.DurationSeconds != 125.4 {
		t.Fatalf("DurationSeconds = %v, want 125.4", metadata.DurationSeconds)
	}
	if metadata.IsLive {
		t.Fatal("IsLive = true, want false")
	}
	if metadata.LiveStatus != "not_live" {
		t.Fatalf("LiveStatus = %q, want not_live", metadata.LiveStatus)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	args := string(argsData)
	for _, want := range []string{"--no-cache-dir", "--no-playlist", "--skip-download", "--dump-single-json"} {
		if !strings.Contains(args, want) {
			t.Fatalf("yt-dlp args missing %q in:\n%s", want, args)
		}
	}
}
