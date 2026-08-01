package tasklogic

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type failOnceLiquidationCoverageClient struct {
	asset.AssetClient

	mu                sync.Mutex
	insuranceFailures int
	backstopFailures  int
}

func (c *failOnceLiquidationCoverageClient) CoverInsuranceDeficit(
	ctx context.Context,
	in *asset.CoverInsuranceDeficitReq,
	opts ...grpc.CallOption,
) (*asset.CoverInsuranceDeficitResp, error) {
	c.mu.Lock()
	shouldFail := c.insuranceFailures == 0
	if shouldFail {
		c.insuranceFailures++
	}
	c.mu.Unlock()
	resp, err := c.AssetClient.CoverInsuranceDeficit(ctx, in, opts...)
	if err != nil || !shouldFail {
		return resp, err
	}
	return nil, status.Error(codes.Unavailable,
		"P0 LIQ-003 injected insurance response loss after commit")
}

func (c *failOnceLiquidationCoverageClient) CoverPlatformBackstopDeficit(
	ctx context.Context,
	in *asset.CoverPlatformBackstopDeficitReq,
	opts ...grpc.CallOption,
) (*asset.CoverPlatformBackstopDeficitResp, error) {
	c.mu.Lock()
	shouldFail := c.backstopFailures == 0
	if shouldFail {
		c.backstopFailures++
	}
	c.mu.Unlock()
	resp, err := c.AssetClient.CoverPlatformBackstopDeficit(ctx, in, opts...)
	if err != nil || !shouldFail {
		return resp, err
	}
	return nil, status.Error(codes.Unavailable,
		"P0 LIQ-003 injected platform backstop response loss after commit")
}

func (c *failOnceLiquidationCoverageClient) failureCounts() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.insuranceFailures, c.backstopFailures
}

