package storages

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Storage struct {
	s3Client  *s3.S3
	bucket    string
	publicURL string
	id        int
}

type s3StorageConfig struct {
	KeyID     string `json:"key_id"`
	Secret    string `json:"secret"`
	Token     string `json:"token"`
	Endpoint  string `json:"endpoint"`
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	PublicURL string `json:"public_url"`
}

func NewS3Storage(name string, id int, config string) (*s3Storage, error) {
	var cfg s3StorageConfig
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return nil, fmt.Errorf("s3 storage: %w", err)
	}

	s3session := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.KeyID, cfg.Secret, cfg.Token),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(true),
	}))

	return &s3Storage{
		s3Client:  s3.New(s3session),
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
		id:        id,
	}, nil
}

func (s *s3Storage) Delete(key string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("s3 storage: %w", err)
	}
	return nil
}

func (s *s3Storage) Put(key string, content []byte, expire sql.NullTime) error {
	var e *time.Time
	if expire.Valid {
		e = &expire.Time
	}

	_, err := s.s3Client.PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(content),
		Bucket:      &s.bucket,
		ContentType: aws.String(http.DetectContentType(content)),
		Expires:     e,
		Key:         &key,
	})
	if err != nil {
		return fmt.Errorf("s3 storage: %w", err)
	}
	return nil
}

func (s *s3Storage) Get(key string) (any, error) {
	return s.publicURL + "/" + key, nil
}

func (s *s3Storage) ID() int {
	return s.id
}
