package storages

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
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

	// create form data
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("file", "blob")
	if err != nil {
		return "", fmt.Errorf("telegraph storage: put: %w", err)
	}

	_, err = part.Write(content)
	if err != nil {
		return "", fmt.Errorf("telegraph storage: put: %w", err)
	}

	writer.Close()

	// create request
	req, err := http.NewRequest(http.MethodPost, "https://telegra.ph/upload", buf)
	if err != nil {
		return "", fmt.Errorf("telegraph storage: put: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("telegraph storage: put: %w", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("telegraph storage: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("telegraph storage: bad status code: %d %s", resp.StatusCode, respBody)
	}

	// parse request
	type respStruct struct {
		Src string `json:"src"`
	}
	var respJson []respStruct
	err = json.Unmarshal(respBody, &respJson)
	if err != nil {
		return "", fmt.Errorf("telegraph storage: malformatted json: %w, %s", err, respBody)
	}

	if len(respJson) == 0 {
		return "", fmt.Errorf("telegraph storage: empty list")
	}

	return filepath.Base(respJson[0].Src), nil
}

func (s *telegraphStorage) Delete(key string) error {
	// telegraph does not support delete images
	return nil
}

func (s *telegraphStorage) Get(key string) (any, error) {
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