func testP0IsolatedShortLiquidationAccounting(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		shortUserID     int64 = 141
		longUserID      int64 = 142
		insuranceUserID int64 = 143
		feeUserID       int64 = 144
	)
	now := time.Now().Unix()
	contract := insertP0LiquidationContract(
		t, ctx, serviceCtx, "P0-ISOLATED-SHORT-LIQUIDATION-CALL",
		insuranceUserID, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "140", "40", now)
	creditAsset(t, ctx, assetClient, shortUserID, "100", "P0-LIQUIDATION-SHORT-SEED")
	creditAsset(t, ctx, assetClient, longUserID, "100", "P0-LIQUIDATION-LONG-SEED")
	transferP0OptionPremium(t, ctx, assetClient, longUserID, shortUserID, "20", "P0-LIQUIDATION-OPEN-PREMIUM")

	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: 8040,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		MarginAmount: decimal.NewFromInt(100), MaintenanceMargin: decimal.NewFromInt(40),
		UnrealizedPnl: decimal.NewFromInt(-60),
		Status:        int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes:   now - 200, UpdateTimes: now - 200,
	})
	insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: 7040,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		UnrealizedPnl: decimal.NewFromInt(60), ExerciseableQty: decimal.NewFromInt(2),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-LIQUIDATION-SHORT-MARGIN", "2", "100", now-190,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, shortPosition, lot, "100")

	liquidation := &models.TOptionLiquidation{
		TenantId: p0AssetE2ETenantID, LiquidationNo: "P0-ISOLATED-SHORT-LIQUIDATION",
		UserId: shortUserID, AccountId: 8040, ContractId: contract.Id, PositionId: shortPosition.Id,
		Quantity: decimal.NewFromInt(2), MarkPrice: decimal.NewFromInt(40),
		MaintenanceMargin: decimal.NewFromInt(40), Equity: decimal.NewFromInt(20),
		Status:            int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		DeficitResolution: int64(option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE),
		CreateTimes:       now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionLiquidationModel.Insert(ctx, liquidation)
	if err != nil {
		t.Fatalf("insert liquidation: %v", err)
	}
	liquidation.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	processP0Liquidations(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	completed, err := serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) ||
		!completed.CollateralAmount.Equal(decimal.NewFromInt(88)) ||
		!completed.LiquidationFee.Equal(decimal.NewFromInt(8)) ||
		!completed.DeficitAmount.IsZero() || !completed.RemainingDeficit.IsZero() ||
		completed.TakeoverPositionId <= 0 {
		t.Fatalf("unexpected completed liquidation: %+v", completed)
	}
	source, err := serviceCtx.OptionPositionModel.FindOne(ctx, shortPosition.Id)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, source, option.PositionStatus_POSITION_STATUS_CLOSED,
		"0", "0", "0", "0", "0", "-60", "8", "-68",
	)
	takeover, err := serviceCtx.OptionPositionModel.FindOne(ctx, completed.TakeoverPositionId)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, takeover, option.PositionStatus_POSITION_STATUS_HOLDING,
		"2", "80", "40", "80", "0", "0", "0", "0",
	)
	takeoverLot, err := serviceCtx.OptionMarginLotModel.FindOneByTenantIdTradeId(
		ctx, p0AssetE2ETenantID, -liquidation.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	if takeoverLot.PositionId != takeover.Id || takeoverLot.CollateralCoin != "USDT" ||
		!takeoverLot.RemainingQuantity.Equal(decimal.NewFromInt(2)) ||
		!takeoverLot.RemainingMargin.Equal(decimal.NewFromInt(80)) ||
		takeoverLot.Status != int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE) {
		t.Fatalf("unexpected takeover margin lot: %+v", takeoverLot)
	}

	assertWalletAmounts(t, ctx, db, shortUserID, "32.000000000000000000", "32.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, longUserID, "80.000000000000000000", "80.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, insuranceUserID, "80.000000000000000000", "0.000000000000000000", "80.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "8.000000000000000000", "8.000000000000000000", "0.000000000000000000")
	assertP0LiquidationEvidence(t, ctx, db, liquidation.Id, contract.Id)

	processP0Liquidations(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0LiquidationEvidence(t, ctx, db, liquidation.Id, contract.Id)
	assertP0PartialLiquidationFailsClosed(t, ctx, db, serviceCtx, contract, now)
}

func testP0LiquidationDeficitFailureRecovery(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		shortUserID     int64 = 146
		insuranceUserID int64 = 147
		feeUserID       int64 = 148
	)
	now := time.Now().Unix()
	contract := insertP0LiquidationContract(
		t, ctx, serviceCtx, "P0-LIQUIDATION-DEFICIT-BACKSTOP-CALL",
		insuranceUserID, feeUserID, now,
	)
	contract.LiquidationDeficitPolicy = int64(
		option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP,
	)
	if err := serviceCtx.OptionContractModel.Update(ctx, contract); err != nil {
		t.Fatalf("enable platform backstop policy: %v", err)
	}
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "140", "40", now)
	seedP0LiquidationPlatformAccount(t, ctx, db, "INSURANCE_FUND", "15", now)
	seedP0LiquidationPlatformAccount(t, ctx, db, "OPTION_BACKSTOP", "0", now)
	creditAsset(t, ctx, assetClient, shortUserID, "50", "P0-LIQ-DEFICIT-SHORT-SEED")

	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: 8042,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(40),
		UnrealizedPnl: decimal.NewFromInt(-60),
		Status:        int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes:   now - 200, UpdateTimes: now - 200,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-LIQ-DEFICIT-SHORT-MARGIN", "2", "50", now-190,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, shortPosition, lot, "50")

	liquidation := &models.TOptionLiquidation{
		TenantId: p0AssetE2ETenantID, LiquidationNo: "P0-LIQUIDATION-DEFICIT-RECOVERY",
		UserId: shortUserID, AccountId: 8042, ContractId: contract.Id, PositionId: shortPosition.Id,
		Quantity: decimal.NewFromInt(2), MarkPrice: decimal.NewFromInt(40),
		MaintenanceMargin: decimal.NewFromInt(40), Equity: decimal.NewFromInt(-10),
		Status:            int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		DeficitResolution: int64(option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE),
		CreateTimes:       now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionLiquidationModel.Insert(ctx, liquidation)
	if err != nil {
		t.Fatalf("insert deficit liquidation: %v", err)
	}
	liquidation.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	coverageFault := &failOnceLiquidationCoverageClient{AssetClient: assetClient}
	deductFault := &failOnceDeductAssetClient{
		AssetClient: coverageFault, targetBizNo: lot.FreezeBizNo, failAfterCommit: true,
	}
	originalClient := serviceCtx.AssetClient
	serviceCtx.AssetClient = deductFault
	defer func() { serviceCtx.AssetClient = originalClient }()

	logic := NewProcessLiquidationsLogic(ctx, serviceCtx)
	if err := logic.processOne(liquidation); err == nil {
		t.Fatal("insurance response loss unexpectedly succeeded")
	}
	assertP0LiquidationPreparationFailure(t, ctx, db, liquidation.Id, 1, "0", "0", 1, 0, 0)

	current, err := serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if err := logic.processOne(current); err == nil {
		t.Fatal("platform backstop response loss unexpectedly succeeded")
	}
	assertP0LiquidationPreparationFailure(t, ctx, db, liquidation.Id, 2, "0", "-23", 1, 1, 1)

	current, err = serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if err := logic.processOne(current); err != nil {
		t.Fatalf("recover liquidation coverage preparation: %v", err)
	}
	insuranceFailures, backstopFailures := coverageFault.failureCounts()
	if insuranceFailures != 1 || backstopFailures != 1 {
		t.Fatalf("coverage failures insurance/backstop=%d/%d want=1/1",
			insuranceFailures, backstopFailures)
	}
	assertP0LiquidationPrepared(t, ctx, db, liquidation.Id)

	processAssetInstructions(t, ctx, serviceCtx)
	if deductFault.failureCount() != 1 {
		t.Fatalf("liquidation collateral response losses=%d want=1", deductFault.failureCount())
	}
	assertP0LiquidationInstructionBarrier(t, ctx, db, liquidation.Id)
	if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET next_retry_at=0 WHERE tenant_id=? AND liquidation_id=? AND status=?`,
		p0AssetE2ETenantID, liquidation.Id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
	); err != nil {
		t.Fatalf("make liquidation debit retryable: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	completed, err := serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) ||
		!completed.CollateralAmount.Equal(decimal.NewFromInt(50)) ||
		!completed.DeficitAmount.Equal(decimal.NewFromInt(38)) ||
		!completed.InsuranceFundAmount.Equal(decimal.NewFromInt(15)) ||
		!completed.BackstopAmount.Equal(decimal.NewFromInt(23)) ||
		!completed.RemainingDeficit.IsZero() ||
		completed.DeficitResolution != int64(option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_INSURANCE_AND_BACKSTOP) {
		t.Fatalf("unexpected recovered deficit liquidation: %+v", completed)
	}
	source, err := serviceCtx.OptionPositionModel.FindOne(ctx, shortPosition.Id)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, source, option.PositionStatus_POSITION_STATUS_CLOSED,
		"0", "0", "0", "0", "0", "-60", "8", "-68",
	)
	takeover, err := serviceCtx.OptionPositionModel.FindOne(ctx, completed.TakeoverPositionId)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, takeover, option.PositionStatus_POSITION_STATUS_HOLDING,
		"2", "80", "40", "80", "0", "0", "0", "0",
	)
	assertWalletAmounts(t, ctx, db, shortUserID, "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, insuranceUserID, "80.000000000000000000", "0.000000000000000000", "80.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "8.000000000000000000", "8.000000000000000000", "0.000000000000000000")
	assertP0DeficitLiquidationEvidence(t, ctx, db, liquidation.Id)

	if err := logic.processOne(completed); err != nil {
		t.Fatalf("replay completed deficit liquidation: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0DeficitLiquidationEvidence(t, ctx, db, liquidation.Id)
}

func seedP0LiquidationPlatformAccount(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	accountType, amount string,
	now int64,
) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `INSERT INTO t_asset_platform_account
		(tenant_id,account_type,coin,available_amount,frozen_amount,status,version,create_times,update_times)
		VALUES (?,?,?, ?,0,1,0,?,?)`,
		p0AssetE2ETenantID, accountType, "USDT", amount, now*1000, now*1000,
	); err != nil {
		t.Fatalf("seed platform account %s: %v", accountType, err)
	}
}

func assertP0LiquidationPreparationFailure(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	liquidationID, wantRetry int64,
	wantInsuranceBalance, wantBackstopBalance string,
	wantInsuranceCovers, wantBackstopCovers, wantOptionCoverFlows int64,
) {
	t.Helper()
	var statusValue, retryCount, instructions int64
	if err := db.QueryRowContext(ctx, `SELECT status,retry_count FROM t_option_liquidation
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&statusValue, &retryCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED) ||
		retryCount != wantRetry || instructions != 0 {
		t.Fatalf("preparation failure status/retry/instructions=%d/%d/%d want=%d/%d/0",
			statusValue, retryCount, instructions,
			int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED), wantRetry)
	}
	assertP0LiquidationPlatformEvidence(
		t, ctx, db, liquidationID, wantInsuranceBalance, wantBackstopBalance,
		wantInsuranceCovers, wantBackstopCovers, wantOptionCoverFlows,
	)
}

