package models

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsDuplicateKeyError(t *testing.T) {
	duplicate := &mysql.MySQLError{Number: 1062, Message: "duplicate"}
	if !isDuplicateKeyError(duplicate) {
		t.Fatal("expected duplicate key error")
	}
	if !isDuplicateKeyError(fmt.Errorf("insert inbox: %w", duplicate)) {
		t.Fatal("expected wrapped duplicate key error")
	}
	if isDuplicateKeyError(&mysql.MySQLError{Number: 1406, Message: "data too long"}) {
		t.Fatal("must not ignore non-duplicate mysql errors")
	}
	if isDuplicateKeyError(errors.New("database unavailable")) {
		t.Fatal("must not ignore generic database errors")
	}
}
