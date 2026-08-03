package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
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
	ProductionOperators    []operatorCredential
	DataOnly               bool
}

type operatorCredential struct {
	Username string
	Password string
	Nickname string
	RoleCode string
}

type migration struct {
	version      string
	path         string
	checksum     string
	baselineSafe bool
}

func main() {
	cfg := loadConfig()
	db := waitForDB(cfg.DSN)
	defer db.Close()

	ctx := context.Background()
	fresh, err := isFreshDatabase(ctx, db)
	must(err)
	if cfg.DataOnly {
		if fresh {
			log.Fatal("data merge requires an initialized database schema; run db-init first")
		}
		must(mergeInitializationData(ctx, db, cfg))
		log.Printf("database initialization data merged")
		return
	}
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
	}
	must(applyPendingMigrations(ctx, db, migrations))

	must(mergeInitializationData(ctx, db, cfg))

	log.Printf("database initialization completed: fresh=%t migrations=%d", fresh, len(migrations))
}

func loadConfig() config {
	cfg := config{
		DSN:                    loadMySqlDSN(),
		Workspace:              getenv("WORKSPACE", "/workspace"),
		AdminUsername:          getenv("ADMIN_USERNAME", "admin"),
		AdminPassword:          os.Getenv("ADMIN_PASSWORD"),
		LiquidityAdminUsername: getenv("LIQUIDITY_ADMIN_USERNAME", "liquidityadmin"),
		LiquidityAdminPassword: os.Getenv("LIQUIDITY_ADMIN_PASSWORD"),
		DataOnly:               strings.EqualFold(getenv("DB_INIT_MODE", "full"), "data"),
	}
	validateCredential("ADMIN_USERNAME", cfg.AdminUsername, 3)
	validateCredential("ADMIN_PASSWORD", cfg.AdminPassword, 12)
	validateCredential("LIQUIDITY_ADMIN_USERNAME", cfg.LiquidityAdminUsername, 3)
	validateCredential("LIQUIDITY_ADMIN_PASSWORD", cfg.LiquidityAdminPassword, 12)
	if strings.EqualFold(getenv("PRODUCTION_OPERATOR_SEED_ENABLED", "false"), "true") {
		cfg.ProductionOperators = []operatorCredential{
			loadOperatorCredential(
				"CONTRACT_ONCALL", "contract_oncall", "合约生产值班", "contract_oncall",
			),
			loadOperatorCredential(
				"INSURANCE_OPERATOR", "insurance_operator", "保险基金操作员", "insurance_fund_operator",
			),
			loadOperatorCredential(
				"DR_OPERATOR", "dr_operator", "灾备操作员", "disaster_recovery_operator",
			),
			loadOperatorCredential(
				"DELIVERY_OPERATOR", "delivery_operator", "交割发布操作员", "delivery_release_operator",
			),
			loadOperatorCredential(
				"PRODUCTION_REVIEWER", "production_reviewer", "生产发布复核员", "production_reviewer",
			),
			loadOperatorCredential(
				"PRODUCTION_APPROVER", "production_approver", "生产发布审批员", "production_approver",
			),
		}
	}
	return cfg
}

func loadOperatorCredential(
	prefix string,
	defaultUsername string,
	nickname string,
	roleCode string,
) operatorCredential {
	usernameKey := prefix + "_USERNAME"
	passwordKey := prefix + "_PASSWORD"
	username := getenv(usernameKey, defaultUsername)
	password := os.Getenv(passwordKey)
	validateCredential(usernameKey, username, 3)
	validateCredential(passwordKey, password, 20)
	return operatorCredential{
		Username: username,
		Password: password,
		Nickname: nickname,
		RoleCode: roleCode,
	}
}

func loadMySqlDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN")); dsn != "" {
		return dsn
	}

	defaultHost := "mysql"
	if strings.EqualFold(getenv("DB_INIT_TARGET", "compose"), "external") {
		defaultHost = getenv("MYSQL_EXTERNAL_HOST", "host.docker.internal")
	}
	password := strings.TrimSpace(os.Getenv("MYSQL_PASSWORD"))
	if password == "" {
		password = getenv("MYSQL_ROOT_PASSWORD", "123456")
	}

	cfg := mysql.NewConfig()
	cfg.User = getenv("MYSQL_USER", "root")
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(getenv("MYSQL_HOST", defaultHost), getenv("MYSQL_PORT", "3306"))
	cfg.DBName = getenv("MYSQL_DATABASE", "wklive")
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.MultiStatements = true
	return cfg.FormatDSN()
}