func assertP0LiquidationPrepared(t *testing.T, ctx context.Context, db *sql.DB, liquidationID int64) {
	t.Helper()
	var statusValue, instructions int64
	var collateral, deficit, insurance, backstop decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT status,collateral_amount,deficit_amount,
		insurance_fund_amount,backstop_amount FROM t_option_liquidation
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&statusValue, &collateral, &deficit, &insurance, &backstop); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING) ||
		instructions != 4 || !collateral.Equal(decimal.NewFromInt(50)) ||
		!deficit.Equal(decimal.NewFromInt(38)) || !insurance.Equal(decimal.NewFromInt(15)) ||
		!backstop.Equal(decimal.NewFromInt(23)) {
		t.Fatalf("unexpected prepared liquidation status/instructions/amounts=%d/%d/%s/%s/%s/%s",
			statusValue, instructions, collateral, deficit, insurance, backstop)
	}
	assertP0LiquidationPlatformEvidence(t, ctx, db, liquidationID, "0", "-23", 1, 1, 1)
}

func assertP0LiquidationInstructionBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	liquidationID int64,
) {
	t.Helper()
	var failed, pending, success, flows int64
	if err := db.QueryRowContext(ctx, `SELECT SUM(status=?),SUM(status=?),SUM(status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND liquidation_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, liquidationID,
	).Scan(&failed, &pending, &success); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=flow.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=?
		   THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.liquidation_id=?`,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		p0AssetE2ETenantID, liquidationID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if failed != 1 || pending != 3 || success != 0 || flows != 1 {
		// The source debit committed but its response was lost. No takeover
		// credit, fee credit, or takeover freeze may run before reconciliation.
		t.Fatalf("liquidation failure barrier failed/pending/success/source_flows=%d/%d/%d/%d",
			failed, pending, success, flows)
	}
}

func assertP0LiquidationPlatformEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	liquidationID int64,
	wantInsuranceBalance, wantBackstopBalance string,
	wantInsuranceCovers, wantBackstopCovers, wantOptionCoverFlows int64,
) {
	t.Helper()
	var insuranceBalance, backstopBalance decimal.Decimal
	var insuranceCovers, backstopCovers, optionCoverFlows int64
	if err := db.QueryRowContext(ctx, `SELECT available_amount FROM t_asset_platform_account
		WHERE tenant_id=? AND account_type='INSURANCE_FUND' AND coin='USDT'`,
		p0AssetE2ETenantID).Scan(&insuranceBalance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT available_amount FROM t_asset_platform_account
		WHERE tenant_id=? AND account_type='OPTION_BACKSTOP' AND coin='USDT'`,
		p0AssetE2ETenantID).Scan(&backstopBalance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_insurance_cover
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&insuranceCovers); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_backstop_cover
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&backstopCovers); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_insurance_fund_flow
		WHERE tenant_id=? AND liquidation_id=? AND flow_type=?`,
		p0AssetE2ETenantID, liquidationID,
		int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER),
	).Scan(&optionCoverFlows); err != nil {
		t.Fatal(err)
	}
	if !insuranceBalance.Equal(decimal.RequireFromString(wantInsuranceBalance)) ||
		!backstopBalance.Equal(decimal.RequireFromString(wantBackstopBalance)) ||
		insuranceCovers != wantInsuranceCovers || backstopCovers != wantBackstopCovers ||
		optionCoverFlows != wantOptionCoverFlows {
		t.Fatalf("platform evidence balances/covers/option_flows=%s/%s/%d/%d/%d want=%s/%s/%d/%d/%d",
			insuranceBalance, backstopBalance, insuranceCovers, backstopCovers, optionCoverFlows,
			wantInsuranceBalance, wantBackstopBalance, wantInsuranceCovers, wantBackstopCovers,
			wantOptionCoverFlows)
	}
}

