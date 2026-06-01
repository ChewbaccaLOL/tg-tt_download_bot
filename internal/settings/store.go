package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type UserSettings struct {
	Quality string `json:"quality"`
}

type FileStore struct {
	path           string
	defaultQuality string
	mu             sync.Mutex
	users          map[int64]UserSettings
}

func NewFileStore(path string, defaultQuality string) (*FileStore, error) {
	store := &FileStore{
		path:           path,
		defaultQuality: defaultQuality,
		users:          make(map[int64]UserSettings),
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	if len(data) == 0 {
		return store, nil
	}
	if err := json.Unmarshal(data, &store.users); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	return store, nil
}

func (s *FileStore) Quality(userID int64) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if settings, ok := s.users[userID]; ok && settings.Quality != "" {
		return settings.Quality
	}
	return s.defaultQuality
}

func (s *FileStore) SetQuality(userID int64, quality string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[userID] = UserSettings{Quality: quality}
	return s.saveLocked()
}

func (s *FileStore) ToggleQuality(userID int64) (string, error) {
	current := s.Quality(userID)
	next := "compact"
	if current == "compact" {
		next = "highest"
	}
	return next, s.SetQuality(userID, next)
}

func (s *FileStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
