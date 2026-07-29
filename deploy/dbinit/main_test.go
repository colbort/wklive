package main

import (
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