func assertP0DeficitLiquidationEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	liquidationID int64,
) {
	t.Helper()
	var instructions, success, reconciled, flows, insuranceFlows, backstopFlows int64
	var walletTotal, platformTotal decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND liquidation_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, liquidationID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction ON instruction.tenant_id=flow.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=? THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.liquidation_id=?`,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		p0AssetE2ETenantID, liquidationID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT SUM(account_type='INSURANCE_FUND'),
		SUM(account_type='OPTION_BACKSTOP') FROM t_asset_platform_flow
		WHERE tenant_id=? AND biz_id=?`, p0AssetE2ETenantID, liquidationID).
		Scan(&insuranceFlows, &backstopFlows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_amount),0) FROM t_user_asset
		WHERE tenant_id=? AND wallet_type=? AND coin='USDT' AND user_id IN (146,147,148)`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION)).
		Scan(&walletTotal); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT SUM(available_amount) FROM t_asset_platform_account
		WHERE tenant_id=? AND coin='USDT' AND account_type IN ('INSURANCE_FUND','OPTION_BACKSTOP')`,
		p0AssetE2ETenantID).Scan(&platformTotal); err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPlatformEvidence(t, ctx, db, liquidationID, "0", "-23", 1, 1, 1)
	if instructions != 4 || success != 4 || reconciled != 4 || flows != 4 ||
		insuranceFlows != 1 || backstopFlows != 1 ||
		!walletTotal.Equal(decimal.NewFromInt(88)) || !platformTotal.Equal(decimal.NewFromInt(-23)) ||
		!walletTotal.Add(platformTotal).Equal(decimal.NewFromInt(65)) {
		t.Fatalf("deficit liquidation evidence instructions/success/reconciled/flows/platform=%d/%d/%d/%d/%d/%d wallet/platform/conserved=%s/%s/%s",
			instructions, success, reconciled, flows, insuranceFlows, backstopFlows,
			walletTotal, platformTotal, walletTotal.Add(platformTotal))
	}
}