func validateCredential(name, value string, min int) {
	if len(strings.TrimSpace(value)) < min {
		log.Fatalf("%s must contain at least %d characters", name, min)
	}
}

func mergeInitializationData(ctx context.Context, db *sql.DB, cfg config) error {
	if err := seedMenus(ctx, db, filepath.Join(cfg.Workspace, "init.sql")); err != nil {
		return err
	}
	if err := applySystemDataMigrations(ctx, db, cfg.Workspace); err != nil {
		return err
	}
	if err := seedAdministrator(
		ctx, db, 1, cfg.AdminUsername, cfg.AdminPassword,
		"超级管理员", "super_admin", "超级管理员",
	); err != nil {
		return err
	}
	if err := seedAdministrator(
		ctx, db, 2, cfg.LiquidityAdminUsername, cfg.LiquidityAdminPassword,
		"做市管理员", "liquidity_admin", "做市管理后台管理员",
	); err != nil {
		return err
	}
	for _, operator := range cfg.ProductionOperators {
		if err := seedScopedOperator(ctx, db, operator); err != nil {
			return err
		}
	}
	return nil
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
				version:      filepath.ToSlash(relative),
				path:         path,
				checksum:     hex.EncodeToString(sum[:]),
				baselineSafe: bytes.Contains(data, []byte("-- dbinit:baseline-safe")),
			})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].version < result[j].version })
	return result, nil
}

