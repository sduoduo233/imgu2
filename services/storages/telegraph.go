package storages

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type telegraphStorage struct {
	name      string
	publicURL string
	id        int
}

type telegraphStorageConfig struct {
	PublicURL string `json:"public_url"`
}

func (s *telegraphStorage) Put(key string, content []byte, expire sql.NullTime) (string, error) {
	return "", fmt.Errorf("telegraph upload is not available")
}

func (s *telegraphStorage) Delete(key string) error {
	// telegraph does not support delete images
	return nil
}

func (s *telegraphStorage) Get(key string) (any, error) {

	if s.publicURL != "" {
		return s.publicURL + "/" + key, nil
	}

	resp, err := http.Get("https://telegra.ph/file/" + key)
	if err != nil {
		return nil, fmt.Errorf("telegraph storage: get: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegraph storage: read: %w", err)
	}

	return content, nil
}

func (s *telegraphStorage) ID() int {
	return s.id
}

func NewTelegraphStorage(name string, id int, config string) (*telegraphStorage, error) {
	var cfg telegraphStorageConfig
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &telegraphStorage{
		name:      name,
		id:        id,
		publicURL: cfg.PublicURL,
	}, nil
}
