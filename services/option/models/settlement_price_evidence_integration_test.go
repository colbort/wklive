package models

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSettlementPriceEvidenceGuardsMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_SETTLEMENT_PRICE_EVIDENCE_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_SETTLEMENT_PRICE_EVIDENCE_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	tenantID := int64(920000 + time.Now().UnixNano()%1_000_000)

	contractResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_contract
(tenant_id,contract_code,underlying_coin,settle_coin,option_type,exercise_style,settlement_type,
 strike_price,expire_time,deliver_time,settlement_price_source,settlement_price_method,
 settlement_window_seconds,settlement_min_samples,seller_margin_mode,create_times,update_times)
VALUES(?,?,'BTC','USDT',1,1,1,100,1000,1100,'authoritative-market','MEDIAN',60,3,1,1,1)`,
		tenantID, fmt.Sprintf("SETTLEMENT-EVIDENCE-%d", tenantID))
	if err != nil {
		t.Fatalf("insert contract: %v", err)
	}
	contractID, err := contractResult.LastInsertId()
	if err != nil {
		t.Fatalf("contract id: %v", err)
	}
	for index, sample := range []struct {
		id    string
		time  int64
		price int64
	}{
		{"AUTO-A", 940, 90}, {"AUTO-B", 970, 100}, {"AUTO-C", 1000, 110},
		{"DUPLICATE", 950, 95}, {"DUPLICATE", 960, 105},
	} {
		if _, err := conn.ExecCtx(ctx, `INSERT INTO t_option_market_snapshot
(tenant_id,contract_id,underlying_price,snapshot_time,source_type,source_snapshot_id,create_times)
VALUES(?,?,?,?,1,?,?)`, tenantID, contractID, sample.price, sample.time, sample.id, int64(index+1)); err != nil {
			t.Fatalf("insert snapshot %s: %v", sample.id, err)
		}
	}

	expectSettlementEvidenceSQLError(t, conn, `INSERT INTO t_option_settlement_price
(tenant_id,contract_id,price_source,window_start,window_end,sample_count,calculation_method,
 delivery_price,source_snapshot_ids,version,status,change_reason,create_times,update_times)
VALUES(?,?,'authoritative-market',940,1000,3,'MEDIAN',999,'["AUTO-A","AUTO-B","AUTO-C"]',1,1,'wrong median',1,1)`, tenantID, contractID)
	expectSettlementEvidenceSQLError(t, conn, `INSERT INTO t_option_settlement_price
(tenant_id,contract_id,price_source,window_start,window_end,sample_count,calculation_method,
 delivery_price,source_snapshot_ids,version,status,change_reason,create_times,update_times)
VALUES(?,?,'authoritative-market',940,1000,3,'MEDIAN',100,'["AUTO-A","AUTO-A","AUTO-C"]',1,1,'duplicate evidence',1,1)`, tenantID, contractID)
	expectSettlementEvidenceSQLError(t, conn, `INSERT INTO t_option_settlement_price
(tenant_id,contract_id,price_source,window_start,window_end,sample_count,calculation_method,
 delivery_price,source_snapshot_ids,version,status,change_reason,create_times,update_times)
VALUES(?,?,'authoritative-market',940,1000,3,'MEDIAN',100,'["DUPLICATE","AUTO-B","AUTO-C"]',1,1,'duplicate snapshot rows',1,1)`, tenantID, contractID)
	expectSettlementEvidenceSQLError(t, conn, `INSERT INTO t_option_settlement_price
(tenant_id,contract_id,price_source,window_start,window_end,sample_count,calculation_method,
 delivery_price,source_snapshot_ids,version,status,change_reason,create_times,update_times)
VALUES(?,?,'manual-correction',940,1000,1,'MANUAL',101,'["CASE-1"]',1,1,'missing creator',1,1)`, tenantID, contractID)

	automaticResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_settlement_price
(tenant_id,contract_id,price_source,window_start,window_end,sample_count,calculation_method,
 delivery_price,source_snapshot_ids,version,status,change_reason,create_times,update_times)
VALUES(?,?,'authoritative-market',940,1000,3,'MEDIAN',100,'["AUTO-A","AUTO-B","AUTO-C"]',1,1,'system calculation',1,1)`, tenantID, contractID)
	if err != nil {
		t.Fatalf("insert valid automatic price: %v", err)
	}
	automaticID, _ := automaticResult.LastInsertId()
	if _, err := conn.ExecCtx(ctx,
		"UPDATE t_option_settlement_price SET status=2,confirmed_by=101,confirmed_at=1001,update_times=2 WHERE id=?",
		automaticID); err != nil {
		t.Fatalf("confirm valid automatic price: %v", err)
	}
	expectSettlementEvidenceSQLError(t, conn,
		"UPDATE t_option_settlement_price SET delivery_price=101 WHERE id=?", automaticID)
	expectSettlementEvidenceSQLError(t, conn,
		"DELETE FROM t_option_settlement_price WHERE id=?", automaticID)

	manualResult, err := conn.ExecCtx(ctx, `INSERT INTO t_option_settlement_price
(tenant_id,contract_id,price_source,window_start,window_end,sample_count,calculation_method,
 delivery_price,source_snapshot_ids,version,status,supersedes_id,change_reason,created_by,create_times,update_times)
VALUES(?,?,'manual-correction',940,1000,1,'MANUAL',101,'["EXTERNAL-CASE-1"]',2,1,?,'approved external correction',201,2,2)`,
		tenantID, contractID, automaticID)
	if err != nil {
		t.Fatalf("insert valid manual correction: %v", err)
	}
	manualID, _ := manualResult.LastInsertId()
	expectSettlementEvidenceSQLError(t, conn,
		"UPDATE t_option_settlement_price SET status=2,confirmed_by=201,confirmed_at=1002 WHERE id=?", manualID)
	if _, err := conn.ExecCtx(ctx,
		"UPDATE t_option_settlement_price SET status=2,confirmed_by=202,confirmed_at=1002,update_times=3 WHERE id=?",
		manualID); err != nil {
		t.Fatalf("independent manual confirmation: %v", err)
	}
	if _, err := conn.ExecCtx(ctx,
		"UPDATE t_option_settlement_price SET status=4,update_times=3 WHERE id=?", automaticID); err != nil {
		t.Fatalf("supersede automatic version: %v", err)
	}

	metrics, err := queryOptionGovernanceMetricsByTenant(ctx, conn, 1200)
	if err != nil {
		t.Fatalf("query governance metrics: %v", err)
	}
	for _, metric := range metrics {
		if metric.TenantID == tenantID && metric.Category == "settlement_price_invalid" {
			t.Fatalf("healthy automatic/manual evidence emitted invalid metric: %+v", metric)
		}
	}
	if os.Getenv("OPTION_SETTLEMENT_PRICE_EXPECT_LEGACY") == "1" {
		var found *OptionOperationsMetric
		for _, metric := range metrics {
			if metric.TenantID == 919051 && metric.Category == "settlement_price_invalid" {
				found = metric
			}
		}
		if found == nil || found.Count != 3 || found.Oldest != 700 {
			t.Fatalf("legacy invalid settlement metric=%+v want count=3 oldest=700", found)
		}
	}
}

func expectSettlementEvidenceSQLError(t *testing.T, conn sqlx.SqlConn, query string, args ...any) {
	t.Helper()
	if _, err := conn.ExecCtx(context.Background(), query, args...); err == nil {
		t.Fatalf("expected settlement evidence SQL rejection: %s", query)
	}
}