func assertP0PartialLiquidationFailsClosed(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	now int64,
) {
	t.Helper()
	position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: 145, AccountId: 8041,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		MarginAmount: decimal.NewFromInt(100), MaintenanceMargin: decimal.NewFromInt(40),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now, UpdateTimes: now,
	})
	liquidation := &models.TOptionLiquidation{
		TenantId: p0AssetE2ETenantID, LiquidationNo: "P0-PARTIAL-LIQUIDATION-REJECTED",
		UserId: 145, AccountId: 8041, ContractId: contract.Id, PositionId: position.Id,
		Quantity: decimal.NewFromInt(1), MarkPrice: decimal.NewFromInt(40),
		MaintenanceMargin: decimal.NewFromInt(40), Status: int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		DeficitResolution: int64(option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE),
		CreateTimes:       now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionLiquidationModel.Insert(ctx, liquidation)
	if err != nil {
		t.Fatal(err)
	}
	liquidation.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	logic := NewProcessLiquidationsLogic(ctx, serviceCtx)
	if err := logic.processOne(liquidation); err == nil {
		t.Fatal("partial liquidation unexpectedly succeeded")
	}
	stored, err := serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED) ||
		stored.RetryCount != 1 || stored.LastErrorMsg != "partial option liquidation is not supported" {
		t.Fatalf("unexpected partial liquidation failure evidence: %+v", stored)
	}
	var instructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidation.Id).
		Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	unchanged, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
	if err != nil {
		t.Fatal(err)
	}
	if instructions != 0 || !unchanged.PositionQty.Equal(decimal.NewFromInt(2)) ||
		!unchanged.MarginAmount.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("partial liquidation changed protected state instructions=%d position=%+v", instructions, unchanged)
	}
}

