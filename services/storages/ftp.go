package storages

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jlaffaye/ftp"
)

type ftpStorage struct {
	c  *ftp.ServerConn
	id int
}

type ftpStorageConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

func NewFTPStorage(name string, id int, config string) (*ftpStorage, error) {
	var cfg ftpStorageConfig
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return nil, fmt.Errorf("ftp storage: %w", err)
	}

	c, err := ftp.Dial(cfg.Address, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("ftp storage: dial: %w", err)
	}

	err = c.Login(cfg.User, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("ftp storage: login: %w", err)
	}

	return &ftpStorage{
		id: id,
		c:  c,
	}, nil
}

func (f *ftpStorage) ID() int {
	return f.id
}

func (f *ftpStorage) Put(key string, content []byte, expire sql.NullTime) error {
	err := f.c.Stor(key, bytes.NewBuffer(content))
	if err != nil {
		return fmt.Errorf("ftp storage: put: %w", err)
	}
	return nil
}

func (f *ftpStorage) Delete(key string) error {
	err := f.c.Delete(key)
	if err != nil {
		return fmt.Errorf("ftp storage: delete: %w", err)
	}
	return nil
}

func (f *ftpStorage) Get(key string) (any, error) {
	resp, err := f.c.Retr(key)
	if err != nil {
		return nil, fmt.Errorf("ftp storage: get: %w", err)
	}

	defer resp.Close()
	buf, err := io.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("ftp storage: get: %w", err)
	}

	return buf, nil
}
