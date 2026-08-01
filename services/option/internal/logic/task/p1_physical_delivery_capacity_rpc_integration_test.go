package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	gosqlx "github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const p1PhysicalDeliveryCapacityUnits = 501

func TestP1PhysicalDeliveryCapacityAssetRPC(t *testing.T) {
	gosqlx.DisableLog()
	if raw := os.Getenv("OPTION_P1_PHYSICAL_CAPACITY_UNITS"); raw == "" {
		t.Skip("OPTION_P1_PHYSICAL_CAPACITY_UNITS is required")
	} else if raw != fmt.Sprint(p1PhysicalDeliveryCapacityUnits) {
		t.Fatalf("OPTION_P1_PHYSICAL_CAPACITY_UNITS=%q must be %d", raw, p1PhysicalDeliveryCapacityUnits)
	}
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	rpcAddr := os.Getenv("OPTION_P0_ASSET_E2E_RPC_ADDR")
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || rpcAddr == "" || redisAddr == "" {
		t.Fatal("OPTION_P0_ASSET_E2E_DSN, OPTION_P0_ASSET_E2E_RPC_ADDR and OPTION_P0_ASSET_E2E_REDIS_ADDR are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
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

	runP1PhysicalDeliveryCapacity(t, ctx, db, assetClient, serviceCtx)
}

func runP1PhysicalDeliveryCapacity(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		contractID  int64 = 999902
		longUserID  int64 = 3151
		shortUserID int64 = 3152
		accountBase int64 = 800000
	)
	now := time.Now().Unix()
	prefix := "P1-PHYSICAL-CALL-CAPACITY-501"
	seedP1PhysicalContract(
		t, ctx, db, contractID, prefix,
		option.OptionType_OPTION_TYPE_CALL, now-10, now-1,
	)
	creditAssetCoin(t, ctx, assetClient, longUserID, "USDT", "50100", prefix+"-LONG-SEED")
	creditAssetCoin(t, ctx, assetClient, shortUserID, "BTC", "501", prefix+"-SHORT-SEED")

	seedStarted := time.Now()
	for i := 0; i < p1PhysicalDeliveryCapacityUnits; i++ {
		insertP1PhysicalPosition(
			t, ctx, serviceCtx, contractID, longUserID, accountBase+int64(i),
			common.PositionSide_POSITION_SIDE_LONG, now-10000+int64(i), decimal.Zero,
		)
	}
	quantity := decimal.NewFromInt(p1PhysicalDeliveryCapacityUnits)
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: shortUserID,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: quantity,
		AvailableQty: quantity, OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(10), PositionValue: quantity.Mul(decimal.NewFromInt(10)),
		MarginAmount: quantity, ExerciseableQty: quantity,
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 9000, UpdateTimes: now - 9000,
	})
	lot := insertP1PhysicalCapacityMarginLot(t, ctx, serviceCtx, shortPosition, prefix, quantity, now-8000)
	freezeP1PhysicalCollateral(t, ctx, assetClient, lot, "BTC", quantity.String())
	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contractID, now-10, now, prefix,
		physicalEvidencePrices("120"), "120",
	)
	seedElapsed := time.Since(seedStarted)

	clearingStarted := time.Now()
	if err := NewProcessContractLifecycleLogic(ctx, serviceCtx).processExpiredContracts(now); err != nil {
		t.Fatalf("create capacity physical delivery: %v", err)
	}
	clearingElapsed := time.Since(clearingStarted)
	assertP1PhysicalCapacityCreated(t, ctx, db, contractID)

	assetStarted := time.Now()
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assetElapsed := time.Since(assetStarted)
	assertP1PhysicalCapacityCompleted(t, ctx, db, contractID, lot.Id, longUserID, shortUserID)

	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1PhysicalCapacityCompleted(t, ctx, db, contractID, lot.Id, longUserID, shortUserID)
	t.Logf("physical_delivery_capacity_units=%d seed=%s clearing=%s asset_rpc=%s instructions=%d",
		p1PhysicalDeliveryCapacityUnits, seedElapsed, clearingElapsed, assetElapsed,
		p1PhysicalDeliveryCapacityUnits*4)
}

func insertP1PhysicalCapacityMarginLot(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	position *models.TOptionPosition,
	prefix string,
	quantity decimal.Decimal,
	createTimes int64,
) *models.TOptionMarginLot {
	t.Helper()
	lot := &models.TOptionMarginLot{
		TenantId: position.TenantId, UserId: position.UserId, AccountId: position.AccountId,
		ContractId: position.ContractId, PositionId: position.Id,
		OriginContractId: position.ContractId, OriginPositionId: position.Id,
		TradeId: -position.Id, FreezeBizNo: prefix + "-SHORT-COLLATERAL", CollateralCoin: "BTC",
		Quantity: quantity, RemainingQuantity: quantity,
		InitialMargin: quantity, RemainingMargin: quantity,
		Status:      int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		CreateTimes: createTimes, UpdateTimes: createTimes,
	}
	result, err := serviceCtx.OptionMarginLotModel.Insert(ctx, lot)
	if err != nil {
		t.Fatalf("insert physical capacity margin lot: %v", err)
	}
	lot.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return lot
}

