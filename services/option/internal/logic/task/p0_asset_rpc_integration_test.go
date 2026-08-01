package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	optionconfig "wklive/services/option/internal/config"
	adminlogic "wklive/services/option/internal/logic/admin"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	zeroredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const p0AssetE2ETenantID int64 = 996031

func TestP0AssetRPCEndToEnd(t *testing.T) {
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	rpcAddr := os.Getenv("OPTION_P0_ASSET_E2E_RPC_ADDR")
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || rpcAddr == "" || redisAddr == "" {
		t.Skip("OPTION_P0_ASSET_E2E_DSN, OPTION_P0_ASSET_E2E_RPC_ADDR and OPTION_P0_ASSET_E2E_REDIS_ADDR are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping acceptance database: %v", err)
	}

	grpcConn, err := grpc.NewClient(rpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("connect Asset RPC: %v", err)
	}
	defer grpcConn.Close()
	assetClient := asset.NewAssetClient(grpcConn)
	waitForAssetRPC(t, ctx, assetClient)

	serviceCtx := newP0AssetE2EServiceContext(dsn, redisAddr, assetClient)

	t.Run("wallet scope and net liquidation equity", func(t *testing.T) {
		testP0RiskWalletAndEquity(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("freeze response loss and release replay", func(t *testing.T) {
		testP0FreezeReleaseReplay(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("physical call and put margin coin release", func(t *testing.T) {
		testP0MarginCoinRelease(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("physical Call Put ordinary order coin lifecycle", func(t *testing.T) {
		testP0PhysicalOrderCoinLifecycle(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("physical delivery Call Put failure isolation and recovery", func(t *testing.T) {
		testP1PhysicalDeliveryAssetRPC(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("confirmed settlement price to cash settlement", func(t *testing.T) {
		testP0CashSettlement(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("cash settlement debit failure barrier and recovery", func(t *testing.T) {
		testP0CashSettlementFailureRecovery(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("cash settlement stale processing recovery", func(t *testing.T) {
		testP0CashSettlementStaleProcessingRecovery(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("cash settlement insufficient balance manual recovery", func(t *testing.T) {
		testP0CashSettlementInsufficientBalanceRecovery(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("American early exercise concurrency and FIFO", func(t *testing.T) {
		testP0AmericanExerciseConcurrencyFIFO(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("American exercise races short close orders", func(t *testing.T) {
		testP0AmericanExerciseCloseOrderRace(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("expiry AUTO DNE actual assignment", func(t *testing.T) {
		testP0ExpiryAutoDNEActualAssignment(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("partial close trade accounting", func(t *testing.T) {
		testP0PartialCloseTradeAccounting(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("isolated short liquidation accounting", func(t *testing.T) {
		testP0IsolatedShortLiquidationAccounting(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("liquidation deficit backstop failure recovery", func(t *testing.T) {
		testP0LiquidationDeficitFailureRecovery(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("portfolio liquidation is sequential and preserves residual collateral", func(t *testing.T) {
		testP0PortfolioLiquidationSequential(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("full order admission to risk accounting", func(t *testing.T) {
		testP0OrderAdmissionToRiskAccounting(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("wallet restriction propagation and cross-account STP", func(t *testing.T) {
		testP0WalletRestrictionAndCrossAccountSTP(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("user cancel IOC and FOK funding lifecycle", func(t *testing.T) {
		testP0OrderCancellationAndImmediateTypes(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("market and post-only funding lifecycle", func(t *testing.T) {
		testP0MarketAndPostOnlyOrders(t, ctx, db, assetClient, serviceCtx)
	})
	t.Run("admin force cancel and funding race", func(t *testing.T) {
		testP0AdminForceCancelAndFundingRace(t, ctx, db, assetClient, serviceCtx)
	})
}

type failOnceDeductAssetClient struct {
	asset.AssetClient

	mu              sync.Mutex
	targetBizNo     string
	failAfterCommit bool
	failures        int
}

func (c *failOnceDeductAssetClient) DeductFrozenAssetByBizNo(
	ctx context.Context,
	in *asset.DeductFrozenAssetByBizNoReq,
	opts ...grpc.CallOption,
) (*asset.ChangeAssetResp, error) {
	c.mu.Lock()
	shouldFail := in.TargetBizNo == c.targetBizNo && c.failures == 0
	if shouldFail {
		c.failures++
	}
	c.mu.Unlock()
	if shouldFail {
		failurePoint := "before debit"
		if c.failAfterCommit {
			if _, err := c.AssetClient.DeductFrozenAssetByBizNo(ctx, in, opts...); err != nil {
				return nil, err
			}
			failurePoint = "after committed debit"
		}
		return nil, status.Errorf(codes.Unavailable, "P0 SET-001 injected Asset response loss %s", failurePoint)
	}
	return c.AssetClient.DeductFrozenAssetByBizNo(ctx, in, opts...)
}

func (c *failOnceDeductAssetClient) failureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failures
}

func newP0AssetE2EServiceContext(
	dsn, redisAddr string,
	assetClient asset.AssetClient,
) *svc.ServiceContext {
	conn := sqlx.NewMysql(dsn)
	redisConf := zeroredis.RedisConf{Host: redisAddr, Type: "node"}
	cacheConf := cache.CacheConf{{RedisConf: redisConf, Weight: 100}}
	config := optionconfig.Config{CacheRedis: cacheConf}
	config.Mysql.DataSource = dsn
	return &svc.ServiceContext{
		Config: config, DB: conn, Redis: zeroredis.MustNewRedis(redisConf), AssetClient: assetClient,
		OptionAssetInstructionModel:         models.NewTOptionAssetInstructionModel(conn, cacheConf),
		OptionOrderModel:                    models.NewTOptionOrderModel(conn, cacheConf),
		OptionTradeModel:                    models.NewTOptionTradeModel(conn, cacheConf),
		OptionOutboxModel:                   models.NewTOptionOutboxModel(conn, cacheConf),
		OptionInboxModel:                    models.NewTOptionInboxModel(conn, cacheConf),
		OptionReconciliationIssueModel:      models.NewTOptionReconciliationIssueModel(conn, cacheConf),
		OptionPhysicalDeliveryUnitModel:     models.NewTOptionPhysicalDeliveryUnitModel(conn, cacheConf),
		OptionContractModel:                 models.NewTOptionContractModel(conn, cacheConf),
		OptionMarketModel:                   models.NewTOptionMarketModel(conn, cacheConf),
		OptionMarketSnapshotModel:           models.NewTOptionMarketSnapshotModel(conn, cacheConf),
		OptionPositionModel:                 models.NewTOptionPositionModel(conn, cacheConf),
		OptionExerciseModel:                 models.NewTOptionExerciseModel(conn, cacheConf),
		OptionExerciseAssignmentModel:       models.NewTOptionExerciseAssignmentModel(conn, cacheConf),
		OptionExerciseInstructionModel:      models.NewTOptionExerciseInstructionModel(conn, cacheConf),
		OptionTradeCorrectionModel:          models.NewTOptionTradeCorrectionModel(conn, cacheConf),
		OptionUserTradingControlModel:       models.NewTOptionUserTradingControlModel(conn, cacheConf),
		OptionTradingControlEventModel:      models.NewTOptionTradingControlEventModel(conn, cacheConf),
		OptionTradingCalendarModel:          models.NewTOptionTradingCalendarModel(conn, cacheConf),
		OptionTradingCalendarSessionModel:   models.NewTOptionTradingCalendarSessionModel(conn, cacheConf),
		OptionTradingCalendarExceptionModel: models.NewTOptionTradingCalendarExceptionModel(conn, cacheConf),
		OptionTradingHaltModel:              models.NewTOptionTradingHaltModel(conn, cacheConf),
		OptionCorporateActionContractModel:  models.NewTOptionCorporateActionContractModel(conn, cacheConf),
		OptionSettlementPriceModel:          models.NewTOptionSettlementPriceModel(conn, cacheConf),
		OptionSettlementModel:               models.NewTOptionSettlementModel(conn, cacheConf),
		OptionSettlementBatchModel:          models.NewTOptionSettlementBatchModel(conn, cacheConf),
		OptionSettlementDetailModel:         models.NewTOptionSettlementDetailModel(conn, cacheConf),
		OptionMarginLotModel:                models.NewTOptionMarginLotModel(conn, cacheConf),
		OptionRiskAccountModel:              models.NewTOptionRiskAccountModel(conn, cacheConf),
		OptionPortfolioRiskConfigModel:      models.NewTOptionPortfolioRiskConfigModel(conn, cacheConf),
		OptionLiquidationModel:              models.NewTOptionLiquidationModel(conn, cacheConf),
		OptionInsuranceFundFlowModel:        models.NewTOptionInsuranceFundFlowModel(conn, cacheConf),
	}
}

func testP0CashSettlement(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		contractID      int64 = 996301
		missingPriceID  int64 = 996302
		longUserID      int64 = 107
		shortUserID     int64 = 108
		settlementPrice       = "120"
	)
	now := time.Now().Unix()
	expireTime := now - 10
	seedP0CashContract(t, ctx, db, contractID, "P0-E2E-CASH-CALL", expireTime, now-1)
	seedP0CashContract(t, ctx, db, missingPriceID, "P0-E2E-CASH-MISSING-PRICE", expireTime, now-1)
	creditAssetCoin(t, ctx, assetClient, longUserID, "USDT", "100", "P0-SETTLE-LONG-SEED")
	creditAssetCoin(t, ctx, assetClient, shortUserID, "USDT", "100", "P0-SETTLE-SHORT-SEED")

	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: 7001,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		ExerciseableQty: decimal.NewFromInt(1),
		Status:          int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
		CreateTimes:     now - 100, UpdateTimes: now - 100,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: 8001,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 90, UpdateTimes: now - 90,
	})
	marginLot := insertP0SettlementMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-SETTLE-SHORT-MARGIN", now,
	)
	freezeResp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "50",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: marginLot.Id, BizNo: marginLot.FreezeBizNo, Remark: "P0 settlement short margin",
	})
	assertAssetOK(t, freezeResp, err)
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "100.000000000000000000", "50.000000000000000000", "50.000000000000000000")

	seedP0SettlementPriceEvidence(t, ctx, db, contractID, expireTime, now, settlementPrice)
	logic := NewProcessContractLifecycleLogic(ctx, serviceCtx)
	if err := logic.processExpiredContracts(now); err != nil {
		t.Fatalf("create cash settlement from confirmed price: %v", err)
	}
	assertMissingPriceBlocksSettlement(t, ctx, db, missingPriceID)
	assertSettlementCreated(t, ctx, db, contractID, longPosition.Id, shortPosition.Id)

	// Step 1 consumes the short's frozen margin. Step 2 credits the long and
	// releases unused short margin; a further pass proves the completed batch is
	// idempotent and creates no extra Asset flow.
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	assertCompletedCashSettlement(t, ctx, db, contractID, longPosition.Id, shortPosition.Id, marginLot.Id)
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "120.000000000000000000", "120.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "80.000000000000000000", "80.000000000000000000", "0.000000000000000000")
	assertOptionMirrorCoin(t, ctx, db, longUserID, "USDT", "120.000000000000000000", "120.000000000000000000", "0.000000000000000000")
	assertOptionMirrorCoin(t, ctx, db, shortUserID, "USDT", "80.000000000000000000", "80.000000000000000000", "0.000000000000000000")
	assertP0SettlementAssetConservation(t, ctx, db, contractID, longUserID, shortUserID)
	if err := logic.processExpiredContracts(now); err != nil {
		t.Fatalf("replay completed/missing-price contract lifecycle: %v", err)
	}
	assertSettlementCount(t, ctx, db, contractID, 1)
	assertMissingPriceBlocksSettlement(t, ctx, db, missingPriceID)
	assertCompletedCashSettlement(t, ctx, db, contractID, longPosition.Id, shortPosition.Id, marginLot.Id)
	assertP0SettlementAssetConservation(t, ctx, db, contractID, longUserID, shortUserID)
}

func testP0CashSettlementFailureRecovery(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	scenarios := []p0CashSettlementFailureScenario{
		{
			name:       "dependency unavailable before debit",
			contractID: 996303, longUserID: 109, shortUserID: 110,
			longAccountID: 7002, shortAccountID: 8002,
			contractCode:               "P0-E2E-CASH-FAIL-RECOVER",
			freezeBizNo:                "P0-SETTLE-FAIL-SHORT-MARGIN",
			longSeedBizNo:              "P0-SETTLE-FAIL-LONG-SEED",
			shortSeedBizNo:             "P0-SETTLE-FAIL-SHORT-SEED",
			shortTotalAfterFailure:     "100.000000000000000000",
			shortAvailableAfterFailure: "50.000000000000000000",
			shortFrozenAfterFailure:    "50.000000000000000000",
		},
		{
			name:       "response lost after committed debit",
			contractID: 996304, longUserID: 111, shortUserID: 112,
			longAccountID: 7003, shortAccountID: 8003,
			contractCode:    "P0-E2E-CASH-RESPONSE-LOSS",
			freezeBizNo:     "P0-SETTLE-LOSS-SHORT-MARGIN",
			longSeedBizNo:   "P0-SETTLE-LOSS-LONG-SEED",
			shortSeedBizNo:  "P0-SETTLE-LOSS-SHORT-SEED",
			failAfterCommit: true, settlementFlowsAfterFailure: 1,
			shortTotalAfterFailure:     "80.000000000000000000",
			shortAvailableAfterFailure: "50.000000000000000000",
			shortFrozenAfterFailure:    "30.000000000000000000",
		},
	}
	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			runP0CashSettlementFailureRecovery(t, ctx, db, assetClient, serviceCtx, scenario)
		})
	}
}

type p0CashSettlementFailureScenario struct {
	name                        string
	contractID                  int64
	longUserID                  int64
	shortUserID                 int64
	longAccountID               int64
	shortAccountID              int64
	contractCode                string
	freezeBizNo                 string
	longSeedBizNo               string
	shortSeedBizNo              string
	failAfterCommit             bool
	settlementFlowsAfterFailure int64
	shortTotalAfterFailure      string
	shortAvailableAfterFailure  string
	shortFrozenAfterFailure     string
}

func runP0CashSettlementFailureRecovery(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	scenario p0CashSettlementFailureScenario,
) {
	t.Helper()
	now := time.Now().Unix()
	expireTime := now - 10
	seedP0CashContract(t, ctx, db, scenario.contractID, scenario.contractCode, expireTime, now-1)
	creditAssetCoin(t, ctx, assetClient, scenario.longUserID, "USDT", "100", scenario.longSeedBizNo)
	creditAssetCoin(t, ctx, assetClient, scenario.shortUserID, "USDT", "100", scenario.shortSeedBizNo)

	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: scenario.longUserID, AccountId: scenario.longAccountID,
		ContractId: scenario.contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		ExerciseableQty: decimal.NewFromInt(1),
		Status:          int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
		CreateTimes:     now - 100, UpdateTimes: now - 100,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: scenario.shortUserID, AccountId: scenario.shortAccountID,
		ContractId: scenario.contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 90, UpdateTimes: now - 90,
	})
	marginLot := insertP0SettlementMarginLot(
		t, ctx, serviceCtx, shortPosition, scenario.freezeBizNo, now,
	)
	freezeResp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: scenario.shortUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "50",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: marginLot.Id, BizNo: marginLot.FreezeBizNo, Remark: "P0 SET-001 short margin",
	})
	assertAssetOK(t, freezeResp, err)
	seedP0SettlementPriceEvidence(t, ctx, db, scenario.contractID, expireTime, now, "120")

	faultClient := &failOnceDeductAssetClient{
		AssetClient:     assetClient,
		targetBizNo:     marginLot.FreezeBizNo,
		failAfterCommit: scenario.failAfterCommit,
	}
	originalClient := serviceCtx.AssetClient
	serviceCtx.AssetClient = faultClient
	defer func() { serviceCtx.AssetClient = originalClient }()

	lifecycle := NewProcessContractLifecycleLogic(ctx, serviceCtx)
	if err := lifecycle.processExpiredContracts(now); err != nil {
		t.Fatalf("create SET-001 cash settlement: %v", err)
	}
	assertSettlementCreated(t, ctx, db, scenario.contractID, longPosition.Id, shortPosition.Id)

	processAssetInstructions(t, ctx, serviceCtx)
	if faultClient.failureCount() != 1 {
		t.Fatalf("injected debit failures=%d want=1", faultClient.failureCount())
	}
	settlementNo, deductInstructionNo := assertP0SettlementDebitFailureBarrier(
		t, ctx, db, scenario.contractID, scenario.longUserID, scenario.shortUserID,
		scenario.settlementFlowsAfterFailure,
	)
	assertWalletCoinAmounts(t, ctx, db, scenario.longUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, scenario.shortUserID, "USDT", scenario.shortTotalAfterFailure, scenario.shortAvailableAfterFailure, scenario.shortFrozenAfterFailure)

	// The scheduler would naturally retry after backoff. Move only its retry
	// timestamp forward so this acceptance remains fast; the economic identity
	// and original instruction number are unchanged.
	if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET next_retry_at=0 WHERE tenant_id=? AND instruction_no=? AND status=?`,
		p0AssetE2ETenantID, deductInstructionNo,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
	); err != nil {
		t.Fatalf("make failed debit immediately retryable: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0SettlementDebitRecoveredBeforeCredit(
		t, ctx, db, settlementNo, deductInstructionNo, scenario.longUserID, scenario.shortUserID, 1,
	)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	assertCompletedCashSettlement(t, ctx, db, scenario.contractID, longPosition.Id, shortPosition.Id, marginLot.Id)
	assertWalletCoinAmounts(t, ctx, db, scenario.longUserID, "USDT", "120.000000000000000000", "120.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, scenario.shortUserID, "USDT", "80.000000000000000000", "80.000000000000000000", "0.000000000000000000")
	assertP0SettlementAssetConservation(t, ctx, db, scenario.contractID, scenario.longUserID, scenario.shortUserID)
	assertP0RecoveredDebitIdentity(t, ctx, db, settlementNo, deductInstructionNo, 1)
}

func testP0CashSettlementStaleProcessingRecovery(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		contractID  int64 = 996305
		longUserID  int64 = 113
		shortUserID int64 = 114
	)
	now := time.Now().Unix()
	expireTime := now - 10
	seedP0CashContract(t, ctx, db, contractID, "P0-E2E-CASH-STALE-PROCESSING", expireTime, now-1)
	creditAssetCoin(t, ctx, assetClient, longUserID, "USDT", "100", "P0-SETTLE-STALE-LONG-SEED")
	creditAssetCoin(t, ctx, assetClient, shortUserID, "USDT", "100", "P0-SETTLE-STALE-SHORT-SEED")

	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: 7004,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		ExerciseableQty: decimal.NewFromInt(1),
		Status:          int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
		CreateTimes:     now - 100, UpdateTimes: now - 100,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: 8004,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 90, UpdateTimes: now - 90,
	})
	marginLot := insertP0SettlementMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-SETTLE-STALE-SHORT-MARGIN", now,
	)
	freezeResp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "50",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: marginLot.Id, BizNo: marginLot.FreezeBizNo, Remark: "P0 OPS-006 short margin",
	})
	assertAssetOK(t, freezeResp, err)
	seedP0SettlementPriceEvidence(t, ctx, db, contractID, expireTime, now, "120")

	lifecycle := NewProcessContractLifecycleLogic(ctx, serviceCtx)
	if err := lifecycle.processExpiredContracts(now); err != nil {
		t.Fatalf("create OPS-006 cash settlement: %v", err)
	}
	assertSettlementCreated(t, ctx, db, contractID, longPosition.Id, shortPosition.Id)

	settlementNo, deductInstruction := findP0SettlementDeductInstruction(t, ctx, db, serviceCtx, contractID)
	if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET status=?,update_times=? WHERE id=? AND status=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING), now-120,
		deductInstruction.Id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
	); err != nil {
		t.Fatalf("simulate claimed stale instruction: %v", err)
	}
	deductResp, err := assetClient.DeductFrozenAssetByBizNo(ctx, &asset.DeductFrozenAssetByBizNoReq{
		TenantId: p0AssetE2ETenantID, TargetBizType: asset.BizType_BIZ_TYPE_OPTION,
		TargetBizNo: deductInstruction.TargetBizNo, Amount: deductInstruction.Amount.String(),
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
		BizId: deductInstruction.Id, BizNo: deductInstruction.InstructionNo,
		Remark: "option asset instruction deduct frozen",
	})
	assertAssetOK(t, deductResp, err)
	assertP0StaleProcessingBarrier(t, ctx, db, settlementNo, deductInstruction.InstructionNo)
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "80.000000000000000000", "50.000000000000000000", "30.000000000000000000")

	processAssetInstructions(t, ctx, serviceCtx)
	assertP0SettlementDebitRecoveredBeforeCredit(
		t, ctx, db, settlementNo, deductInstruction.InstructionNo, longUserID, shortUserID, 0,
	)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	assertCompletedCashSettlement(t, ctx, db, contractID, longPosition.Id, shortPosition.Id, marginLot.Id)
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "120.000000000000000000", "120.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "80.000000000000000000", "80.000000000000000000", "0.000000000000000000")
	assertP0SettlementAssetConservation(t, ctx, db, contractID, longUserID, shortUserID)
	assertP0RecoveredDebitIdentity(t, ctx, db, settlementNo, deductInstruction.InstructionNo, 0)
}

func testP0CashSettlementInsufficientBalanceRecovery(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		contractID  int64 = 996306
		longUserID  int64 = 115
		shortUserID int64 = 116
	)
	now := time.Now().Unix()
	expireTime := now - 10
	seedP0CashContract(t, ctx, db, contractID, "P0-E2E-CASH-INSUFFICIENT", expireTime, now-1)
	creditAssetCoin(t, ctx, assetClient, longUserID, "USDT", "100", "P0-SETTLE-INSUFFICIENT-LONG-SEED")
	creditAssetCoin(t, ctx, assetClient, shortUserID, "USDT", "100", "P0-SETTLE-INSUFFICIENT-SHORT-SEED")

	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: 7005,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		ExerciseableQty: decimal.NewFromInt(1),
		Status:          int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
		CreateTimes:     now - 100, UpdateTimes: now - 100,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: 8005,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 90, UpdateTimes: now - 90,
	})
	marginLot := insertP0SettlementMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-SETTLE-INSUFFICIENT-SHORT-MARGIN", now,
	)
	freezeResp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "50",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: marginLot.Id, BizNo: marginLot.FreezeBizNo, Remark: "P0 insufficient balance short margin",
	})
	assertAssetOK(t, freezeResp, err)
	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contractID, expireTime, now,
		"P0-SETTLE-INSUFFICIENT", []string{"219", "220", "221"}, "220",
	)

	lifecycle := NewProcessContractLifecycleLogic(ctx, serviceCtx)
	if err := lifecycle.processExpiredContracts(now); err != nil {
		t.Fatalf("create insufficient-balance cash settlement: %v", err)
	}
	assertSettlementCreatedWithAmount(
		t, ctx, db, contractID, longPosition.Id, shortPosition.Id, "120.0000000000000000",
	)
	settlementNo, availableDebit := findP0SettlementInstructionByAction(
		t, ctx, db, serviceCtx, contractID,
		option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE,
	)

	processAssetInstructions(t, ctx, serviceCtx)
	assertP0AssetInstructionRetryState(t, ctx, db, availableDebit.Id, 1, false)
	for retry := int64(2); retry <= 20; retry++ {
		if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
			SET next_retry_at=0 WHERE id=? AND status=?`, availableDebit.Id,
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		); err != nil {
			t.Fatalf("make insufficient debit retry %d runnable: %v", retry, err)
		}
		processAssetInstructions(t, ctx, serviceCtx)
		assertP0AssetInstructionRetryState(t, ctx, db, availableDebit.Id, retry, retry == 20)
	}
	assertP0InsufficientBalanceBarrier(t, ctx, db, settlementNo, availableDebit.InstructionNo)
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "50.000000000000000000", "50.000000000000000000", "0.000000000000000000")

	creditAssetCoin(t, ctx, assetClient, shortUserID, "USDT", "20", "P0-SETTLE-INSUFFICIENT-TOPUP-20")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "70.000000000000000000", "70.000000000000000000", "0.000000000000000000")
	missingReasonCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUserType, fmt.Sprint(utils.SysUserTypeSystemAdmin),
		utils.CtxKeyUid, "9002",
	))
	missingReasonResp, err := adminlogic.NewRetryAssetInstructionLogic(
		missingReasonCtx, serviceCtx,
	).RetryAssetInstruction(&option.RetryAssetInstructionReq{
		TenantId: p0AssetE2ETenantID, InstructionId: availableDebit.Id,
	})
	assertP0ManualRetryRejected(t, ctx, db, availableDebit.Id, missingReasonResp, err)
	tooLongReasonResp, err := adminlogic.NewRetryAssetInstructionLogic(
		missingReasonCtx, serviceCtx,
	).RetryAssetInstruction(&option.RetryAssetInstructionReq{
		TenantId: p0AssetE2ETenantID, InstructionId: availableDebit.Id,
		Reason: strings.Repeat("补", 65),
	})
	assertP0ManualRetryRejected(t, ctx, db, availableDebit.Id, tooLongReasonResp, err)

	missingOperatorCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUserType, fmt.Sprint(utils.SysUserTypeSystemAdmin),
	))
	missingOperatorResp, err := adminlogic.NewRetryAssetInstructionLogic(
		missingOperatorCtx, serviceCtx,
	).RetryAssetInstruction(&option.RetryAssetInstructionReq{
		TenantId: p0AssetE2ETenantID, InstructionId: availableDebit.Id,
		Reason: "SETTLEMENT_BALANCE_TOPUP_VERIFIED",
	})
	assertP0ManualRetryRejected(t, ctx, db, availableDebit.Id, missingOperatorResp, err)

	adminCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUserType, fmt.Sprint(utils.SysUserTypeSystemAdmin),
		utils.CtxKeyUid, "9002",
	))
	retryResp, err := adminlogic.NewRetryAssetInstructionLogic(
		adminCtx, serviceCtx,
	).RetryAssetInstruction(&option.RetryAssetInstructionReq{
		TenantId: p0AssetE2ETenantID, InstructionId: availableDebit.Id,
		Reason: "SETTLEMENT_BALANCE_TOPUP_VERIFIED",
	})
	if err != nil {
		t.Fatalf("manual retry insufficient debit: %v", err)
	}
	if retryResp == nil || retryResp.Base == nil || retryResp.Base.Code != 200 {
		t.Fatalf("manual retry response: %+v", retryResp)
	}
	assertP0AssetInstructionRetryState(t, ctx, db, availableDebit.Id, 0, false)

	processAssetInstructions(t, ctx, serviceCtx)
	assertP0AvailableDebitRecoveredBeforeCredit(
		t, ctx, db, settlementNo, availableDebit.InstructionNo, longUserID, shortUserID,
	)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	assertCompletedCashSettlementWithPayoff(
		t, ctx, db, contractID, longPosition.Id, shortPosition.Id, marginLot.Id,
		"240.0000000000000000",
	)
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "220.000000000000000000", "220.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertP0SettlementAssetConservationWithTotal(
		t, ctx, db, contractID, longUserID, shortUserID, "220.000000000000000000",
	)
	assertP0InsufficientTopupEvidence(
		t, ctx, db, contractID, shortUserID, settlementNo, availableDebit.Id, availableDebit.InstructionNo,
	)
}

func seedP0CashContract(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	code string,
	expireTime, deliverTime int64,
) {
	t.Helper()
	_, err := db.ExecContext(ctx, `
		INSERT INTO t_option_contract (
			id,tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
			option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
			max_order_qty,price_tick,qty_step,multiplier,list_time,exercise_cutoff_time,expire_time,
			deliver_time,max_user_long_qty,max_user_short_qty,max_open_interest,order_price_band_ratio,
			circuit_breaker_ratio,greeks_max_age_seconds,seller_margin_mode,initial_margin_rate,
			maintenance_margin_rate,min_margin_rate,status,is_deleted,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		contractID, p0AssetE2ETenantID, code, "BTCUSDT", "BTC", "USDT", "USDT",
		int64(option.OptionType_OPTION_TYPE_CALL), int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		int64(option.SettlementType_SETTLEMENT_TYPE_CASH), "100", "1", "1", "1000", "0.1", "1", "1",
		expireTime-3600, expireTime, expireTime, deliverTime, "10000", "10000", "10000", "0.2", "0.5", 60,
		int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED), "0.5", "0.2", "0.1",
		int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED), int64(common.YesNo_YES_NO_NO), expireTime-3600, expireTime,
	)
	if err != nil {
		t.Fatalf("seed cash settlement contract %s: %v", code, err)
	}
}

func insertP0SettlementPosition(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	position *models.TOptionPosition,
) *models.TOptionPosition {
	t.Helper()
	result, err := serviceCtx.OptionPositionModel.Insert(ctx, position)
	if err != nil {
		t.Fatalf("insert settlement position: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	stored, err := serviceCtx.OptionPositionModel.FindOne(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func insertP0SettlementMarginLot(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	position *models.TOptionPosition,
	freezeBizNo string,
	now int64,
) *models.TOptionMarginLot {
	t.Helper()
	lot := &models.TOptionMarginLot{
		TenantId: position.TenantId, UserId: position.UserId, AccountId: position.AccountId,
		ContractId: position.ContractId, PositionId: position.Id,
		OriginContractId: position.ContractId, OriginPositionId: position.Id,
		TradeId: -position.Id, FreezeBizNo: freezeBizNo, CollateralCoin: "USDT",
		Quantity: decimal.NewFromInt(1), RemainingQuantity: decimal.NewFromInt(1),
		InitialMargin: decimal.NewFromInt(50), RemainingMargin: decimal.NewFromInt(50),
		Status:      int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		CreateTimes: now - 80, UpdateTimes: now - 80,
	}
	result, err := serviceCtx.OptionMarginLotModel.Insert(ctx, lot)
	if err != nil {
		t.Fatalf("insert settlement margin lot: %v", err)
	}
	lot.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return lot
}

func seedP0SettlementPriceEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, expireTime, now int64,
	deliveryPrice string,
) {
	t.Helper()
	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contractID, expireTime, now,
		"P0-SETTLE", []string{"119", "120", "121"}, deliveryPrice,
	)
}

func seedP0SettlementPriceEvidenceWithSamples(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, expireTime, now int64,
	sourcePrefix string,
	prices []string,
	deliveryPrice string,
) {
	t.Helper()
	if len(prices) != 3 {
		t.Fatalf("settlement evidence prices=%d want=3", len(prices))
	}
	ids := []string{sourcePrefix + "-S1", sourcePrefix + "-S2", sourcePrefix + "-S3"}
	for i, price := range prices {
		if _, err := db.ExecContext(ctx, `INSERT INTO t_option_market_snapshot
			(tenant_id,contract_id,underlying_price,snapshot_time,source_type,source_snapshot_id,create_times)
			VALUES (?,?,?,?,1,?,?)`,
			p0AssetE2ETenantID, contractID, price, expireTime-30+int64(i*10), ids[i], now,
		); err != nil {
			t.Fatalf("insert settlement snapshot %s: %v", ids[i], err)
		}
	}
	evidence := fmt.Sprintf(`[%q,%q,%q]`, ids[0], ids[1], ids[2])
	insertPrice := func(price string) error {
		_, err := db.ExecContext(ctx, `INSERT INTO t_option_settlement_price
			(tenant_id,contract_id,price_source,window_start,window_end,sample_count,
			 calculation_method,delivery_price,source_snapshot_ids,version,status,
			 change_reason,created_by,confirmed_by,confirmed_at,create_times,update_times)
			VALUES (?,?,?,?,?,3,'MEDIAN',?,?,1,2,'P0 confirmed automatic settlement',0,9001,?,?,?)`,
			p0AssetE2ETenantID, contractID, "authoritative-market", expireTime-60, expireTime,
			price, evidence, now, now, now,
		)
		return err
	}
	wrongPrice := decimal.RequireFromString(deliveryPrice).Add(decimal.NewFromInt(1)).String()
	if err := insertPrice(wrongPrice); err == nil {
		t.Fatal("database accepted a confirmed settlement price that is not the snapshot median")
	}
	if err := insertPrice(deliveryPrice); err != nil {
		t.Fatalf("insert confirmed settlement price: %v", err)
	}
}

func assertMissingPriceBlocksSettlement(t *testing.T, ctx context.Context, db *sql.DB, contractID int64) {
	t.Helper()
	assertSettlementCount(t, ctx, db, contractID, 0)
	var status int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_contract WHERE id=?`, contractID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED) {
		t.Fatalf("missing-price contract status=%d want EXPIRED", status)
	}
}

func assertSettlementCount(t *testing.T, ctx context.Context, db *sql.DB, contractID, want int64) {
	t.Helper()
	var count int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_settlement
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("contract %d settlement count=%d want=%d", contractID, count, want)
	}
}

func assertSettlementCreated(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longPositionID, shortPositionID int64,
) {
	assertSettlementCreatedWithAmount(
		t, ctx, db, contractID, longPositionID, shortPositionID, "20.0000000000000000",
	)
}

func assertSettlementCreatedWithAmount(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longPositionID, shortPositionID int64,
	wantAmount string,
) {
	t.Helper()
	var settlementStatus, batchStatus, instructionCount, detailCount, settledPositions int64
	var totalCredit, totalDebit string
	if err := db.QueryRowContext(ctx, `SELECT s.status,b.status,b.instruction_count,
		CAST(b.total_credit AS CHAR),CAST(b.total_debit AS CHAR),
		(SELECT COUNT(*) FROM t_option_settlement_detail d WHERE d.batch_id=b.id),
		(SELECT COUNT(*) FROM t_option_position p WHERE p.id IN (?,?) AND p.status=? AND p.position_qty=0)
		FROM t_option_settlement s
		JOIN t_option_settlement_batch b ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
		WHERE s.tenant_id=? AND s.contract_id=?`,
		longPositionID, shortPositionID, int64(option.PositionStatus_POSITION_STATUS_SETTLED),
		p0AssetE2ETenantID, contractID,
	).Scan(&settlementStatus, &batchStatus, &instructionCount, &totalCredit, &totalDebit, &detailCount, &settledPositions); err != nil {
		t.Fatal(err)
	}
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_PROCESSING) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_INSTRUCTIONS_CREATED) ||
		instructionCount != 3 || totalCredit != wantAmount || totalDebit != wantAmount ||
		detailCount != 2 || settledPositions != 2 {
		t.Fatalf("created settlement evidence=%d/%d/%d/%s/%s/%d/%d",
			settlementStatus, batchStatus, instructionCount, totalCredit, totalDebit, detailCount, settledPositions)
	}
}

func assertP0SettlementDebitFailureBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longUserID, shortUserID int64,
	wantSettlementFlows int64,
) (string, string) {
	t.Helper()
	var settlementNo, instructionNo string
	var instructionStatus, retryCount, batchStatus, batchSuccess, stepTwoSuccess, settlementFlows int64
	if err := db.QueryRowContext(ctx, `SELECT s.settlement_no,instruction.instruction_no,
		instruction.status,instruction.retry_count,batch.status,batch.success_count,
		(SELECT COUNT(*) FROM t_option_asset_instruction later
		 WHERE later.tenant_id=s.tenant_id AND later.biz_no=s.settlement_no
		   AND later.step_no=2 AND later.status=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction linked
		   ON linked.tenant_id=flow.tenant_id AND linked.instruction_no=flow.biz_no
		 WHERE linked.tenant_id=s.tenant_id AND linked.biz_no=s.settlement_no)
		FROM t_option_settlement s
		JOIN t_option_settlement_batch batch
		  ON batch.tenant_id=s.tenant_id AND batch.batch_no=s.settlement_no
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=s.tenant_id AND instruction.biz_no=s.settlement_no
		 AND instruction.action=?
		WHERE s.tenant_id=? AND s.contract_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
		p0AssetE2ETenantID, contractID,
	).Scan(&settlementNo, &instructionNo, &instructionStatus, &retryCount,
		&batchStatus, &batchSuccess, &stepTwoSuccess, &settlementFlows); err != nil {
		t.Fatal(err)
	}
	if instructionStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED) ||
		retryCount != 1 ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_INSTRUCTIONS_CREATED) ||
		batchSuccess != 0 || stepTwoSuccess != 0 || settlementFlows != wantSettlementFlows {
		t.Fatalf("failed debit barrier evidence status/retry/batch/success/step2/flows=%d/%d/%d/%d/%d/%d",
			instructionStatus, retryCount, batchStatus, batchSuccess, stepTwoSuccess, settlementFlows)
	}
	var creditedUsers int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND biz_type='option' AND scene_type='trade_match'
		  AND user_id IN (?,?) AND biz_no IN (
		    SELECT instruction_no FROM t_option_asset_instruction
		    WHERE tenant_id=? AND biz_no=? AND step_no=2
		  )`, p0AssetE2ETenantID, longUserID, shortUserID,
		p0AssetE2ETenantID, settlementNo,
	).Scan(&creditedUsers); err != nil {
		t.Fatal(err)
	}
	if creditedUsers != 0 {
		t.Fatalf("step-2 Asset flows before debit recovery=%d want=0", creditedUsers)
	}
	return settlementNo, instructionNo
}

func findP0SettlementDeductInstruction(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID int64,
) (string, *models.TOptionAssetInstruction) {
	return findP0SettlementInstructionByAction(
		t, ctx, db, serviceCtx, contractID,
		option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN,
	)
}

func findP0SettlementInstructionByAction(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID int64,
	action option.AssetInstructionAction,
) (string, *models.TOptionAssetInstruction) {
	t.Helper()
	var settlementNo string
	if err := db.QueryRowContext(ctx, `SELECT settlement_no FROM t_option_settlement
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID,
	).Scan(&settlementNo); err != nil {
		t.Fatal(err)
	}
	instructions, err := serviceCtx.OptionAssetInstructionModel.FindByBizNo(
		ctx, p0AssetE2ETenantID, settlementNo,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, instruction := range instructions {
		if instruction.Action == int64(action) {
			return settlementNo, instruction
		}
	}
	t.Fatalf("settlement %s has no instruction action %s", settlementNo, action)
	return "", nil
}

func assertP0StaleProcessingBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	settlementNo, instructionNo string,
) {
	t.Helper()
	var statusValue, retryCount, debitFlows, stepTwoSuccess, stepTwoFlows int64
	if err := db.QueryRowContext(ctx, `SELECT instruction.status,instruction.retry_count,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no),
		(SELECT COUNT(*) FROM t_option_asset_instruction later
		 WHERE later.tenant_id=instruction.tenant_id AND later.biz_no=instruction.biz_no
		   AND later.step_no=2 AND later.status=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction later
		   ON later.tenant_id=flow.tenant_id AND later.instruction_no=flow.biz_no
		 WHERE later.tenant_id=instruction.tenant_id AND later.biz_no=instruction.biz_no
		   AND later.step_no=2)
		FROM t_option_asset_instruction instruction
		WHERE instruction.tenant_id=? AND instruction.instruction_no=? AND instruction.biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, instructionNo, settlementNo,
	).Scan(&statusValue, &retryCount, &debitFlows, &stepTwoSuccess, &stepTwoFlows); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING) ||
		retryCount != 0 || debitFlows != 1 || stepTwoSuccess != 0 || stepTwoFlows != 0 {
		t.Fatalf("stale processing barrier status/retry/debitFlows/step2Success/step2Flows=%d/%d/%d/%d/%d",
			statusValue, retryCount, debitFlows, stepTwoSuccess, stepTwoFlows)
	}
}

func assertP0SettlementDebitRecoveredBeforeCredit(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	settlementNo, instructionNo string,
	longUserID, shortUserID int64,
	wantRetryCount int64,
) {
	t.Helper()
	var statusValue, retryCount, reconciliationStatus, flowCount, stepTwoSuccess int64
	if err := db.QueryRowContext(ctx, `SELECT instruction.status,instruction.retry_count,
		instruction.reconciliation_status,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no),
		(SELECT COUNT(*) FROM t_option_asset_instruction later
		 WHERE later.tenant_id=instruction.tenant_id AND later.biz_no=instruction.biz_no
		   AND later.step_no=2 AND later.status=?)
		FROM t_option_asset_instruction instruction
		WHERE instruction.tenant_id=? AND instruction.instruction_no=? AND instruction.biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, instructionNo, settlementNo,
	).Scan(&statusValue, &retryCount, &reconciliationStatus, &flowCount, &stepTwoSuccess); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) ||
		retryCount != wantRetryCount ||
		reconciliationStatus != int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED) ||
		flowCount != 1 || stepTwoSuccess != 0 {
		t.Fatalf("recovered debit-before-credit evidence status/retry/reconcile/flows/step2=%d/%d/%d/%d/%d",
			statusValue, retryCount, reconciliationStatus, flowCount, stepTwoSuccess)
	}
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "80.000000000000000000", "50.000000000000000000", "30.000000000000000000")
}

func assertP0RecoveredDebitIdentity(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	settlementNo, instructionNo string,
	wantRetryCount int64,
) {
	t.Helper()
	var instructionCount, successCount, flowCount, recoveredDebitFlows, retryCount int64
	if err := db.QueryRowContext(ctx, `SELECT
		COUNT(*),SUM(instruction.status=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction linked
		   ON linked.tenant_id=flow.tenant_id AND linked.instruction_no=flow.biz_no
		 WHERE linked.tenant_id=? AND linked.biz_no=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=? AND flow.biz_type='option'
		   AND flow.scene_type='trade_match' AND flow.biz_no=?),
		MAX(CASE WHEN instruction.instruction_no=? THEN instruction.retry_count ELSE 0 END)
		FROM t_option_asset_instruction instruction
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, settlementNo,
		p0AssetE2ETenantID, instructionNo, instructionNo,
		p0AssetE2ETenantID, settlementNo,
	).Scan(&instructionCount, &successCount, &flowCount, &recoveredDebitFlows, &retryCount); err != nil {
		t.Fatal(err)
	}
	if instructionCount != 3 || successCount != 3 || flowCount != 3 ||
		recoveredDebitFlows != 1 || retryCount != wantRetryCount {
		t.Fatalf("recovered debit identity instructions/success/flows/debitFlows/retry=%d/%d/%d/%d/%d",
			instructionCount, successCount, flowCount, recoveredDebitFlows, retryCount)
	}
}

func assertP0AssetInstructionRetryState(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	instructionID, wantRetryCount int64,
	wantManual bool,
) {
	t.Helper()
	wantStatus := int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED)
	if wantRetryCount == 0 {
		wantStatus = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING)
	}
	if wantManual {
		wantStatus = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW)
	}
	var statusValue, retryCount, nextRetryAt int64
	var lastError string
	if err := db.QueryRowContext(ctx, `SELECT status,retry_count,next_retry_at,last_error_msg
		FROM t_option_asset_instruction WHERE id=?`, instructionID,
	).Scan(&statusValue, &retryCount, &nextRetryAt, &lastError); err != nil {
		t.Fatal(err)
	}
	if statusValue != wantStatus || retryCount != wantRetryCount ||
		(wantManual && nextRetryAt != 0) ||
		(wantRetryCount > 0 && lastError == "") ||
		(wantRetryCount == 0 && lastError != "") {
		t.Fatalf("instruction retry state status/retry/next/error=%d/%d/%d/%q wantStatus/retry=%d/%d",
			statusValue, retryCount, nextRetryAt, lastError, wantStatus, wantRetryCount)
	}
}

func assertP0InsufficientBalanceBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	settlementNo, debitInstructionNo string,
) {
	t.Helper()
	var instructionCount, successCount, manualCount, debitFlows, stepTwoSuccess, stepTwoFlows int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(instruction.status=?),
		SUM(instruction.status=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=? AND flow.biz_no=?),
		SUM(instruction.step_no=2 AND instruction.status=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction later
		   ON later.tenant_id=flow.tenant_id AND later.instruction_no=flow.biz_no
		 WHERE later.tenant_id=? AND later.biz_no=? AND later.step_no=2)
		FROM t_option_asset_instruction instruction
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW),
		p0AssetE2ETenantID, debitInstructionNo,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, settlementNo,
		p0AssetE2ETenantID, settlementNo,
	).Scan(&instructionCount, &successCount, &manualCount, &debitFlows, &stepTwoSuccess, &stepTwoFlows); err != nil {
		t.Fatal(err)
	}
	if instructionCount != 3 || successCount != 1 || manualCount != 1 ||
		debitFlows != 0 || stepTwoSuccess != 0 || stepTwoFlows != 0 {
		t.Fatalf("insufficient balance barrier instructions/success/manual/debitFlows/step2Success/step2Flows=%d/%d/%d/%d/%d/%d",
			instructionCount, successCount, manualCount, debitFlows, stepTwoSuccess, stepTwoFlows)
	}
}

func assertP0AvailableDebitRecoveredBeforeCredit(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	settlementNo, debitInstructionNo string,
	longUserID, shortUserID int64,
) {
	t.Helper()
	var statusValue, retryCount, reconciliationStatus, debitFlows, stepTwoSuccess, settlementFlows int64
	if err := db.QueryRowContext(ctx, `SELECT instruction.status,instruction.retry_count,
		instruction.reconciliation_status,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no),
		(SELECT COUNT(*) FROM t_option_asset_instruction later
		 WHERE later.tenant_id=instruction.tenant_id AND later.biz_no=instruction.biz_no
		   AND later.step_no=2 AND later.status=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction linked
		   ON linked.tenant_id=flow.tenant_id AND linked.instruction_no=flow.biz_no
		 WHERE linked.tenant_id=instruction.tenant_id AND linked.biz_no=instruction.biz_no)
		FROM t_option_asset_instruction instruction
		WHERE instruction.tenant_id=? AND instruction.instruction_no=? AND instruction.biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, debitInstructionNo, settlementNo,
	).Scan(&statusValue, &retryCount, &reconciliationStatus, &debitFlows,
		&stepTwoSuccess, &settlementFlows); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) ||
		retryCount != 0 ||
		reconciliationStatus != int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED) ||
		debitFlows != 1 || stepTwoSuccess != 0 || settlementFlows != 2 {
		t.Fatalf("available debit recovery status/retry/reconcile/debitFlows/step2/settlementFlows=%d/%d/%d/%d/%d/%d",
			statusValue, retryCount, reconciliationStatus, debitFlows, stepTwoSuccess, settlementFlows)
	}
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
}

func assertP0ManualRetryRejected(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	instructionID int64,
	resp *option.CommonResp,
	err error,
) {
	t.Helper()
	if err != nil {
		t.Fatalf("rejected manual retry returned transport error: %v", err)
	}
	if resp == nil || resp.Base == nil || resp.Base.Code == 200 {
		t.Fatalf("manual retry without required audit input must be rejected: %+v", resp)
	}
	var statusValue, retryCount, eventCount int64
	if err := db.QueryRowContext(ctx, `SELECT status,retry_count,
		(SELECT COUNT(*) FROM t_option_trading_control_event
		 WHERE tenant_id=? AND event_type='ASSET_INSTRUCTION_MANUAL_RETRY')
		FROM t_option_asset_instruction WHERE id=?`, p0AssetE2ETenantID, instructionID,
	).Scan(&statusValue, &retryCount, &eventCount); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW) ||
		retryCount != 20 || eventCount != 0 {
		t.Fatalf("rejected manual retry changed instruction/audit state status/retry/events=%d/%d/%d",
			statusValue, retryCount, eventCount)
	}
}

func assertP0InsufficientTopupEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, shortUserID int64,
	settlementNo string,
	debitInstructionID int64,
	debitInstructionNo string,
) {
	t.Helper()
	var topupFlows, debitFlows, settlementFlows, successCount, reconciledCount int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=? AND flow.user_id=? AND flow.biz_no='P0-SETTLE-INSUFFICIENT-TOPUP-20'),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=? AND flow.user_id=? AND flow.biz_no=?),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction linked
		   ON linked.tenant_id=flow.tenant_id AND linked.instruction_no=flow.biz_no
		 WHERE linked.tenant_id=? AND linked.biz_no=?),
		SUM(instruction.status=?),SUM(instruction.reconciliation_status=?)
		FROM t_option_asset_instruction instruction
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		p0AssetE2ETenantID, shortUserID,
		p0AssetE2ETenantID, shortUserID, debitInstructionNo,
		p0AssetE2ETenantID, settlementNo,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, settlementNo,
	).Scan(&topupFlows, &debitFlows, &settlementFlows, &successCount, &reconciledCount); err != nil {
		t.Fatal(err)
	}
	if topupFlows != 1 || debitFlows != 1 || settlementFlows != 3 ||
		successCount != 3 || reconciledCount != 3 {
		t.Fatalf("insufficient topup evidence topup/debit/settlement/success/reconciled=%d/%d/%d/%d/%d",
			topupFlows, debitFlows, settlementFlows, successCount, reconciledCount)
	}

	var eventCount, eventID, operatorID, eventUserID, eventContractID int64
	var eventType, reason, detail string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(MAX(id),0),
		COALESCE(MAX(event_type),''),COALESCE(MAX(reason),''),COALESCE(MAX(operator_id),0),
		COALESCE(MAX(user_id),0),COALESCE(MAX(contract_id),0),COALESCE(MAX(detail),'')
		FROM t_option_trading_control_event
		WHERE tenant_id=? AND event_type='ASSET_INSTRUCTION_MANUAL_RETRY'`,
		p0AssetE2ETenantID,
	).Scan(&eventCount, &eventID, &eventType, &reason, &operatorID,
		&eventUserID, &eventContractID, &detail); err != nil {
		t.Fatal(err)
	}
	wantDetailParts := []string{
		fmt.Sprintf("instructionId=%d", debitInstructionID),
		"fromStatus=5",
		"fromRetryCount=20",
		"lastError=",
	}
	if eventCount != 1 || eventID <= 0 || eventType != "ASSET_INSTRUCTION_MANUAL_RETRY" ||
		reason != "SETTLEMENT_BALANCE_TOPUP_VERIFIED" || operatorID != 9002 ||
		eventUserID != shortUserID || eventContractID != contractID {
		t.Fatalf("manual retry audit count/id/type/reason/operator/user/contract=%d/%d/%q/%q/%d/%d/%d detail=%q",
			eventCount, eventID, eventType, reason, operatorID, eventUserID, eventContractID, detail)
	}
	for _, part := range wantDetailParts {
		if !strings.Contains(detail, part) {
			t.Fatalf("manual retry audit detail missing %q: %q", part, detail)
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_trading_control_event SET reason='TAMPERED' WHERE id=?`, eventID); err == nil {
		t.Fatal("manual retry audit event update unexpectedly succeeded")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_trading_control_event WHERE id=?`, eventID); err == nil {
		t.Fatal("manual retry audit event delete unexpectedly succeeded")
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trading_control_event WHERE id=?`, eventID).Scan(&eventCount); err != nil {
		t.Fatal(err)
	}
	if eventCount != 1 {
		t.Fatalf("manual retry audit event must remain immutable, count=%d", eventCount)
	}
}

func assertCompletedCashSettlement(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longPositionID, shortPositionID, marginLotID int64,
) {
	assertCompletedCashSettlementWithPayoff(
		t, ctx, db, contractID, longPositionID, shortPositionID, marginLotID,
		"40.0000000000000000",
	)
}

func assertCompletedCashSettlementWithPayoff(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longPositionID, shortPositionID, marginLotID int64,
	wantDetailPayoff string,
) {
	t.Helper()
	var settlementStatus, batchStatus, successCount, contractStatus, instructionTotal, instructionSuccess, reconciled int64
	var remainingMargin, pendingMargin string
	if err := db.QueryRowContext(ctx, `SELECT s.status,b.status,b.success_count,c.status,
		(SELECT COUNT(*) FROM t_option_asset_instruction i WHERE i.biz_no=s.settlement_no),
		(SELECT COUNT(*) FROM t_option_asset_instruction i WHERE i.biz_no=s.settlement_no AND i.status=?),
		(SELECT COUNT(*) FROM t_option_asset_instruction i WHERE i.biz_no=s.settlement_no AND i.reconciliation_status=?),
		CAST(l.remaining_margin AS CHAR),CAST(l.pending_margin AS CHAR)
		FROM t_option_settlement s
		JOIN t_option_settlement_batch b ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
		JOIN t_option_contract c ON c.id=s.contract_id
		JOIN t_option_margin_lot l ON l.id=?
		WHERE s.tenant_id=? AND s.contract_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		marginLotID, p0AssetE2ETenantID, contractID,
	).Scan(&settlementStatus, &batchStatus, &successCount, &contractStatus,
		&instructionTotal, &instructionSuccess, &reconciled, &remainingMargin, &pendingMargin); err != nil {
		t.Fatal(err)
	}
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE) ||
		successCount != 3 || contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) ||
		instructionTotal != 3 || instructionSuccess != 3 || reconciled != 3 || remainingMargin != "0.0000000000000000" ||
		pendingMargin != "0.0000000000000000" {
		t.Fatalf("completed settlement evidence=%d/%d/%d/%d/%d/%d/%d/%s/%s",
			settlementStatus, batchStatus, successCount, contractStatus,
			instructionTotal, instructionSuccess, reconciled, remainingMargin, pendingMargin)
	}
	var detailPayoff string
	if err := db.QueryRowContext(ctx, `SELECT CAST(SUM(payoff) AS CHAR)
		FROM t_option_settlement_detail WHERE position_id IN (?,?)`, longPositionID, shortPositionID,
	).Scan(&detailPayoff); err != nil {
		t.Fatal(err)
	}
	if detailPayoff != wantDetailPayoff {
		t.Fatalf("settlement detail total payoff=%s want=%s", detailPayoff, wantDetailPayoff)
	}
}

func assertP0SettlementAssetConservation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longUserID, shortUserID int64,
) {
	assertP0SettlementAssetConservationWithTotal(
		t, ctx, db, contractID, longUserID, shortUserID, "200.000000000000000000",
	)
}

func assertP0SettlementAssetConservationWithTotal(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, longUserID, shortUserID int64,
	wantTotal string,
) {
	t.Helper()
	var total, available, frozen string
	var settlementFlows int64
	if err := db.QueryRowContext(ctx, `SELECT CAST(SUM(total_amount) AS CHAR),
		CAST(SUM(available_amount) AS CHAR),CAST(SUM(frozen_amount) AS CHAR)
		FROM t_user_asset WHERE tenant_id=? AND wallet_type=? AND coin='USDT' AND user_id IN (?,?)`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION), longUserID, shortUserID,
	).Scan(&total, &available, &frozen); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM t_asset_flow AS flow
		JOIN t_option_asset_instruction AS instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		JOIN t_option_settlement_batch AS batch
		  ON batch.tenant_id=instruction.tenant_id AND batch.batch_no=instruction.biz_no
		WHERE flow.tenant_id=? AND batch.contract_id=?
		  AND flow.biz_type='option' AND flow.user_id IN (?,?)`,
		p0AssetE2ETenantID, contractID, longUserID, shortUserID,
	).Scan(&settlementFlows); err != nil {
		t.Fatal(err)
	}
	if total != wantTotal || available != wantTotal ||
		frozen != "0.000000000000000000" || settlementFlows != 3 {
		t.Fatalf("settlement conservation total/available/frozen/flows=%s/%s/%s/%d",
			total, available, frozen, settlementFlows)
	}
}

func testP0MarginCoinRelease(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		callUserID int64 = 105
		putUserID  int64 = 106
	)
	now := time.Now().Unix()
	seedP0PhysicalContract(t, ctx, db, 996201, "P0-E2E-PHYSICAL-CALL", option.OptionType_OPTION_TYPE_CALL, now)
	seedP0PhysicalContract(t, ctx, db, 996202, "P0-E2E-PHYSICAL-PUT", option.OptionType_OPTION_TYPE_PUT, now)
	creditAssetCoin(t, ctx, assetClient, callUserID, "BTC", "1", "P0-MARGIN-CALL-SEED")
	creditAssetCoin(t, ctx, assetClient, putUserID, "USDT", "1000", "P0-MARGIN-PUT-SEED")

	callOrder := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P0-MARGIN-CALL", UserId: callUserID,
		AccountId: 5001, ContractId: 996201, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.Side_SIDE_SELL), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType: int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(10),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1), FeeCoin: "USDT",
		MarginAmount: decimal.RequireFromString("0.25"), MarginCoin: "BTC",
		Source: int64(option.OrderSource_ORDER_SOURCE_APP), ReduceOnly: int64(common.YesNo_YES_NO_NO),
		Mmp: int64(common.YesNo_YES_NO_NO), Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now, UpdateTimes: now,
	})
	putOrder := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P0-MARGIN-PUT", UserId: putUserID,
		AccountId: 6001, ContractId: 996202, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.Side_SIDE_SELL), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType: int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(10),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1), FeeCoin: "USDT",
		MarginAmount: decimal.NewFromInt(300), MarginCoin: "USDT",
		Source: int64(option.OrderSource_ORDER_SOURCE_APP), ReduceOnly: int64(common.YesNo_YES_NO_NO),
		Mmp: int64(common.YesNo_YES_NO_NO), Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now, UpdateTimes: now,
	})
	wrongCall := *callOrder
	wrongCall.OrderNo = "P0-MARGIN-CALL-WRONG-COIN"
	wrongCall.MarginCoin = "USDT"
	if _, err := serviceCtx.OptionOrderModel.Insert(ctx, &wrongCall); err == nil {
		t.Fatal("physical Call seller order with USDT collateral must be rejected by the database")
	}

	for _, freeze := range []struct {
		order  *models.TOptionOrder
		coin   string
		amount string
	}{
		{order: callOrder, coin: "BTC", amount: "0.25"},
		{order: putOrder, coin: "USDT", amount: "300"},
	} {
		resp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
			TenantId: p0AssetE2ETenantID, UserId: freeze.order.UserId,
			WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: freeze.coin, Amount: freeze.amount,
			BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
			BizId: freeze.order.Id, BizNo: freeze.order.OrderNo, Remark: "P0 margin coin acceptance freeze",
		})
		assertAssetOK(t, resp, err)
	}
	assertWalletCoinAmounts(t, ctx, db, callUserID, "BTC", "1.000000000000000000", "0.750000000000000000", "0.250000000000000000")
	assertWalletCoinAmounts(t, ctx, db, putUserID, "USDT", "1000.000000000000000000", "700.000000000000000000", "300.000000000000000000")

	for _, order := range []*models.TOptionOrder{callOrder, putOrder} {
		canceled, err := applogic.CancelOrderByControl(ctx, serviceCtx, order.Id, "P0_MARGIN_COIN_ACCEPTANCE")
		if err != nil {
			t.Fatalf("cancel order %s: %v", order.OrderNo, err)
		}
		if canceled == nil || canceled.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELING) {
			t.Fatalf("order %s did not enter canceling: %+v", order.OrderNo, canceled)
		}
	}
	assertPendingReleaseCoin(t, ctx, db, callOrder.Id, "BTC")
	assertPendingReleaseCoin(t, ctx, db, putOrder.Id, "USDT")

	processAssetInstructions(t, ctx, serviceCtx)
	assertCompletedMarginOrder(t, ctx, serviceCtx, callOrder.Id)
	assertCompletedMarginOrder(t, ctx, serviceCtx, putOrder.Id)
	assertWalletCoinAmounts(t, ctx, db, callUserID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, putUserID, "USDT", "1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
	assertOptionMirrorCoin(t, ctx, db, callUserID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	assertOptionMirrorCoin(t, ctx, db, putUserID, "USDT", "1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
	assertAssetBizFlowCount(t, ctx, db, callUserID, "BTC", "place_order", 1)
	assertAssetBizFlowCount(t, ctx, db, callUserID, "BTC", "cancel_order", 1)
	assertAssetBizFlowCount(t, ctx, db, putUserID, "USDT", "place_order", 1)
	assertAssetBizFlowCount(t, ctx, db, putUserID, "USDT", "cancel_order", 1)
	processAssetInstructions(t, ctx, serviceCtx)
	assertAssetBizFlowCount(t, ctx, db, callUserID, "BTC", "cancel_order", 1)
	assertAssetBizFlowCount(t, ctx, db, putUserID, "USDT", "cancel_order", 1)
}

func seedP0PhysicalContract(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	code string,
	optionType option.OptionType,
	now int64,
) {
	t.Helper()
	_, err := db.ExecContext(ctx, `
		INSERT INTO t_option_contract (
			id,tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
			option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
			max_order_qty,price_tick,qty_step,multiplier,list_time,exercise_cutoff_time,expire_time,
			deliver_time,max_user_long_qty,max_user_short_qty,max_open_interest,order_price_band_ratio,
			circuit_breaker_ratio,greeks_max_age_seconds,seller_margin_mode,initial_margin_rate,
			maintenance_margin_rate,min_margin_rate,status,is_deleted,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		contractID, p0AssetE2ETenantID, code, "BTCUSDT", "BTC", "USDT", "USDT",
		int64(optionType), int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL), "100", "1", "1", "1000", "0.1", "1", "1",
		now-3600, now+3600, now+7200, now+7200, "10000", "10000", "10000", "0.2", "0.5", 60,
		int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY), "0.2", "0.1", "0.05",
		int64(option.ContractStatus_CONTRACT_STATUS_PENDING), int64(common.YesNo_YES_NO_NO), now, now,
	)
	if err != nil {
		t.Fatalf("seed physical contract %s: %v", code, err)
	}
}

func insertP0MarginOrder(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	order *models.TOptionOrder,
) *models.TOptionOrder {
	t.Helper()
	result, err := serviceCtx.OptionOrderModel.Insert(ctx, order)
	if err != nil {
		t.Fatalf("insert margin order %s: %v", order.OrderNo, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	stored, err := serviceCtx.OptionOrderModel.FindOne(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func assertPendingReleaseCoin(t *testing.T, ctx context.Context, db *sql.DB, orderID int64, coin string) {
	t.Helper()
	var count int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND order_id=? AND action=? AND status=? AND coin=?`,
		p0AssetE2ETenantID, orderID,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING), coin,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("order %d pending release coin %s count=%d want=1", orderID, coin, count)
	}
}

func assertCompletedMarginOrder(t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, orderID int64) {
	t.Helper()
	order, err := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) || !order.MarginAmount.IsZero() {
		t.Fatalf("completed margin order=%+v", order)
	}
}

func waitForAssetRPC(t *testing.T, ctx context.Context, client asset.AssetClient) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		_, err := client.GetAssetBalance(ctx, &asset.GetUserAssetDetailReq{
			TenantId: p0AssetE2ETenantID, UserId: 1,
			WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT",
		})
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Asset RPC did not become ready: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func testP0RiskWalletAndEquity(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	for _, seed := range []struct {
		userID int64
		amount string
	}{
		{userID: 101, amount: "980"},
		{userID: 102, amount: "1020"},
		{userID: 103, amount: "1000"},
	} {
		creditAsset(t, ctx, assetClient, seed.userID, seed.amount, fmt.Sprintf("P0-RISK-SEED-%d", seed.userID))
	}

	now := time.Now().Unix()
	_, err := db.ExecContext(ctx, `
		INSERT INTO t_option_contract (
			id,tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
			option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
			max_order_qty,price_tick,qty_step,multiplier,list_time,exercise_cutoff_time,expire_time,
			deliver_time,max_user_long_qty,max_user_short_qty,max_open_interest,order_price_band_ratio,
			circuit_breaker_ratio,greeks_max_age_seconds,seller_margin_mode,initial_margin_rate,
			maintenance_margin_rate,min_margin_rate,status,is_deleted,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		996101, p0AssetE2ETenantID, "P0-E2E-CALL", "BTCUSDT", "BTC", "USDT", "USDT",
		int64(option.OptionType_OPTION_TYPE_CALL), int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		int64(option.SettlementType_SETTLEMENT_TYPE_CASH), "100", "1", "1", "1000", "0.1", "1", "1",
		now-3600, now+3600, now+7200, now+7200, "10000", "10000", "10000", "0.2", "0.5", 60,
		int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED), "0.2", "0.1", "0.05",
		int64(option.ContractStatus_CONTRACT_STATUS_PENDING), int64(common.YesNo_YES_NO_NO), now, now,
	)
	if err != nil {
		t.Fatalf("seed option contract: %v", err)
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO t_option_market (
			id,tenant_id,contract_id,underlying_price,mark_price,last_price,bid_price,ask_price,
			theoretical_price,snapshot_time,underlying_snapshot_time,mark_snapshot_time,
			greeks_snapshot_time,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		996111, p0AssetE2ETenantID, 996101, "100", "10", "10", "9.9", "10.1", "10",
		now, now, now, now, now, now,
	)
	if err != nil {
		t.Fatalf("seed option market: %v", err)
	}
	positions := []struct {
		id, userID, accountID, side int64
		qty, margin                 string
	}{
		{id: 996121, userID: 101, accountID: 1001, side: int64(common.PositionSide_POSITION_SIDE_LONG), qty: "2", margin: "0"},
		{id: 996122, userID: 102, accountID: 2001, side: int64(common.PositionSide_POSITION_SIDE_SHORT), qty: "2", margin: "50"},
		{id: 996123, userID: 103, accountID: 3001, side: int64(common.PositionSide_POSITION_SIDE_LONG), qty: "2", margin: "0"},
		{id: 996124, userID: 103, accountID: 3002, side: int64(common.PositionSide_POSITION_SIDE_SHORT), qty: "1", margin: "25"},
	}
	for _, position := range positions {
		_, err = db.ExecContext(ctx, `
			INSERT INTO t_option_position (
				id,tenant_id,user_id,account_id,contract_id,underlying_symbol,side,position_qty,
				available_qty,open_avg_price,margin_amount,exerciseable_qty,status,create_times,update_times
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			position.id, p0AssetE2ETenantID, position.userID, position.accountID, 996101, "BTCUSDT",
			position.side, position.qty, position.qty, "10", position.margin, position.qty,
			int64(option.PositionStatus_POSITION_STATUS_HOLDING), now, now,
		)
		if err != nil {
			t.Fatalf("seed position %d: %v", position.id, err)
		}
	}

	resp, err := NewProcessRiskAccountsLogic(ctx, serviceCtx).ProcessRiskAccounts(&option.OptionTaskReq{
		TenantId: p0AssetE2ETenantID,
	})
	if err != nil {
		t.Fatalf("process risk accounts through real Asset RPC: %v", err)
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("unexpected risk task response: %+v", resp)
	}

	type riskRow struct {
		userID, accountID                                   int64
		equity, netValue, positionMargin, maintenanceMargin string
	}
	rows, err := db.QueryContext(ctx, `
		SELECT user_id,account_id,
			CAST(equity AS CHAR),CAST(net_option_value AS CHAR),
			CAST(position_margin AS CHAR),CAST(maintenance_margin AS CHAR)
		FROM t_option_risk_account
		WHERE tenant_id=? AND settle_coin='USDT'
		ORDER BY user_id`, p0AssetE2ETenantID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []riskRow
	for rows.Next() {
		var row riskRow
		if err := rows.Scan(&row.userID, &row.accountID, &row.equity, &row.netValue, &row.positionMargin, &row.maintenanceMargin); err != nil {
			t.Fatal(err)
		}
		got = append(got, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("risk account rows=%d want=3: %+v", len(got), got)
	}
	want := []riskRow{
		{userID: 101, accountID: 0, equity: "1000.0000000000000000", netValue: "20.0000000000000000", positionMargin: "0.0000000000000000", maintenanceMargin: "0.0000000000000000"},
		{userID: 102, accountID: 0, equity: "1000.0000000000000000", netValue: "-20.0000000000000000", positionMargin: "50.0000000000000000", maintenanceMargin: "20.0000000000000000"},
		{userID: 103, accountID: 0, equity: "1010.0000000000000000", netValue: "10.0000000000000000", positionMargin: "25.0000000000000000", maintenanceMargin: "10.0000000000000000"},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("risk row[%d]=%+v want=%+v", i, got[i], want[i])
		}
	}
}

func testP0FreezeReleaseReplay(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const userID int64 = 104
	creditAsset(t, ctx, assetClient, userID, "1000", "P0-FREEZE-SEED")
	now := time.Now().Unix()
	freezeA := insertAssetInstruction(t, ctx, serviceCtx, &models.TOptionAssetInstruction{
		TenantId: p0AssetE2ETenantID, InstructionNo: "P0-FREEZE-A-INSTRUCTION",
		UserId: userID, AccountId: 4001,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		TargetBizNo: "P0-FREEZE-A", Coin: "USDT", Amount: decimal.NewFromInt(100),
		StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	})
	insertAssetInstruction(t, ctx, serviceCtx, &models.TOptionAssetInstruction{
		TenantId: p0AssetE2ETenantID, InstructionNo: "P0-FREEZE-B-INSTRUCTION",
		UserId: userID, AccountId: 4002,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		TargetBizNo: "P0-FREEZE-B", Coin: "USDT", Amount: decimal.NewFromInt(200),
		StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	})

	freezeReq := &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "100",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: freezeA.Id, BizNo: freezeA.TargetBizNo, Remark: "option asset instruction freeze",
	}
	freezeResp, err := assetClient.FreezeAsset(ctx, freezeReq)
	assertAssetOK(t, freezeResp, err)
	changed := proto.Clone(freezeReq).(*asset.FreezeAssetReq)
	changed.Amount = "101"
	if _, err := assetClient.FreezeAsset(ctx, changed); err == nil {
		t.Fatal("reusing a freeze idempotency key with a different amount must fail")
	}
	// Simulate a freeze committed by the pre-fix Asset service: the freeze and
	// flow exist, but there is no idempotency row because the RPC response was
	// lost before Option could persist success. The retry must adopt that
	// evidence rather than freezing the wallet a second time.
	if _, err := db.ExecContext(ctx, `
		DELETE FROM t_asset_idempotent
		WHERE tenant_id=? AND biz_type='option' AND scene_type='place_order' AND biz_no=?`,
		p0AssetE2ETenantID, freezeA.TargetBizNo,
	); err != nil {
		t.Fatalf("simulate legacy freeze response loss: %v", err)
	}

	processAssetInstructions(t, ctx, serviceCtx)
	assertWalletAmounts(t, ctx, db, userID, "1000.000000000000000000", "700.000000000000000000", "300.000000000000000000")
	assertFlowCount(t, ctx, db, "place_order", 2)
	assertOptionMirror(t, ctx, db, userID, "1000.000000000000000000", "700.000000000000000000", "300.000000000000000000")
	processAssetInstructions(t, ctx, serviceCtx)
	assertFlowCount(t, ctx, db, "place_order", 2)

	releaseA := insertAssetInstruction(t, ctx, serviceCtx, &models.TOptionAssetInstruction{
		TenantId: p0AssetE2ETenantID, InstructionNo: "P0-RELEASE-A-INSTRUCTION",
		UserId: userID, AccountId: 4001,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
		TargetBizNo: "P0-FREEZE-A", Coin: "USDT", Amount: decimal.NewFromInt(100),
		StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	})
	insertAssetInstruction(t, ctx, serviceCtx, &models.TOptionAssetInstruction{
		TenantId: p0AssetE2ETenantID, InstructionNo: "P0-RELEASE-B-INSTRUCTION",
		UserId: userID, AccountId: 4002,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
		TargetBizNo: "P0-FREEZE-B", Coin: "USDT", Amount: decimal.NewFromInt(200),
		StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	})
	releaseResp, err := assetClient.UnfreezeAssetByBizNo(ctx, &asset.UnfreezeAssetByBizNoReq{
		TenantId: p0AssetE2ETenantID, TargetBizType: asset.BizType_BIZ_TYPE_OPTION,
		TargetBizNo: releaseA.TargetBizNo, Amount: releaseA.Amount.String(),
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_CANCEL_ORDER,
		BizId: releaseA.Id, BizNo: releaseA.InstructionNo, Remark: "option asset instruction release frozen",
	})
	assertAssetOK(t, releaseResp, err)

	processAssetInstructions(t, ctx, serviceCtx)
	assertWalletAmounts(t, ctx, db, userID, "1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
	assertFlowCount(t, ctx, db, "cancel_order", 2)
	assertOptionMirror(t, ctx, db, userID, "1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")

	var successful, reconciled, distinctFlows int64
	if err := db.QueryRowContext(ctx, `
		SELECT SUM(status=3),SUM(reconciliation_status=2),COUNT(DISTINCT asset_flow_no)
		FROM t_option_asset_instruction WHERE tenant_id=?`, p0AssetE2ETenantID,
	).Scan(&successful, &reconciled, &distinctFlows); err != nil {
		t.Fatal(err)
	}
	if successful != 4 || reconciled != 4 || distinctFlows != 4 {
		t.Fatalf("instruction evidence success=%d reconciled=%d flows=%d want=4/4/4", successful, reconciled, distinctFlows)
	}
}

func creditAsset(t *testing.T, ctx context.Context, client asset.AssetClient, userID int64, amount, bizNo string) {
	creditAssetCoin(t, ctx, client, userID, "USDT", amount, bizNo)
}

func creditAssetCoin(
	t *testing.T,
	ctx context.Context,
	client asset.AssetClient,
	userID int64,
	coin, amount, bizNo string,
) {
	t.Helper()
	resp, err := client.AddAvailable(ctx, &asset.AddAvailableReq{
		TenantId: p0AssetE2ETenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: coin, Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
		BizNo: bizNo, Remark: "P0 Asset RPC acceptance seed",
	})
	assertAssetOK(t, resp, err)
}

func assertAssetOK(t *testing.T, response interface{ GetBase() *common.RespBase }, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.GetBase() == nil || response.GetBase().Code != 200 {
		t.Fatalf("Asset RPC rejected request: %+v", response)
	}
}

func insertAssetInstruction(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	instruction *models.TOptionAssetInstruction,
) *models.TOptionAssetInstruction {
	t.Helper()
	result, err := serviceCtx.OptionAssetInstructionModel.Insert(ctx, instruction)
	if err != nil {
		t.Fatalf("insert instruction %s: %v", instruction.InstructionNo, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	stored, err := serviceCtx.OptionAssetInstructionModel.FindOne(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func processAssetInstructions(t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext) {
	t.Helper()
	resp, err := NewProcessAssetInstructionsLogic(ctx, serviceCtx).ProcessAssetInstructions(&option.OptionTaskReq{
		TenantId: p0AssetE2ETenantID,
	})
	if err != nil {
		t.Fatalf("process asset instructions: %v", err)
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("unexpected asset task response: %+v", resp)
	}
}

func assertWalletAmounts(t *testing.T, ctx context.Context, db *sql.DB, userID int64, total, available, frozen string) {
	assertWalletCoinAmounts(t, ctx, db, userID, "USDT", total, available, frozen)
}

func assertWalletCoinAmounts(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID int64,
	coin, total, available, frozen string,
) {
	t.Helper()
	var gotTotal, gotAvailable, gotFrozen string
	if err := db.QueryRowContext(ctx, `
		SELECT CAST(total_amount AS CHAR),CAST(available_amount AS CHAR),CAST(frozen_amount AS CHAR)
		FROM t_user_asset WHERE tenant_id=? AND user_id=? AND wallet_type=? AND coin=?`,
		p0AssetE2ETenantID, userID, int64(common.WalletType_WALLET_TYPE_OPTION), coin,
	).Scan(&gotTotal, &gotAvailable, &gotFrozen); err != nil {
		t.Fatal(err)
	}
	if gotTotal != total || gotAvailable != available || gotFrozen != frozen {
		t.Fatalf("Asset wallet=%s/%s/%s want=%s/%s/%s", gotTotal, gotAvailable, gotFrozen, total, available, frozen)
	}
}

func assertOptionMirror(t *testing.T, ctx context.Context, db *sql.DB, userID int64, total, available, frozen string) {
	assertOptionMirrorCoin(t, ctx, db, userID, "USDT", total, available, frozen)
}

func assertOptionMirrorCoin(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID int64,
	coin, total, available, frozen string,
) {
	t.Helper()
	var count, accountID int64
	var gotTotal, gotAvailable, gotFrozen string
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*),COALESCE(MAX(account_id),-1),
			COALESCE(CAST(MAX(balance) AS CHAR),''),
			COALESCE(CAST(MAX(available_balance) AS CHAR),''),
			COALESCE(CAST(MAX(frozen_balance) AS CHAR),'')
		FROM t_option_account WHERE tenant_id=? AND user_id=? AND margin_coin=?`,
		p0AssetE2ETenantID, userID, coin,
	).Scan(&count, &accountID, &gotTotal, &gotAvailable, &gotFrozen); err != nil {
		t.Fatal(err)
	}
	if count != 1 || accountID != 0 || gotTotal != total || gotAvailable != available || gotFrozen != frozen {
		t.Fatalf("Option mirror count/account/amounts=%d/%d/%s/%s/%s", count, accountID, gotTotal, gotAvailable, gotFrozen)
	}
}

func assertAssetBizFlowCount(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID int64,
	coin, scene string,
	want int64,
) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND user_id=? AND wallet_type=? AND coin=?
		  AND biz_type='option' AND scene_type=?`,
		p0AssetE2ETenantID, userID, int64(common.WalletType_WALLET_TYPE_OPTION), coin, scene,
	).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Asset flow user=%d coin=%s scene=%s count=%d want=%d", userID, coin, scene, got, want)
	}
}

func assertFlowCount(t *testing.T, ctx context.Context, db *sql.DB, scene string, want int64) {
	t.Helper()
	var count int64
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND user_id=104 AND wallet_type=? AND biz_type='option' AND scene_type=?`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION), scene,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("Asset %s flow count=%d want=%d", scene, count, want)
	}
}
