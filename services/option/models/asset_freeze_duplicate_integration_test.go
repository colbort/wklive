package models

import (
	"context"
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestOptionAssetFreezeDuplicateMetricsMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	if dsn == "" {
		t.Skip("OPTION_P0_ASSET_E2E_DSN is not set")
	}
	conn := sqlx.NewMysql(dsn)
	metrics, err := queryOptionAssetFreezeDuplicateMetricsByTenant(context.Background(), conn)
	if err != nil {
		t.Fatalf("query duplicate Option freeze metrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("duplicate Option freeze tenant metrics=%d want=1: %+v", len(metrics), metrics)
	}
	got := metrics[0]
	if got.TenantID != 996032 || got.Category != "asset_freeze_duplicate" || got.Count != 1 || got.Oldest != 700 {
		t.Fatalf("duplicate Option freeze metric=%+v", got)
	}

	type migrationEvidence struct {
		IndexCount                int64 `db:"index_count"`
		UniqueIdempotencyCount    int64 `db:"unique_idempotency_count"`
		DuplicateIdempotencyCount int64 `db:"duplicate_idempotency_count"`
	}
	var evidence migrationEvidence
	if err := conn.QueryRowCtx(context.Background(), &evidence, `SELECT
  (SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics
   WHERE table_schema=DATABASE() AND table_name='t_asset_freeze'
     AND index_name='idx_asset_freeze_option_business_key') index_count,
  (SELECT COUNT(1) FROM t_asset_idempotent
   WHERE tenant_id=996030 AND biz_type='option' AND scene_type='place_order'
     AND biz_no='P0-MIGRATION-UNIQUE' AND status=2) unique_idempotency_count,
  (SELECT COUNT(1) FROM t_asset_idempotent
   WHERE tenant_id=996032 AND biz_type='option' AND scene_type='place_order'
     AND biz_no='P0-MIGRATION-DUP') duplicate_idempotency_count`); err != nil {
		t.Fatalf("query Option freeze migration evidence: %v", err)
	}
	if evidence.IndexCount != 1 || evidence.UniqueIdempotencyCount != 1 || evidence.DuplicateIdempotencyCount != 0 {
		t.Fatalf("unexpected Option freeze migration evidence: %+v", evidence)
	}
}
