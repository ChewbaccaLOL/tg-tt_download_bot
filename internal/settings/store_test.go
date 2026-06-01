package settings

import (
	"path/filepath"
	"testing"
)

func TestFileStorePersistsQuality(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	store, err := NewFileStore(path, "compact")
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	if got := store.Quality(42); got != "compact" {
		t.Fatalf("default quality = %q, want compact", got)
	}
	if err := store.SetQuality(42, "highest"); err != nil {
		t.Fatalf("SetQuality: %v", err)
	}

	reopened, err := NewFileStore(path, "compact")
	if err != nil {
		t.Fatalf("reopen NewFileStore: %v", err)
	}
	if got := reopened.Quality(42); got != "highest" {
		t.Fatalf("persisted quality = %q, want highest", got)
	}
}

func TestFileStoreToggleQuality(t *testing.T) {
	store, err := NewFileStore(filepath.Join(t.TempDir(), "settings.json"), "compact")
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	if got, err := store.ToggleQuality(42); err != nil || got != "highest" {
		t.Fatalf("first toggle = %q, %v; want highest, nil", got, err)
	}
	if got, err := store.ToggleQuality(42); err != nil || got != "compact" {
		t.Fatalf("second toggle = %q, %v; want compact, nil", got, err)
	}
}
