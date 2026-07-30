package models

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindSystemObjectStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT config_value").
		WithArgs(int64(0), "OBJECT_STORAGE").
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(
			`{"oss_type":3,"minio":{"endpoint":" storage.example.test ","access_key_id":" key ","access_key_secret":" secret ","bucket_name":" assets ","bucket_url":" https://assets.example.test "}}`,
		))

	config, err := NewSystemObjectStorageModel(db).FindSystemObjectStorage(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if config.OssType != 3 ||
		config.Minio.Endpoint != "storage.example.test" ||
		config.Minio.AccessKeyID != "key" ||
		config.Minio.AccessKeySecret != "secret" ||
		config.Minio.BucketName != "assets" ||
		config.Minio.BucketURL != "https://assets.example.test" {
		t.Fatalf("unexpected config: %+v", config)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindSystemObjectStorageMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT config_value").
		WithArgs(int64(0), "OBJECT_STORAGE").
		WillReturnError(sql.ErrNoRows)

	_, err = NewSystemObjectStorageModel(db).FindSystemObjectStorage(context.Background())
	if err == nil || err.Error() != "system OBJECT_STORAGE config does not exist" {
		t.Fatalf("unexpected error: %v", err)
	}
}
