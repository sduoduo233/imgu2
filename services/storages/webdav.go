package storages

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"github.com/emersion/go-webdav"
)

type webdavStorage struct {
	client *webdav.Client
	id     int
}

type webdavStorageConfig struct {
	Endpoint string `json:"endpoint"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func NewWebDAVStorage(name string, id int, config string) (*webdavStorage, error) {
	var cfg webdavStorageConfig
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return nil, fmt.Errorf("WebDAV storage: %w", err)
	}

	client, err := webdav.NewClient(webdav.HTTPClientWithBasicAuth(nil, cfg.User, cfg.Password), cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("WebDAV storage: %w", err)
	}

	return &webdavStorage{client, id}, nil
}

func (w *webdavStorage) ID() int {
	return w.id
}

func (w *webdavStorage) Put(key string, content []byte, expire sql.NullTime) error {
	writer, err := w.client.Create(context.Background(), key)
	if err != nil {
		return fmt.Errorf("WebDAV storage: %w", err)
	}
	defer writer.Close()

	_, err = writer.Write(content)
	if err != nil {
		return fmt.Errorf("WebDAV storage: %w", err)
	}

	return nil
}

func (w *webdavStorage) Delete(key string) error {
	err := w.client.RemoveAll(context.Background(), key)
	if err != nil {
		return fmt.Errorf("WebDAV storage: %w", err)
	}

	return nil
}

func (w *webdavStorage) Get(key string) (any, error) {
	reader, err := w.client.Open(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("WebDAV storage: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("WebDAV storage: %w", err)
	}

	return content, nil
}
