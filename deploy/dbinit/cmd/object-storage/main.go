package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"wklive/deploy/dbinit/models"
)

const operationTimeout = 2 * time.Minute

type runtimeConfig struct {
	Mode                    string
	Bucket                  string
	ObjectKey               string
	LocalFile               string
	AllowMutation           bool
	RequirePrivateVersioned bool
}

func main() {
	cfg, err := loadRuntimeConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", loadMySQLDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	storageConfig, err := models.NewSystemObjectStorageModel(db).FindSystemObjectStorage(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err := validateSystemConfig(storageConfig); err != nil {
		log.Fatal(err)
	}

	systemBucket := storageConfig.Minio.BucketName
	if cfg.Bucket == "" {
		cfg.Bucket = systemBucket
	}
	if err := validateBucketName(cfg.Bucket); err != nil {
		log.Fatal(err)
	}

	client, endpoint, err := newStorageClient(storageConfig)
	if err != nil {
		log.Fatal(err)
	}

	switch cfg.Mode {
	case "inspect":
		err = inspect(ctx, client, endpoint, systemBucket, cfg)
	case "ensure-private-versioned":
		err = ensurePrivateVersioned(ctx, client, systemBucket, cfg)
	case "upload":
		err = upload(ctx, client, cfg)
	case "download":
		err = download(ctx, client, cfg)
	default:
		err = fmt.Errorf("unsupported object storage operation %q", cfg.Mode)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func loadRuntimeConfig(args []string) (runtimeConfig, error) {
	if len(args) != 1 {
		return runtimeConfig{}, errors.New(
			"usage: object-storage {inspect|ensure-private-versioned|upload|download}",
		)
	}
	cfg := runtimeConfig{
		Mode:          args[0],
		Bucket:        strings.TrimSpace(os.Getenv("OBJECT_STORAGE_BUCKET")),
		ObjectKey:     strings.TrimSpace(os.Getenv("OBJECT_STORAGE_OBJECT_KEY")),
		LocalFile:     strings.TrimSpace(os.Getenv("OBJECT_STORAGE_LOCAL_FILE")),
		AllowMutation: strings.EqualFold(os.Getenv("OBJECT_STORAGE_ALLOW_MUTATION"), "true"),
		RequirePrivateVersioned: strings.EqualFold(
			os.Getenv("OBJECT_STORAGE_REQUIRE_PRIVATE_VERSIONED"),
			"true",
		),
	}
	if cfg.Mode == "upload" || cfg.Mode == "download" {
		if cfg.Bucket == "" || cfg.ObjectKey == "" || cfg.LocalFile == "" {
			return runtimeConfig{}, errors.New(
				"OBJECT_STORAGE_BUCKET, OBJECT_STORAGE_OBJECT_KEY, and OBJECT_STORAGE_LOCAL_FILE are required",
			)
		}
		if err := validateObjectKey(cfg.ObjectKey); err != nil {
			return runtimeConfig{}, err
		}
		cleanFile := filepath.Clean(cfg.LocalFile)
		if cleanFile != "/backup" && !strings.HasPrefix(cleanFile, "/backup/") {
			return runtimeConfig{}, errors.New("OBJECT_STORAGE_LOCAL_FILE must be under /backup")
		}
		cfg.LocalFile = cleanFile
	}
	return cfg, nil
}

func validateSystemConfig(config models.SystemObjectStorage) error {
	if config.OssType != 3 {
		return errors.New("system OBJECT_STORAGE must select S3-compatible MinIO storage (oss_type=3)")
	}
	if config.Minio.Endpoint == "" ||
		config.Minio.AccessKeyID == "" ||
		config.Minio.AccessKeySecret == "" ||
		config.Minio.BucketName == "" {
		return errors.New("system MinIO endpoint, bucket, access key, and secret key are required")
	}
	return nil
}

func newStorageClient(
	config models.SystemObjectStorage,
) (*minio.Client, string, error) {
	endpoint := config.Minio.Endpoint
	secure := true
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Host != "" {
		switch parsed.Scheme {
		case "https":
			endpoint = parsed.Host
		case "http":
			if !strings.EqualFold(os.Getenv("OBJECT_STORAGE_ALLOW_INSECURE_ENDPOINT"), "true") {
				return nil, "", errors.New("production object storage endpoint must use HTTPS")
			}
			secure = false
			endpoint = parsed.Host
		default:
			return nil, "", fmt.Errorf("unsupported object storage endpoint scheme %q", parsed.Scheme)
		}
	} else if strings.Contains(endpoint, "://") {
		return nil, "", errors.New("invalid object storage endpoint")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.Minio.AccessKeyID,
			config.Minio.AccessKeySecret,
			"",
		),
		Secure: secure,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create S3-compatible client: %w", err)
	}
	scheme := "https"
	if !secure {
		scheme = "http"
	}
	return client, scheme + "://" + endpoint, nil
}

