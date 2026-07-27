package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const migrationTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) NOT NULL,
  checksum CHAR(64) NOT NULL,
  applied_at BIGINT NOT NULL,
  PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Docker deployment migration history'`

type config struct {
	DSN                    string
	Workspace              string
	AdminUsername          string
	AdminPassword          string
	LiquidityAdminUsername string
	LiquidityAdminPassword string
}

type migration struct {
	version  string
	path     string
	checksum string
}

func main() {
	cfg := loadConfig()
	db := waitForDB(cfg.DSN)
	defer db.Close()

	ctx := context.Background()
	fresh, err := isFreshDatabase(ctx, db)
	must(err)
	if fresh {
		must(loadBaseSchemas(ctx, db, cfg.Workspace))
	}

	existed, err := migrationTableExists(ctx, db)
	must(err)
	_, err = db.ExecContext(ctx, migrationTable)
	must(err)

	migrations, err := findMigrations(cfg.Workspace)
	must(err)
	if !existed {
		must(baselineMigrations(ctx, db, migrations))
	} else {
		must(applyPendingMigrations(ctx, db, migrations))
	}

	must(seedMenus(ctx, db, filepath.Join(cfg.Workspace, "init.sql")))
	must(applySystemDataMigrations(ctx, db, cfg.Workspace))
	must(seedAdministrator(ctx, db, 1, cfg.AdminUsername, cfg.AdminPassword, "超级管理员", "super_admin", "超级管理员"))
	must(seedAdministrator(ctx, db, 2, cfg.LiquidityAdminUsername, cfg.LiquidityAdminPassword, "做市管理员", "liquidity_admin", "做市管理后台管理员"))

	log.Printf("database initialization completed: fresh=%t migrations=%d", fresh, len(migrations))
}

func loadConfig() config {
	mysqlPassword := getenv("MYSQL_ROOT_PASSWORD", "123456")
	cfg := config{
		DSN:                    fmt.Sprintf("root:%s@tcp(mysql:3306)/wklive?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true", mysqlPassword),
		Workspace:              getenv("WORKSPACE", "/workspace"),
		AdminUsername:          getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		LiquidityAdminUsername: getenv("LIQUIDITY_ADMIN_USERNAME", "liquidityadmin"),
		LiquidityAdminPassword: os.Getenv("LIQUIDITY_ADMIN_PASSWORD"),
	}
	validateCredential("ADMIN_USERNAME", cfg.AdminUsername, 3)
	validateCredential("ADMIN_PASSWORD", cfg.AdminPassword, 12)
	validateCredential("LIQUIDITY_ADMIN_USERNAME", cfg.LiquidityAdminUsername, 3)
	validateCredential("LIQUIDITY_ADMIN_PASSWORD", cfg.LiquidityAdminPassword, 12)
	return cfg
}

func validateCredential(name, value string, min int) {
	if len(strings.TrimSpace(value)) < min {
		log.Fatalf("%s must contain at least %d characters", name, min)
	}
}

func waitForDB(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	must(err)
	for attempt := 1; attempt <= 60; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return db
		}
		time.Sleep(time.Second)
	}
	log.Fatalf("mysql did not become ready: %v", err)
	return nil
}

func isFreshDatabase(ctx context.Context, db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_name <> 'schema_migrations'`).Scan(&count)
	return count == 0, err
}

func loadBaseSchemas(ctx context.Context, db *sql.DB, workspace string) error {
	serviceDirs, err := filepath.Glob(filepath.Join(workspace, "services", "*"))
	if err != nil {
		return err
	}
	sort.Strings(serviceDirs)
	loaded := 0
	for _, serviceDir := range serviceDirs {
		info, statErr := os.Stat(serviceDir)
		if statErr != nil || !info.IsDir() {
			continue
		}
		schema := filepath.Join(serviceDir, filepath.Base(serviceDir)+".sql")
		if _, statErr = os.Stat(schema); errors.Is(statErr, fs.ErrNotExist) {
			continue
		}
		log.Printf("loading base schema: %s", schema)
		if err = execSQLFile(ctx, db, schema); err != nil {
			return fmt.Errorf("load base schema %s: %w", schema, err)
		}
		loaded++
	}
	if loaded == 0 {
		return errors.New("no base schema files found")
	}
	return nil
}

func migrationTableExists(ctx context.Context, db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_name = 'schema_migrations'`).Scan(&count)
	return count == 1, err
}

func findMigrations(workspace string) ([]migration, error) {
	patterns := []string{
		filepath.Join(workspace, "services", "*", "migrations", "*.sql"),
		filepath.Join(workspace, "services", "*", "*migration*.sql"),
	}
	seen := make(map[string]struct{})
	var result []migration
	for _, pattern := range patterns {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			relative, err := filepath.Rel(workspace, path)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[relative]; ok {
				continue
			}
			seen[relative] = struct{}{}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			sum := sha256.Sum256(data)
			result = append(result, migration{
				version:  filepath.ToSlash(relative),
				path:     path,
				checksum: hex.EncodeToString(sum[:]),
			})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].version < result[j].version })
	return result, nil
}

