package models

import (
	"database/sql"
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

func TestResolveSnapshotInboxClaim(t *testing.T) {
	claimed, err := resolveSnapshotInboxClaim(fakeSQLResult{rowsAffected: 1}, nil)
	if err != nil || !claimed {
		t.Fatalf("successful claim = (%v, %v), want (true, nil)", claimed, err)
	}

	claimed, err = resolveSnapshotInboxClaim(nil, &mysql.MySQLError{Number: 1062, Message: "duplicate"})
	if err != nil || claimed {
		t.Fatalf("duplicate claim = (%v, %v), want (false, nil)", claimed, err)
	}

	wantErr := errors.New("database unavailable")
	claimed, err = resolveSnapshotInboxClaim(nil, wantErr)
	if claimed || !errors.Is(err, wantErr) {
		t.Fatalf("failed claim = (%v, %v), want (false, %v)", claimed, err, wantErr)
	}
}

type fakeSQLResult struct {
	rowsAffected int64
	err          error
}

func (r fakeSQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r fakeSQLResult) RowsAffected() (int64, error) {
	return r.rowsAffected, r.err
}

var _ sql.Result = fakeSQLResult{}