func assertP1PhysicalCapacityCreated(t *testing.T, ctx context.Context, db *sql.DB, contractID int64) {
	t.Helper()
	var units, instructions, pending, details int64
	var firstUnit, lastUnit string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		COALESCE(MIN(delivery_unit_no),''),COALESCE(MAX(delivery_unit_no),'')
		FROM t_option_physical_delivery_unit WHERE tenant_id=? AND contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&units, &firstUnit, &lastUnit); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(instruction.status=?)
		FROM t_option_asset_instruction instruction
		JOIN t_option_physical_delivery_unit unit
		  ON unit.tenant_id=instruction.tenant_id AND unit.id=instruction.delivery_unit_id
		WHERE unit.tenant_id=? AND unit.contract_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		p0AssetE2ETenantID, contractID,
	).Scan(&instructions, &pending); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_settlement_detail
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&details); err != nil {
		t.Fatal(err)
	}
	if units != p1PhysicalDeliveryCapacityUnits || instructions != p1PhysicalDeliveryCapacityUnits*4 ||
		pending != instructions || details != p1PhysicalDeliveryCapacityUnits+1 ||
		firstUnit == "" || firstUnit[len(firstUnit)-8:] != "DU000001" ||
		lastUnit == "" || lastUnit[len(lastUnit)-8:] != "DU000501" {
		t.Fatalf("physical capacity creation units/instructions/pending/details/first/last=%d/%d/%d/%d/%q/%q",
			units, instructions, pending, details, firstUnit, lastUnit)
	}
}

func assertP1PhysicalCapacityCompleted(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, lotID, longUserID, shortUserID int64,
) {
	t.Helper()
	var settlementStatus, batchStatus, contractStatus, batchInstructions, batchSuccess int64
	if err := db.QueryRowContext(ctx, `SELECT settlement.status,batch.status,contract.status,
		batch.instruction_count,batch.success_count
		FROM t_option_settlement settlement
		JOIN t_option_settlement_batch batch
		  ON batch.tenant_id=settlement.tenant_id AND batch.batch_no=settlement.settlement_no
		JOIN t_option_contract contract
		  ON contract.tenant_id=settlement.tenant_id AND contract.id=settlement.contract_id
		WHERE settlement.tenant_id=? AND settlement.contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&settlementStatus, &batchStatus, &contractStatus, &batchInstructions, &batchSuccess); err != nil {
		t.Fatal(err)
	}
	wantInstructions := int64(p1PhysicalDeliveryCapacityUnits * 4)
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE) ||
		contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) ||
		batchInstructions != wantInstructions || batchSuccess != wantInstructions {
		t.Fatalf("physical capacity terminal settlement/batch/contract/instructions/success=%d/%d/%d/%d/%d",
			settlementStatus, batchStatus, contractStatus, batchInstructions, batchSuccess)
	}

	var units, completed, events, flowAnomalies, duplicateInstructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?)
		FROM t_option_physical_delivery_unit WHERE tenant_id=? AND contract_id=?`,
		int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_COMPLETED),
		p0AssetE2ETenantID, contractID,
	).Scan(&units, &completed); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trading_control_event
		WHERE tenant_id=? AND contract_id=? AND event_type='PHYSICAL_DELIVERY_COMPLETED'`,
		p0AssetE2ETenantID, contractID,
	).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (
		SELECT instruction.id FROM t_option_asset_instruction instruction
		JOIN t_option_physical_delivery_unit unit
		  ON unit.tenant_id=instruction.tenant_id AND unit.id=instruction.delivery_unit_id
		LEFT JOIN t_asset_flow flow
		  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
		WHERE unit.tenant_id=? AND unit.contract_id=?
		GROUP BY instruction.id HAVING COUNT(flow.id)<>1
	) anomaly`, p0AssetE2ETenantID, contractID).Scan(&flowAnomalies); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (
		SELECT instruction.instruction_no FROM t_option_asset_instruction instruction
		JOIN t_option_physical_delivery_unit unit
		  ON unit.tenant_id=instruction.tenant_id AND unit.id=instruction.delivery_unit_id
		WHERE unit.tenant_id=? AND unit.contract_id=?
		GROUP BY instruction.instruction_no HAVING COUNT(*)<>1
	) duplicate_instruction`, p0AssetE2ETenantID, contractID).Scan(&duplicateInstructions); err != nil {
		t.Fatal(err)
	}
	if units != p1PhysicalDeliveryCapacityUnits || completed != units || events != units ||
		flowAnomalies != 0 || duplicateInstructions != 0 {
		t.Fatalf("physical capacity units/completed/events/flow-anomalies/duplicate-instructions=%d/%d/%d/%d/%d",
			units, completed, events, flowAnomalies, duplicateInstructions)
	}
	assertP1PhysicalInstructionCounts(t, ctx, db, contractID, wantInstructions, wantInstructions, wantInstructions)

	var lotStatus int64
	var remaining, pending string
	if err := db.QueryRowContext(ctx, `SELECT status,CAST(remaining_margin AS CHAR),CAST(pending_margin AS CHAR)
		FROM t_option_margin_lot WHERE id=?`, lotID).Scan(&lotStatus, &remaining, &pending); err != nil {
		t.Fatal(err)
	}
	if lotStatus != int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED) ||
		remaining != "0.0000000000000000" || pending != "0.0000000000000000" {
		t.Fatalf("physical capacity lot status/remaining/pending=%d/%s/%s", lotStatus, remaining, pending)
	}
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, longUserID, "BTC", "501.000000000000000000", "501.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "BTC", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "50100.000000000000000000", "50100.000000000000000000", "0.000000000000000000")
}
