package applogic

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// withPublicMarketSnapshot makes the consistency contract of public market
// responses independent from the database server's configured default.
// Every query used to build one response runs in the same read-only,
// repeatable-read transaction.
func withPublicMarketSnapshot(
	ctx context.Context,
	conn sqlx.SqlConn,
	fn func(context.Context, sqlx.SqlConn) error,
) error {
	db, err := conn.RawDB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	txConn := sqlx.NewSqlConnFromSession(sqlx.NewSessionFromTx(tx))
	if err = fn(ctx, txConn); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
