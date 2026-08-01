package models

import (
	"context"
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestAssetEvidenceModelsRejectMissingCoin(t *testing.T) {
	ctx := context.Background()
	instructionModel := &customTOptionAssetInstructionModel{}
	if _, err := instructionModel.Insert(ctx, &TOptionAssetInstruction{Coin: " "}); err == nil {
		t.Fatal("expected empty asset instruction coin to fail before SQL")
	}
	marginLotModel := &customTOptionMarginLotModel{}
	if _, err := marginLotModel.Insert(ctx, &TOptionMarginLot{CollateralCoin: " "}); err == nil {
		t.Fatal("expected empty margin lot collateral coin to fail before SQL")
	}
}

func TestMarginCoinDatabaseGuardsMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_MARGIN_COIN_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_MARGIN_COIN_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	const tenantID int64 = 919040

	cleanup := func() {
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_asset_instruction WHERE tenant_id=?", tenantID)
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_margin_lot WHERE tenant_id=?", tenantID)
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_order WHERE tenant_id=?", tenantID)
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_contract WHERE tenant_id=?", tenantID)
	}
	cleanup()
	t.Cleanup(cleanup)

	contractResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_contract
(tenant_id,contract_code,underlying_coin,settle_coin,option_type,settlement_type,seller_margin_mode,create_times,update_times)
VALUES(?,'PHYSICAL-CALL-COIN','BTC','USDT',1,2,4,1000,1000)`, tenantID)
	if err != nil {
		t.Fatalf("insert physical Call contract: %v", err)
	}
	contractID, err := contractResult.LastInsertId()
	if err != nil {
		t.Fatalf("contract id: %v", err)
	}

	expectMarginCoinExecError(t, conn, `INSERT INTO t_option_order
(tenant_id,order_no,contract_id,side,position_effect,margin_amount,margin_coin,create_times,update_times)
VALUES(?,'ORDER-EMPTY',?,2,1,1,'',1000,1000)`, tenantID, contractID)
	expectMarginCoinExecError(t, conn, `INSERT INTO t_option_order
(tenant_id,order_no,contract_id,side,position_effect,margin_amount,margin_coin,create_times,update_times)
VALUES(?,'ORDER-WRONG',?,2,1,1,'USDT',1000,1000)`, tenantID, contractID)

	orderResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_order
(tenant_id,order_no,contract_id,side,position_effect,margin_amount,margin_coin,create_times,update_times)
VALUES(?,'ORDER-CORRECT',?,2,1,1,'BTC',1000,1000)`, tenantID, contractID)
	if err != nil {
		t.Fatalf("insert correctly denominated order: %v", err)
	}
	orderID, err := orderResult.LastInsertId()
	if err != nil {
		t.Fatalf("order id: %v", err)
	}
	expectMarginCoinExecError(t, conn,
		"UPDATE t_option_order SET margin_coin='USDT' WHERE id=?", orderID)

	expectMarginCoinExecError(t, conn, `INSERT INTO t_option_margin_lot
(tenant_id,contract_id,order_id,trade_id,collateral_coin,quantity,remaining_quantity,initial_margin,remaining_margin,status,create_times,update_times)
VALUES(?,?,?,1,'',1,1,1,1,1,1000,1000)`, tenantID, contractID, orderID)
	expectMarginCoinExecError(t, conn, `INSERT INTO t_option_margin_lot
(tenant_id,contract_id,order_id,trade_id,collateral_coin,quantity,remaining_quantity,initial_margin,remaining_margin,status,create_times,update_times)
VALUES(?,?,?,2,'USDT',1,1,1,1,1,1000,1000)`, tenantID, contractID, orderID)
	lotResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_margin_lot
(tenant_id,contract_id,order_id,trade_id,collateral_coin,quantity,remaining_quantity,initial_margin,remaining_margin,status,create_times,update_times)
VALUES(?,?,?,3,'BTC',1,1,1,1,1,1000,1000)`, tenantID, contractID, orderID)
	if err != nil {
		t.Fatalf("insert correctly denominated margin lot: %v", err)
	}
	lotID, err := lotResult.LastInsertId()
	if err != nil {
		t.Fatalf("margin lot id: %v", err)
	}
	expectMarginCoinExecError(t, conn,
		"UPDATE t_option_margin_lot SET collateral_coin='USDT' WHERE id=?", lotID)

	expectMarginCoinExecError(t, conn, `INSERT INTO t_option_asset_instruction
(tenant_id,instruction_no,biz_no,order_id,action,coin,amount,step_no,create_times,update_times)
VALUES(?,'INSTRUCTION-EMPTY','ORDER-CORRECT',?,3,'',1,1,1000,1000)`, tenantID, orderID)
	if _, err := conn.ExecCtx(ctx, `INSERT INTO t_option_asset_instruction
(tenant_id,instruction_no,biz_no,order_id,action,coin,amount,step_no,create_times,update_times)
VALUES(?,'INSTRUCTION-CORRECT','ORDER-CORRECT',?,3,'BTC',1,1,1000,1000)`, tenantID, orderID); err != nil {
		t.Fatalf("insert correctly denominated asset instruction: %v", err)
	}

	metrics, err := queryOptionMarginCoinMetricsByTenant(ctx, conn)
	if err != nil {
		t.Fatalf("query margin coin metrics: %v", err)
	}
	for _, metric := range metrics {
		if metric.TenantID == tenantID {
			t.Fatalf("healthy tenant emitted invalid margin coin metric: %+v", metric)
		}
	}
	if os.Getenv("OPTION_MARGIN_COIN_EXPECT_LEGACY") == "1" {
		expectMarginCoinExecError(t, conn,
			"UPDATE t_option_order SET margin_coin='USDT' WHERE tenant_id=919041 AND create_times=900")
		expectMarginCoinExecError(t, conn,
			"UPDATE t_option_margin_lot SET collateral_coin='USDT' WHERE tenant_id=919041 AND create_times=901")
		var found *OptionOperationsMetric
		for _, metric := range metrics {
			if metric.TenantID == 919041 && metric.Category == "margin_coin_invalid" {
				found = metric
			}
		}
		if found == nil || found.Count != 3 || found.Oldest != 900 {
			t.Fatalf("legacy invalid margin coin metric=%+v want count=3 oldest=900", found)
		}
	}
}

func expectMarginCoinExecError(t *testing.T, conn sqlx.SqlConn, query string, args ...any) {
	t.Helper()
	if _, err := conn.ExecCtx(context.Background(), query, args...); err == nil {
		t.Fatalf("expected SQL to reject invalid margin coin: %s", query)
	}
}