func testP0PortfolioLiquidationFailsClosed(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	contract := insertP0LiquidationContract(
		t, ctx, serviceCtx, "P0-PORTFOLIO-LIQUIDATION-REJECTED-CALL", 149, 150, now,
	)
	contract.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)
	if err := serviceCtx.OptionContractModel.Update(ctx, contract); err != nil {
		t.Fatalf("enable portfolio margin for fail-closed acceptance: %v", err)
	}
	position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: 151, AccountId: 8043,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		MarginAmount: decimal.NewFromInt(100), MaintenanceMargin: decimal.NewFromInt(40),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now, UpdateTimes: now,
	})
	liquidation := &models.TOptionLiquidation{
		TenantId: p0AssetE2ETenantID, LiquidationNo: "P0-PORTFOLIO-LIQUIDATION-REJECTED",
		UserId: 151, AccountId: 8043, ContractId: contract.Id, PositionId: position.Id,
		Quantity: decimal.NewFromInt(2), MarkPrice: decimal.NewFromInt(40),
		MaintenanceMargin: decimal.NewFromInt(40),
		Status:            int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		DeficitResolution: int64(option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE),
		CreateTimes:       now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionLiquidationModel.Insert(ctx, liquidation)
	if err != nil {
		t.Fatal(err)
	}
	liquidation.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if err := NewProcessLiquidationsLogic(ctx, serviceCtx).processOne(liquidation); err == nil {
		t.Fatal("portfolio liquidation unexpectedly succeeded")
	}
	stored, err := serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
	if err != nil {
		t.Fatal(err)
	}
	var instructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidation.Id).
		Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if stored.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED) ||
		stored.RetryCount != 1 || stored.LastErrorMsg != "portfolio option liquidation is not supported" ||
		instructions != 0 || !unchanged.PositionQty.Equal(decimal.NewFromInt(2)) ||
		!unchanged.MarginAmount.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("portfolio liquidation did not fail closed liquidation=%+v instructions=%d position=%+v",
			stored, instructions, unchanged)
	}
}