func inspect(
	ctx context.Context,
	client *minio.Client,
	endpoint string,
	systemBucket string,
	cfg runtimeConfig,
) error {
	bucket := cfg.Bucket
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("inspect bucket: %w", err)
	}
	versioning := "missing"
	policy := "unknown"
	if exists {
		versionConfig, versionErr := client.GetBucketVersioning(ctx, bucket)
		if versionErr != nil {
			return fmt.Errorf("inspect bucket versioning: %w", versionErr)
		}
		versioning = versionConfig.Status
		if versioning == "" {
			versioning = "Disabled"
		}
		bucketPolicy, policyErr := client.GetBucketPolicy(ctx, bucket)
		if policyErr != nil {
			return fmt.Errorf("inspect bucket policy: %w", policyErr)
		}
		if strings.TrimSpace(bucketPolicy) == "" {
			policy = "private"
		} else {
			policy = "configured"
		}
	}
	if cfg.RequirePrivateVersioned {
		if !exists {
			return errors.New("required DR backup bucket does not exist")
		}
		if versioning != minio.Enabled {
			return fmt.Errorf(
				"required DR backup bucket versioning is %s, expected Enabled",
				versioning,
			)
		}
		if policy != "private" {
			return fmt.Errorf(
				"required DR backup bucket policy is %s, expected private",
				policy,
			)
		}
		if bucket == systemBucket {
			return errors.New(
				"required DR backup bucket must be separate from the system asset bucket",
			)
		}
	}
	fmt.Printf(
		"OBJECT_STORAGE_RESULT=PASS\n"+
			"OBJECT_STORAGE_ENDPOINT=%s\n"+
			"OBJECT_STORAGE_SYSTEM_BUCKET=%s\n"+
			"OBJECT_STORAGE_BUCKET=%s\n"+
			"OBJECT_STORAGE_BUCKET_EXISTS=%t\n"+
			"OBJECT_STORAGE_VERSIONING=%s\n"+
			"OBJECT_STORAGE_POLICY=%s\n",
		endpoint,
		systemBucket,
		bucket,
		exists,
		versioning,
		policy,
	)
	return nil
}

func ensurePrivateVersioned(
	ctx context.Context,
	client *minio.Client,
	systemBucket string,
	cfg runtimeConfig,
) error {
	if !cfg.AllowMutation {
		return errors.New("OBJECT_STORAGE_ALLOW_MUTATION=true is required")
	}
	if cfg.Bucket == systemBucket {
		return errors.New("DR backup bucket must be separate from the system asset bucket")
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("inspect DR backup bucket: %w", err)
	}
	created := false
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create DR backup bucket: %w", err)
		}
		created = true
	}
	if err := client.EnableVersioning(ctx, cfg.Bucket); err != nil {
		return fmt.Errorf("enable DR backup bucket versioning: %w", err)
	}
	versionConfig, err := client.GetBucketVersioning(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("verify DR backup bucket versioning: %w", err)
	}
	if !versionConfig.Enabled() {
		return fmt.Errorf("DR backup bucket versioning is %q, expected Enabled", versionConfig.Status)
	}
	policy, err := client.GetBucketPolicy(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("verify DR backup bucket policy: %w", err)
	}
	if strings.TrimSpace(policy) != "" {
		return errors.New("DR backup bucket has a bucket policy; expected private bucket with no public policy")
	}

	fmt.Printf(
		"OBJECT_STORAGE_ENSURE_RESULT=PASS\n"+
			"OBJECT_STORAGE_BUCKET=%s\n"+
			"OBJECT_STORAGE_BUCKET_CREATED=%t\n"+
			"OBJECT_STORAGE_VERSIONING=Enabled\n"+
			"OBJECT_STORAGE_POLICY=private\n",
		cfg.Bucket,
		created,
	)
	return nil
}

