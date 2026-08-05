package main

import (
	"strings"
	"testing"

	"wklive/deploy/dbinit/models"
)

func TestLoadRuntimeConfigUpload(t *testing.T) {
	t.Setenv("OBJECT_STORAGE_BUCKET", "wklive-dr-backup")
	t.Setenv("OBJECT_STORAGE_OBJECT_KEY", "wklive/mysql/backup.cms")
	t.Setenv("OBJECT_STORAGE_LOCAL_FILE", "/backup/backup.cms")

	cfg, err := loadRuntimeConfig([]string{"upload"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != "upload" ||
		cfg.Bucket != "wklive-dr-backup" ||
		cfg.ObjectKey != "wklive/mysql/backup.cms" ||
		cfg.LocalFile != "/backup/backup.cms" {
		t.Fatalf("unexpected runtime config: %+v", cfg)
	}
}

func TestLoadRuntimeConfigRejectsUnsafePaths(t *testing.T) {
	t.Setenv("OBJECT_STORAGE_BUCKET", "wklive-dr-backup")
	t.Setenv("OBJECT_STORAGE_OBJECT_KEY", "../backup.cms")
	t.Setenv("OBJECT_STORAGE_LOCAL_FILE", "/private/tmp/backup.cms")

	_, err := loadRuntimeConfig([]string{"upload"})
	if err == nil || !strings.Contains(err.Error(), "traversal") {
		t.Fatalf("expected object-key traversal error, got %v", err)
	}

	t.Setenv("OBJECT_STORAGE_OBJECT_KEY", "wklive/mysql/backup.cms")
	_, err = loadRuntimeConfig([]string{"upload"})
	if err == nil || !strings.Contains(err.Error(), "under /backup") {
		t.Fatalf("expected local path error, got %v", err)
	}
}

func TestValidateSystemConfig(t *testing.T) {
	var config models.SystemObjectStorage
	config.OssType = 3
	config.Minio.Endpoint = "storage.example.test"
	config.Minio.AccessKeyID = "access"
	config.Minio.AccessKeySecret = "secret"
	config.Minio.BucketName = "assets"
	if err := validateSystemConfig(config); err != nil {
		t.Fatal(err)
	}

	config.OssType = 1
	if err := validateSystemConfig(config); err == nil {
		t.Fatal("expected non-S3-compatible storage type to fail")
	}
}

func TestValidateBucketName(t *testing.T) {
	for _, bucket := range []string{
		"wklive-dr-backup",
		"prod.backup.2026",
	} {
		if err := validateBucketName(bucket); err != nil {
			t.Fatalf("valid bucket %q rejected: %v", bucket, err)
		}
	}
	for _, bucket := range []string{
		"EXV",
		"-backup",
		"backup-",
		"ab",
	} {
		if err := validateBucketName(bucket); err == nil {
			t.Fatalf("invalid bucket %q accepted", bucket)
		}
	}
}