func baselineMigrations(ctx context.Context, db *sql.DB, migrations []migration) error {
	recorded := 0
	pending := 0
	for _, item := range migrations {
		if item.baselineSafe {
			pending++
			continue
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO schema_migrations(version, checksum, applied_at) VALUES (?, ?, ?)`,
			item.version, item.checksum, time.Now().UnixMilli()); err != nil {
			return err
		}
		recorded++
	}
	log.Printf(
		"recorded current schema as migration baseline: recorded=%d pending_safe=%d",
		recorded,
		pending,
	)
	return nil
}

func applyPendingMigrations(ctx context.Context, db *sql.DB, migrations []migration) error {
	for _, item := range migrations {
		var checksum string
		err := db.QueryRowContext(ctx,
			`SELECT checksum FROM schema_migrations WHERE version = ?`, item.version).Scan(&checksum)
		if errors.Is(err, sql.ErrNoRows) {
			if legacyVersion, ok := legacyMigrationVersion(item.version); ok {
				err = db.QueryRowContext(ctx,
					`SELECT checksum FROM schema_migrations WHERE version = ?`, legacyVersion).Scan(&checksum)
				switch {
				case err == nil:
					if checksum != item.checksum {
						return fmt.Errorf("applied migration was modified: %s (renamed from %s)",
							item.version, legacyVersion)
					}
					if _, err = db.ExecContext(ctx,
						`INSERT INTO schema_migrations(version, checksum, applied_at) VALUES (?, ?, ?)`,
						item.version, item.checksum, time.Now().UnixMilli()); err != nil {
						return err
					}
					log.Printf("recorded renamed migration: %s -> %s", legacyVersion, item.version)
					continue
				case !errors.Is(err, sql.ErrNoRows):
					return err
				}
			}
		}
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

func legacyMigrationVersion(version string) (string, bool) {
	const marketPrefix = "services/market/"
	if !strings.HasPrefix(version, marketPrefix) {
		return "", false
	}
	return "services/itick/" + strings.TrimPrefix(version, marketPrefix), true
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
		"20260728_add_contract_reconciliation_admin_menu.sql",
		"20260728_add_cross_account_liquidation_admin_menu.sql",
		"20260728_add_trade_contract_jobs.sql",
		"20260729_add_itick_authority_registry_permissions.sql",
		"20260730_add_contract_production_roles.sql",
		"20260730_add_option_control_permissions.sql",
		"20260730_add_option_exercise_retry_menu.sql",
		"20260730_add_option_jobs.sql",
		"20260730_add_option_risk_menu.sql",
		"20260730_add_option_settlement_retry_menu.sql",
		"20260730_zj_option_operations_permissions.sql",
		"20260731_zl_option_trading_calendar_permissions.sql",
		"20260731_zm_option_corporate_action_permissions.sql",
		"20260731_zn_option_contract_series_permissions.sql",
		"20260731_zq_option_combo_operations_permissions.sql",
		"20260731_zt_option_daily_reconciliation_job.sql",
		"20260802_asset_platform_backstop_policy_permissions.sql",
		"20260802_option_insurance_inventory_exit_permissions.sql",
		"20260802_zz_admin_extension_permissions.sql",
		"20260802_zzz_remove_tenant_display_init_menu.sql",
		"20260802_zzzz_split_option_risk_workbench.sql",
		"20260803_add_staking_jobs.sql",
		"20260803_add_staking_operation_menu.sql",
		"20260803_add_staking_reconciliation_job_menu.sql",
		"20260803_refresh_staking_admin_permissions.sql",
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

func seedScopedOperator(ctx context.Context, db *sql.DB, operator operatorCredential) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(operator.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var roleID int64
	if err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM sys_role
		WHERE tenant_id = 0 AND app_scope = 1 AND code = ? AND enabled = 1`,
		operator.RoleCode).Scan(&roleID); err != nil {
		return fmt.Errorf("find production role %s: %w", operator.RoleCode, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO sys_user
		  (tenant_id, app_scope, user_type, is_owner, username, password, nickname, avatar,
		   enabled, google_secret, google_enabled, perms_ver, last_login_ip, last_login_at,
		   create_by, create_times, update_times)
		VALUES (0, 1, 1, 2, ?, ?, ?, '', 1, '', 2, 1, '', 0, 0, ?, ?)
		ON DUPLICATE KEY UPDATE
		  password = VALUES(password), nickname = VALUES(nickname), enabled = 1,
		  app_scope = 1, update_times = VALUES(update_times)`,
		operator.Username, string(hash), operator.Nickname, now, now)
	if err != nil {
		return err
	}

	var userID int64
	if err = tx.QueryRowContext(ctx,
		`SELECT id FROM sys_user WHERE tenant_id = 0 AND app_scope = 1 AND username = ?`,
		operator.Username).Scan(&userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		DELETE FROM sys_user_role
		WHERE tenant_id = 0 AND user_id = ? AND role_id <> ?`,
		userID, roleID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT IGNORE INTO sys_user_role(tenant_id, user_id, role_id) VALUES (0, ?, ?)`,
		userID, roleID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	log.Printf(
		"seeded production operator: username=%s role=%s",
		operator.Username,
		operator.RoleCode,
	)
	return nil
}

func execSQLFile(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, statement := range splitSQLScript(string(data)) {
		if _, err = db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

// splitSQLScript removes mysql-client DELIMITER directives and returns
// executable protocol statements. Ordinary SQL remains grouped so existing
// multi-statement migrations keep their transaction and session variables;
// stored programs and triggers using a custom delimiter are emitted one by one.
func splitSQLScript(script string) []string {
	var statements []string
	var current strings.Builder
	delimiter := ""
	flush := func() {
		statement := strings.TrimSpace(current.String())
		current.Reset()
		if statement != "" {
			statements = append(statements, statement)
		}
	}

	for _, line := range strings.SplitAfter(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "DELIMITER ") {
			flush()
			delimiter = strings.TrimSpace(trimmed[len("DELIMITER "):])
			if delimiter == ";" {
				delimiter = ""
			}
			continue
		}

		if delimiter == "" {
			current.WriteString(line)
			continue
		}

		lineWithoutNewline := strings.TrimSuffix(line, "\n")
		lineWithoutNewline = strings.TrimSuffix(lineWithoutNewline, "\r")
		trimmedRight := strings.TrimRight(lineWithoutNewline, " \t")
		if strings.HasSuffix(trimmedRight, delimiter) {
			current.WriteString(strings.TrimSuffix(trimmedRight, delimiter))
			current.WriteByte('\n')
			flush()
			continue
		}
		current.WriteString(line)
	}
	flush()
	return statements
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
