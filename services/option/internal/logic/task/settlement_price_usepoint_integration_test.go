package tasklogic

import (
	"context"
	"os"
	"testing"

	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSettlementPriceUsePointFailsClosedMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_SETTLEMENT_PRICE_EVIDENCE_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_SETTLEMENT_PRICE_EVIDENCE_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	cacheConf := cache.CacheConf{{
		RedisConf: redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"},
		Weight:    100,
	}}
	var contract models.TOptionContract
	if err := conn.QueryRowCtx(ctx, &contract, "SELECT * FROM t_option_contract WHERE id=951001"); err != nil {
		t.Fatalf("find settlement contract: %v", err)
	}
	logic := NewProcessContractLifecycleLogic(ctx, &svc.ServiceContext{
		DB:                        conn,
		OptionMarketSnapshotModel: models.NewTOptionMarketSnapshotModel(conn, cacheConf),
	})

	var wrongMedian models.TOptionSettlementPrice
	if err := conn.QueryRowCtx(ctx, &wrongMedian, "SELECT * FROM t_option_settlement_price WHERE id=951201"); err != nil {
		t.Fatalf("find wrong median evidence: %v", err)
	}
	if err := logic.validateSettlementPriceForUse(&contract, &wrongMedian, true); err == nil {
		t.Fatal("settlement use point accepted a price that does not match immutable snapshot median")
	}

	var manual models.TOptionSettlementPrice
	if err := conn.QueryRowCtx(ctx, &manual, "SELECT * FROM t_option_settlement_price WHERE id=951204"); err != nil {
		t.Fatalf("find manual evidence: %v", err)
	}
	if err := logic.validateSettlementPriceForUse(&contract, &manual, true); err != nil {
		t.Fatalf("valid governed manual correction rejected: %v", err)
	}

	var automatic models.TOptionSettlementPrice
	if err := conn.QueryRowCtx(ctx, &automatic, "SELECT * FROM t_option_settlement_price WHERE id=951205"); err != nil {
		t.Fatalf("find automatic evidence: %v", err)
	}
	if err := logic.validateSettlementPriceForUse(&contract, &automatic, true); err != nil {
		t.Fatalf("valid automatic evidence rejected: %v", err)
	}
	missingSnapshot := automatic
	missingSnapshot.SourceSnapshotIds = `["LEGACY-A","LEGACY-B","MISSING"]`
	if err := logic.validateSettlementPriceForUse(&contract, &missingSnapshot, true); err == nil {
		t.Fatal("settlement use point accepted missing snapshot evidence")
	}
	wrongPrice := automatic
	wrongPrice.DeliveryPrice = decimal.NewFromInt(101)
	if err := logic.validateSettlementPriceForUse(&contract, &wrongPrice, true); err == nil {
		t.Fatal("settlement use point accepted a changed delivery price")
	}
}