func insertP0LiquidationContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractCode string,
	insuranceUserID, feeUserID, now int64,
) *models.TOptionContract {
	t.Helper()
	contract := &models.TOptionContract{
		TenantId: p0AssetE2ETenantID, ContractCode: contractCode,
		UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		ExerciseStyle:  int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		StrikePrice:    decimal.NewFromInt(100), ContractUnit: decimal.NewFromInt(1),
		MinOrderQty: decimal.RequireFromString("0.5"), MaxOrderQty: decimal.NewFromInt(1000),
		PriceTick: decimal.RequireFromString("0.1"), QtyStep: decimal.RequireFromString("0.5"),
		Multiplier: decimal.NewFromInt(1), ListTime: now - 3600,
		ExerciseCutoffTime: now + 3600, ExpireTime: now + 7200, DeliverTime: now + 7200,
		AutoExerciseThreshold: decimal.NewFromInt(10), MaxUserLongQty: decimal.NewFromInt(10000),
		MaxUserShortQty: decimal.NewFromInt(10000), MaxOpenInterest: decimal.NewFromInt(10000),
		OrderPriceBandRatio: decimal.RequireFromString("0.2"),
		CircuitBreakerRatio: decimal.RequireFromString("0.5"), GreeksMaxAgeSeconds: 60,
		SettlementPriceSource: "authoritative-market", SettlementPriceMethod: "MEDIAN",
		SettlementWindowSeconds: 60, SettlementMinSamples: 3,
		IsAutoExercise:  int64(common.YesNo_YES_NO_NO),
		ExerciseFeeRate: decimal.RequireFromString("0.1"), FeeUserId: feeUserID, FeeAccountId: 9041,
		SellerMarginMode:      int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
		InitialMarginRate:     decimal.RequireFromString("0.5"),
		MaintenanceMarginRate: decimal.RequireFromString("0.2"), MinMarginRate: decimal.RequireFromString("0.1"),
		LiquidationFeeRate: decimal.RequireFromString("0.1"),
		InsuranceUserId:    insuranceUserID, InsuranceAccountId: 9040,
		LiquidationDeficitPolicy: int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW),
		TradingCalendarCode:      "CONTINUOUS_24_7", Status: int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		IsDeleted: int64(common.YesNo_YES_NO_NO), CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert liquidation contract: %v", err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func processP0Liquidations(t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext) {
	t.Helper()
	resp, err := NewProcessLiquidationsLogic(ctx, serviceCtx).ProcessLiquidations(&option.OptionTaskReq{
		TenantId: p0AssetE2ETenantID,
	})
	if err != nil {
		t.Fatalf("process liquidations: %v", err)
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("unexpected liquidation task response: %+v", resp)
	}
}

func assertP0LiquidationPosition(
	t *testing.T,
	position *models.TOptionPosition,
	status option.PositionStatus,
	qty, margin, maintenance, value, unrealized, tradeRealized, fee, total string,
) {
	t.Helper()
	wants := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"qty", position.PositionQty, qty}, {"margin", position.MarginAmount, margin},
		{"maintenance", position.MaintenanceMargin, maintenance},
		{"value", position.PositionValue, value}, {"unrealized", position.UnrealizedPnl, unrealized},
		{"trade_realized", position.TradeRealizedPnl, tradeRealized},
		{"fee", position.FeePaid, fee}, {"total", position.TotalReturn, total},
		{"realized", position.RealizedPnl, total},
	}
	for _, item := range wants {
		if !item.got.Equal(decimal.RequireFromString(item.want)) {
			t.Fatalf("liquidation position %d %s=%s want=%s", position.Id, item.name, item.got, item.want)
		}
	}
	if position.Status != int64(status) {
		t.Fatalf("liquidation position %d status=%d want=%d", position.Id, position.Status, status)
	}
}

func assertP0LiquidationEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	liquidationID, contractID int64,
) {
	t.Helper()
	var instructions, success, reconciled, flows, liquidations, feeFlows int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND liquidation_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, liquidationID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction ON instruction.tenant_id=flow.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=? THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.liquidation_id=?`,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		p0AssetE2ETenantID, liquidationID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_liquidation
		WHERE tenant_id=? AND id=? AND contract_id=? AND status=?`,
		p0AssetE2ETenantID, liquidationID, contractID,
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE),
	).Scan(&liquidations); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_insurance_fund_flow
		WHERE tenant_id=? AND liquidation_id=? AND flow_type=? AND amount=8`,
		p0AssetE2ETenantID, liquidationID,
		int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_LIQUIDATION_FEE),
	).Scan(&feeFlows); err != nil {
		t.Fatal(err)
	}
	if instructions != 5 || success != 5 || reconciled != 5 || flows != 5 ||
		liquidations != 1 || feeFlows != 1 {
		t.Fatalf("liquidation evidence instructions/success/reconciled/flows/liquidations/fees=%d/%d/%d/%d/%d/%d",
			instructions, success, reconciled, flows, liquidations, feeFlows)
	}
}
