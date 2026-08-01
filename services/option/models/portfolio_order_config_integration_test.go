package models

import (
	"context"
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestPortfolioOrderConfigEvidenceMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_PORTFOLIO_ORDER_EVIDENCE_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_PORTFOLIO_ORDER_EVIDENCE_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	const tenantID int64 = 919004

	cleanup := func() {
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_order WHERE tenant_id=?", tenantID)
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_portfolio_risk_config WHERE tenant_id=?", tenantID)
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_contract WHERE tenant_id=?", tenantID)
	}
	cleanup()
	t.Cleanup(cleanup)

	portfolioContractResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_contract
(tenant_id,contract_code,settle_coin,seller_margin_mode,create_times,update_times)
VALUES(?, 'PORTFOLIO-EVIDENCE', 'USDT', 3, 900, 900)`, tenantID)
	if err != nil {
		t.Fatalf("insert portfolio contract: %v", err)
	}
	portfolioContractID, err := portfolioContractResult.LastInsertId()
	if err != nil {
		t.Fatalf("portfolio contract id: %v", err)
	}

	isolatedContractResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_contract
(tenant_id,contract_code,settle_coin,seller_margin_mode,create_times,update_times)
VALUES(?, 'ISOLATED-EVIDENCE', 'USDT', 2, 900, 900)`, tenantID)
	if err != nil {
		t.Fatalf("insert isolated contract: %v", err)
	}
	isolatedContractID, err := isolatedContractResult.LastInsertId()
	if err != nil {
		t.Fatalf("isolated contract id: %v", err)
	}

	configResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_portfolio_risk_config
(tenant_id,settle_coin,version,status,effective_from,effective_until,create_times,update_times)
VALUES(?, 'USDT', 7, 2, 1000, 2000, 900, 900)`, tenantID)
	if err != nil {
		t.Fatalf("insert portfolio config: %v", err)
	}
	configID, err := configResult.LastInsertId()
	if err != nil {
		t.Fatalf("portfolio config id: %v", err)
	}

	expectOrderInsertError(t, conn, tenantID, "MISSING-EVIDENCE", portfolioContractID, 2, 0, 0, 1000)
	expectOrderInsertError(t, conn, tenantID, "WRONG-VERSION", portfolioContractID, 2, configID, 8, 1000)
	expectOrderInsertError(t, conn, tenantID, "EXPIRED-BOUNDARY", portfolioContractID, 2, configID, 7, 2000)
	expectOrderInsertError(t, conn, tenantID, "BUY-WITH-EVIDENCE", portfolioContractID, 1, configID, 7, 1000)
	expectOrderInsertError(t, conn, tenantID, "ISOLATED-WITH-EVIDENCE", isolatedContractID, 2, configID, 7, 1000)
	expectOrderInsertError(t, conn, tenantID, "UNPAIRED-EVIDENCE", portfolioContractID, 2, configID, 0, 1000)

	accepted, err := conn.ExecCtx(ctx, `INSERT INTO t_option_order
(tenant_id,order_no,contract_id,side,portfolio_risk_config_id,portfolio_risk_config_version,status,create_times,update_times)
VALUES(?, 'VALID-EVIDENCE', ?, 2, ?, 7, 1, 1000, 1000)`, tenantID, portfolioContractID, configID)
	if err != nil {
		t.Fatalf("insert valid portfolio seller order: %v", err)
	}
	orderID, err := accepted.LastInsertId()
	if err != nil {
		t.Fatalf("valid order id: %v", err)
	}

	if _, err := conn.ExecCtx(ctx, `UPDATE t_option_order
SET portfolio_risk_config_version=8,update_times=1001 WHERE id=?`, orderID); err == nil {
		t.Fatal("expected immutable portfolio evidence update to fail")
	}
	if _, err := conn.ExecCtx(ctx, `UPDATE t_option_order
SET status=4,cancel_time=1001,update_times=1001 WHERE id=?`, orderID); err != nil {
		t.Fatalf("normal order lifecycle update must retain evidence: %v", err)
	}

	if _, err := conn.ExecCtx(ctx, `INSERT INTO t_option_order
(tenant_id,order_no,contract_id,side,portfolio_risk_config_id,portfolio_risk_config_version,status,create_times,update_times)
VALUES(?, 'ISOLATED-ZERO-EVIDENCE', ?, 2, 0, 0, 1, 1000, 1000)`, tenantID, isolatedContractID); err != nil {
		t.Fatalf("isolated seller order with 0/0 evidence: %v", err)
	}
}

func expectOrderInsertError(
	t *testing.T,
	conn sqlx.SqlConn,
	tenantID int64,
	orderNo string,
	contractID int64,
	side int64,
	configID int64,
	configVersion int64,
	createTimes int64,
) {
	t.Helper()
	_, err := conn.ExecCtx(context.Background(), `INSERT INTO t_option_order
(tenant_id,order_no,contract_id,side,portfolio_risk_config_id,portfolio_risk_config_version,status,create_times,update_times)
VALUES(?,?,?,?,?,?,1,?,?)`, tenantID, orderNo, contractID, side, configID, configVersion, createTimes, createTimes)
	if err == nil {
		t.Fatalf("order %s: expected insert to fail", orderNo)
	}
}
