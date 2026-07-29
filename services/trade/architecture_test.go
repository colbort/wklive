package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogicDoesNotAccessDatabaseDirectly(t *testing.T) {
	t.Helper()

	forbidden := []string{
		`core/stores/sqlx`,
		`.DB`,
		`TransactCtx`,
		`NewSqlConnFromSession`,
		`QueryRowCtx`,
		`QueryRowsCtx`,
		`ExecCtx`,
		`PrepareCtx`,
		`models.NewT`,
	}

	err := filepath.WalkDir("internal/logic", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		source, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, token := range forbidden {
			if strings.Contains(string(source), token) {
				t.Errorf("%s contains forbidden database dependency %q; move the operation to models", path, token)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
