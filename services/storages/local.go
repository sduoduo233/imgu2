package storages

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type localStorage struct {
	name string
	path string
}

type localStorageConfig struct {
	Path string `json:"path"`
}

func (s *localStorage) Put(key string, content []byte, expire time.Time) error {
	path := filepath.Join(s.path, key)

	err := os.WriteFile(path, content, 0)
	if err != nil {
		return fmt.Errorf("local storage: %w", err)
	}

	return nil
}

func (s *localStorage) Delete(key string) error {
	path := filepath.Join(s.path, key)

	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("local storage: %w", err)
	}

	return nil
}

func (s *localStorage) Get(key string) (any, error) {
	path := filepath.Join(s.path, key)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("local storage: %w", err)
	}

	return content, nil
}

func NewLocalStorage(name string, config string) (*localStorage, error) {
	var cfg localStorageConfig
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Path == "" {
		return nil, fmt.Errorf("empty path")
	}

	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	err = os.MkdirAll(absPath, fs.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	return &localStorage{
		name: name,
		path: absPath,
	}, nil
}