func upload(ctx context.Context, client *minio.Client, cfg runtimeConfig) error {
	info, err := client.FPutObject(
		ctx,
		cfg.Bucket,
		cfg.ObjectKey,
		cfg.LocalFile,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return fmt.Errorf("upload encrypted DR object: %w", err)
	}
	fmt.Printf(
		"OBJECT_STORAGE_UPLOAD_RESULT=PASS\nOBJECT_STORAGE_BUCKET=%s\n"+
			"OBJECT_STORAGE_OBJECT_KEY=%s\nOBJECT_STORAGE_OBJECT_SIZE=%d\n",
		cfg.Bucket,
		cfg.ObjectKey,
		info.Size,
	)
	return nil
}

func download(ctx context.Context, client *minio.Client, cfg runtimeConfig) error {
	if _, err := os.Stat(cfg.LocalFile); err == nil {
		return errors.New("download target already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect download target: %w", err)
	}
	if err := client.FGetObject(
		ctx,
		cfg.Bucket,
		cfg.ObjectKey,
		cfg.LocalFile,
		minio.GetObjectOptions{},
	); err != nil {
		return fmt.Errorf("download encrypted DR object: %w", err)
	}
	info, err := os.Stat(cfg.LocalFile)
	if err != nil {
		return fmt.Errorf("inspect downloaded DR object: %w", err)
	}
	fmt.Printf(
		"OBJECT_STORAGE_DOWNLOAD_RESULT=PASS\nOBJECT_STORAGE_BUCKET=%s\n"+
			"OBJECT_STORAGE_OBJECT_KEY=%s\nOBJECT_STORAGE_OBJECT_SIZE=%d\n",
		cfg.Bucket,
		cfg.ObjectKey,
		info.Size(),
	)
	return nil
}

func validateBucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 63 {
		return errors.New("object storage bucket name must contain 3-63 characters")
	}
	for _, char := range bucket {
		if (char < 'a' || char > 'z') &&
			(char < '0' || char > '9') &&
			char != '.' && char != '-' {
			return errors.New("object storage bucket name must use lowercase DNS-safe characters")
		}
	}
	if bucket[0] == '.' || bucket[0] == '-' ||
		bucket[len(bucket)-1] == '.' || bucket[len(bucket)-1] == '-' {
		return errors.New("object storage bucket name must start and end with a letter or digit")
	}
	return nil
}

func validateObjectKey(key string) error {
	if strings.HasPrefix(key, "/") || strings.Contains(key, "..") ||
		strings.ContainsAny(key, "\r\n\t") {
		return errors.New("object storage key must be relative and contain no traversal or control characters")
	}
	return nil
}

func loadMySQLDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN")); dsn != "" {
		return dsn
	}
	password := strings.TrimSpace(os.Getenv("MYSQL_PASSWORD"))
	if password == "" {
		password = strings.TrimSpace(os.Getenv("MYSQL_ROOT_PASSWORD"))
	}
	cfg := mysql.NewConfig()
	cfg.User = getenv("MYSQL_USER", "root")
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(
		getenv("MYSQL_HOST", "mysql"),
		getenv("MYSQL_PORT", "3306"),
	)
	cfg.DBName = getenv("MYSQL_DATABASE", "wklive")
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.ParseTime = true
	cfg.Loc = time.Local
	return cfg.FormatDSN()
}

func getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