func baselineMigrations(ctx context.Context, db *sql.DB, migrations []migration) error {
	for _, item := range migrations {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO schema_migrations(version, checksum, applied_at) VALUES (?, ?, ?)`,
			item.version, item.checksum, time.Now().UnixMilli()); err != nil {
			return err
		}
	}
	log.Printf("recorded current schema as migration baseline")
	return nil
}

func applyPendingMigrations(ctx context.Context, db *sql.DB, migrations []migration) error {
	for _, item := range migrations {
		var checksum string
		err := db.QueryRowContext(ctx,
			`SELECT checksum FROM schema_migrations WHERE version = ?`, item.version).Scan(&checksum)
		switch {
		case err == nil:
			if checksum != item.checksum {
				return fmt.Errorf("applied migration was modified: %s", item.version)
			}
			continue
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}

		log.Printf("applying migration: %s", item.version)
		if err = execSQLFile(ctx, db, item.path); err != nil {
			return fmt.Errorf("apply migration %s: %w", item.version, err)
		}
		if _, err = db.ExecContext(ctx,
			`INSERT INTO schema_migrations(version, checksum, applied_at) VALUES (?, ?, ?)`,
			item.version, item.checksum, time.Now().UnixMilli()); err != nil {
			return err
		}
	}
	return nil
}

func seedMenus(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	const marker = "INSERT INTO sys_menu"
	index := strings.Index(string(data), marker)
	if index < 0 {
		return errors.New("admin menu seed marker not found")
	}
	log.Printf("seeding management menus from %s", path)
	seed := strings.ReplaceAll(string(data[index:]), "INSERT INTO sys_menu", "INSERT IGNORE INTO sys_menu")
	_, err = db.ExecContext(ctx, seed)
	return err
}

func applySystemDataMigrations(ctx context.Context, db *sql.DB, workspace string) error {
	// system.sql already contains the current table shape, so structural ALTER
	// migrations are part of the baseline. These migrations own evolving menu,
	// role and scheduled-job data and must run after the legacy init.sql seed.
	files := []string{
		"20260724_add_liquidity_admin_permissions.sql",
		"20260724_add_liquidity_admin_role.sql",
		"20260725_add_liquidity_config_options_permission.sql",
		"20260725_add_liquidity_jobs.sql",
		"20260725_add_liquidity_order_archive_job.sql",
		"20260725_add_liquidity_provider_provision_permission.sql",
		"20260725_remove_tenant_pay_platform_menu.sql",
		"20260727_add_liquidity_strategy_detail_update_permissions.sql",
	}
	for _, name := range files {
		path := filepath.Join(workspace, "services", "system", "migrations", name)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("required system data migration %s: %w", name, err)
		}
		log.Printf("applying repeatable system data migration: %s", name)
		if err := execSQLFile(ctx, db, path); err != nil {
			return fmt.Errorf("apply system data migration %s: %w", name, err)
		}
	}
	return nil
}

func seedAdministrator(
	ctx context.Context,
	db *sql.DB,
	appScope int,
	username string,
	password string,
	nickname string,
	roleCode string,
	roleName string,
) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO sys_role
		  (tenant_id, app_scope, name, code, enabled, remark, create_times, update_times)
		VALUES (0, ?, ?, ?, 1, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  name = VALUES(name), enabled = 1, remark = VALUES(remark), update_times = VALUES(update_times)`,
		appScope, roleName, roleCode, roleName, now, now)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO sys_user
		  (tenant_id, app_scope, user_type, is_owner, username, password, nickname, avatar,
		   enabled, google_secret, google_enabled, perms_ver, last_login_ip, last_login_at,
		   create_by, create_times, update_times)
		VALUES (0, ?, 1, 2, ?, ?, ?, '', 1, '', 2, 1, '', 0, 0, ?, ?)
		ON DUPLICATE KEY UPDATE
		  password = VALUES(password), nickname = VALUES(nickname), enabled = 1,
		  update_times = VALUES(update_times)`,
		appScope, username, string(hash), nickname, now, now)
	if err != nil {
		return err
	}

	var userID, roleID int64
	if err = tx.QueryRowContext(ctx,
		`SELECT id FROM sys_user WHERE tenant_id = 0 AND app_scope = ? AND username = ?`,
		appScope, username).Scan(&userID); err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx,
		`SELECT id FROM sys_role WHERE tenant_id = 0 AND app_scope = ? AND code = ?`,
		appScope, roleCode).Scan(&roleID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT IGNORE INTO sys_user_role(tenant_id, user_id, role_id) VALUES (0, ?, ?)`,
		userID, roleID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT IGNORE INTO sys_role_menu(tenant_id, role_id, menu_id)
		SELECT 0, ?, id FROM sys_menu WHERE app_scope = ?`, roleID, appScope); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	log.Printf("seeded administrator: scope=%d username=%s", appScope, username)
	return nil
}

func execSQLFile(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(data))
	return err
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
