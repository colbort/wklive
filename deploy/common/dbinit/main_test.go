package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
)

func TestLoadMySqlDSNUsesExternalTargetDefaults(t *testing.T) {
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("DB_INIT_TARGET", "external")
	t.Setenv("MYSQL_EXTERNAL_HOST", "host.docker.internal")
	t.Setenv("MYSQL_HOST", "")
	t.Setenv("MYSQL_PORT", "3307")
	t.Setenv("MYSQL_USER", "wklive")
	t.Setenv("MYSQL_PASSWORD", "secret")
	t.Setenv("MYSQL_DATABASE", "trade")

	cfg, err := mysql.ParseDSN(loadMySqlDSN())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "host.docker.internal:3307" {
		t.Fatalf("addr=%q", cfg.Addr)
	}
	if cfg.User != "wklive" || cfg.Passwd != "secret" || cfg.DBName != "trade" {
		t.Fatalf("unexpected mysql config: user=%q database=%q", cfg.User, cfg.DBName)
	}
	if !cfg.ParseTime || !cfg.MultiStatements {
		t.Fatalf("required mysql options are disabled: parseTime=%t multiStatements=%t", cfg.ParseTime, cfg.MultiStatements)
	}
}

func TestLoadMySqlDSNPrefersExplicitDSN(t *testing.T) {
	const dsn = "custom:secret@tcp(db.example:3306)/existing?parseTime=true"
	t.Setenv("MYSQL_DSN", dsn)
	if got := loadMySqlDSN(); got != dsn {
		t.Fatalf("dsn=%q want=%q", got, dsn)
	}
}

func TestFindMigrationsMarksOnlyExplicitBaselineSafeFiles(t *testing.T) {
	workspace := t.TempDir()
	migrationsDir := filepath.Join(workspace, "services", "trade", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacyPath := filepath.Join(migrationsDir, "20260728_legacy.sql")
	safePath := filepath.Join(migrationsDir, "20260729_reconcile.sql")
	if err := os.WriteFile(legacyPath, []byte("SELECT 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		safePath,
		[]byte("-- dbinit:baseline-safe\nSELECT 1;\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	migrations, err := findMigrations(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != 2 {
		t.Fatalf("migrations=%d want=2", len(migrations))
	}
	if migrations[0].baselineSafe {
		t.Fatalf("legacy migration unexpectedly baseline safe: %s", migrations[0].version)
	}
	if !migrations[1].baselineSafe {
		t.Fatalf("reconciliation migration not baseline safe: %s", migrations[1].version)
	}
}

func TestLegacyMigrationVersion(t *testing.T) {
	got, ok := legacyMigrationVersion(
		"services/market/migrations/20260722_add_snapshot_outbox_cleanup_index.sql",
	)
	if !ok {
		t.Fatal("market migration should have a legacy version")
	}
	const want = "services/itick/migrations/20260722_add_snapshot_outbox_cleanup_index.sql"
	if got != want {
		t.Fatalf("legacy version=%q want=%q", got, want)
	}

	if got, ok = legacyMigrationVersion("services/trade/migrations/20260728_example.sql"); ok {
		t.Fatalf("trade migration unexpectedly has legacy version: %q", got)
	}
}

func TestSplitSQLScriptSupportsMySQLDelimiter(t *testing.T) {
	script := `SET @before = 1;
DELIMITER $$
CREATE TRIGGER trg_example BEFORE INSERT ON example
FOR EACH ROW
BEGIN
  SET NEW.value = 1;
  SET NEW.updated = 2;
END$$
CREATE PROCEDURE sp_example()
BEGIN
  SELECT 1;
END$$
DELIMITER ;
SET @after = 2;
`

	statements := splitSQLScript(script)
	if len(statements) != 4 {
		t.Fatalf("statements=%d want=4: %#v", len(statements), statements)
	}
	if statements[0] != "SET @before = 1;" {
		t.Fatalf("prefix=%q", statements[0])
	}
	if strings.Contains(statements[1], "DELIMITER") || !strings.HasSuffix(statements[1], "END") {
		t.Fatalf("trigger statement=%q", statements[1])
	}
	if !strings.Contains(statements[2], "CREATE PROCEDURE") {
		t.Fatalf("procedure statement=%q", statements[2])
	}
	if statements[3] != "SET @after = 2;" {
		t.Fatalf("suffix=%q", statements[3])
	}
}

func TestParsePositiveInt64(t *testing.T) {
	if got, err := parsePositiveInt64("TENANT_ID", " 7 "); err != nil || got != 7 {
		t.Fatalf("got=%d err=%v", got, err)
	}
	for _, value := range []string{"", "0", "-1", "tenant"} {
		if _, err := parsePositiveInt64("TENANT_ID", value); err == nil {
			t.Fatalf("value %q unexpectedly accepted", value)
		}
	}
}

func TestPreprodBootstrapIsLimitedToApprovedSymbols(t *testing.T) {
	path := filepath.Join("..", "..", "environments", "preprod", "data", "bootstrap.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	if strings.Contains(script, "LAST_INSERT_ID(id)") {
		t.Fatal("pre-production bootstrap contains an ambiguous duplicate-key id reference")
	}
	if !strings.Contains(script, "COLLATE utf8mb4_general_ci") {
		t.Fatal("pre-production market seed must match the legacy market table collation")
	}
	for _, table := range []string{
		"sys_tenant", "t_itick_tenant_product", "t_trade_symbol",
		"t_option_contract", "t_stake_product",
	} {
		if !strings.Contains(script, table) {
			t.Fatalf("pre-production bootstrap does not seed %s", table)
		}
	}

	allowed := map[string]bool{"BTCUSDT": true, "ETHUSDT": true}
	matches := regexp.MustCompile(`'([A-Z0-9]+USDT)'`).FindAllStringSubmatch(script, -1)
	seen := make(map[string]bool)
	for _, match := range matches {
		seen[match[1]] = true
		if !allowed[match[1]] {
			t.Fatalf("unexpected pre-production quote symbol: %s", match[1])
		}
	}
	for symbol := range allowed {
		if !seen[symbol] {
			t.Fatalf("missing pre-production quote symbol: %s", symbol)
		}
	}
}
