package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type SystemObjectStorage struct {
	OssType int64 `json:"oss_type"`
	Minio   struct {
		Endpoint        string `json:"endpoint"`
		AccessKeyID     string `json:"access_key_id"`
		AccessKeySecret string `json:"access_key_secret"`
		BucketName      string `json:"bucket_name"`
		BucketURL       string `json:"bucket_url"`
	} `json:"minio"`
}

type SystemObjectStorageModel interface {
	FindSystemObjectStorage(ctx context.Context) (SystemObjectStorage, error)
}

type defaultSystemObjectStorageModel struct {
	db *sql.DB
}

func NewSystemObjectStorageModel(db *sql.DB) SystemObjectStorageModel {
	return &defaultSystemObjectStorageModel{db: db}
}

func (m *defaultSystemObjectStorageModel) FindSystemObjectStorage(
	ctx context.Context,
) (SystemObjectStorage, error) {
	var raw []byte
	err := m.db.QueryRowContext(
		ctx,
		`SELECT config_value
FROM sys_config
WHERE tenant_id=? AND config_key=?
LIMIT 1`,
		int64(0),
		"OBJECT_STORAGE",
	).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return SystemObjectStorage{}, errors.New("system OBJECT_STORAGE config does not exist")
	}
	if err != nil {
		return SystemObjectStorage{}, fmt.Errorf("query system OBJECT_STORAGE config: %w", err)
	}

	var config SystemObjectStorage
	if err := json.Unmarshal(raw, &config); err != nil {
		return SystemObjectStorage{}, fmt.Errorf("decode system OBJECT_STORAGE config: %w", err)
	}
	config.Minio.Endpoint = strings.TrimSpace(config.Minio.Endpoint)
	config.Minio.AccessKeyID = strings.TrimSpace(config.Minio.AccessKeyID)
	config.Minio.AccessKeySecret = strings.TrimSpace(config.Minio.AccessKeySecret)
	config.Minio.BucketName = strings.TrimSpace(config.Minio.BucketName)
	config.Minio.BucketURL = strings.TrimSpace(config.Minio.BucketURL)
	return config, nil
}
