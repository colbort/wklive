package tasklogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	applogic "wklive/services/option/internal/logic/app"
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
	assertP1InsuranceInventoryGovernedExit(
		t, ctx, db, assetClient, serviceCtx, contract, takeover, takeoverLot,
		longUserID, insuranceUserID,
	)
	assertP0PartialLiquidationAccounting(t, ctx, db, assetClient, serviceCtx, now)
}

func assertP1InsuranceInventoryGovernedExit(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	takeover *models.TOptionPosition,
	takeoverLot *models.TOptionMarginLot,
	longUserID, insuranceUserID int64,
) {
	t.Helper()
	const (
		creatorID  int64 = 9141
		reviewerID int64 = 9142
		executorID int64 = 9143
	)
	creditAsset(t, ctx, assetClient, insuranceUserID, "100", "P1-005-INSURANCE-EXIT-PREMIUM-BUDGET")
	makerResp, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, longUserID), serviceCtx,
	).PlaceOrder(&option.PlaceOrderReq{
		AccountId: 7040, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_CLOSE,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "40", Qty: "1",
		ClientOrderId: "P1-005-INSURANCE-EXIT-LIQUIDITY",
		ReduceOnly:    common.YesNo_YES_NO_YES,
	})
	if err != nil || makerResp == nil || makerResp.Base == nil || makerResp.Base.Code != 200 ||
		makerResp.Data == nil || makerResp.Data.OrderId <= 0 {
		t.Fatalf("create insurance-exit maker liquidity response=%+v err=%v", makerResp, err)
	}
	createResp, err := adminlogic.NewCreateInsuranceInventoryExitLogic(
		p0AdminContext(ctx, creatorID, p0AssetE2ETenantID), serviceCtx,
	).CreateInsuranceInventoryExit(&option.CreateInsuranceInventoryExitReq{
		TenantId: p0AssetE2ETenantID, PositionId: takeover.Id,
		Quantity: "2", LimitPrice: "40",
		Reason:      "reduce concentrated insurance takeover inventory",
		EvidenceRef: "P1-005-INSURANCE-EXIT-DEPTH-AND-RISK-EVIDENCE",
	})
	if err != nil || createResp == nil || createResp.Base == nil || createResp.Base.Code != 200 ||
		createResp.Data == nil || createResp.Data.Id <= 0 {
		t.Fatalf("create insurance inventory exit response=%+v err=%v", createResp, err)
	}
	exitID := createResp.Data.Id
	selfReview, err := adminlogic.NewReviewInsuranceInventoryExitLogic(
		p0AdminContext(ctx, creatorID, p0AssetE2ETenantID), serviceCtx,
	).ReviewInsuranceInventoryExit(&option.ReviewInsuranceInventoryExitReq{
		TenantId: p0AssetE2ETenantID, ExitId: exitID, Approve: true,
		Reason: "creator must not self approve",
	})
	if err != nil || selfReview == nil || selfReview.Base == nil || selfReview.Base.Code == 200 {
		t.Fatalf("insurance inventory exit self-review result=%+v err=%v", selfReview, err)
	}
	approved, err := adminlogic.NewReviewInsuranceInventoryExitLogic(
		p0AdminContext(ctx, reviewerID, p0AssetE2ETenantID), serviceCtx,
	).ReviewInsuranceInventoryExit(&option.ReviewInsuranceInventoryExitReq{
		TenantId: p0AssetE2ETenantID, ExitId: exitID, Approve: true,
		Reason: "independent inventory risk approval",
	})
	if err != nil || approved == nil || approved.Base == nil || approved.Base.Code != 200 ||
		approved.Data == nil || approved.Data.Status !=
		option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_APPROVED {
		t.Fatalf("approve insurance inventory exit response=%+v err=%v", approved, err)
	}

	type executeResult struct {
		resp *option.GetInsuranceInventoryExitResp
		err  error
	}
	results := make(chan executeResult, 20)
	var wait sync.WaitGroup
	for index := 0; index < 20; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			resp, executeErr := adminlogic.NewExecuteInsuranceInventoryExitLogic(
				p0AdminContext(ctx, executorID, p0AssetE2ETenantID), serviceCtx,
			).ExecuteInsuranceInventoryExit(&option.ExecuteInsuranceInventoryExitReq{
				TenantId: p0AssetE2ETenantID, ExitId: exitID,
			})
			results <- executeResult{resp: resp, err: executeErr}
		}()
	}
	wait.Wait()
	close(results)
	orderID := int64(0)
	for result := range results {
		if result.err != nil || result.resp == nil || result.resp.Base == nil ||
			result.resp.Base.Code != 200 || result.resp.Data == nil ||
			result.resp.Data.Status != option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_SUBMITTED ||
			result.resp.Data.OrderId <= 0 {
			t.Fatalf("concurrent insurance exit execution response=%+v err=%v", result.resp, result.err)
		}
		if orderID == 0 {
			orderID = result.resp.Data.OrderId
		} else if orderID != result.resp.Data.OrderId {
			t.Fatalf("concurrent insurance exit created orders %d and %d", orderID, result.resp.Data.OrderId)
		}
	}

	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	order, err := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if order.Source != int64(option.OrderSource_ORDER_SOURCE_ADMIN) ||
		order.OrderType != int64(option.OrderType_ORDER_TYPE_IOC) ||
		order.Side != int64(common.Side_SIDE_BUY) ||
		order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_CLOSE) ||
		order.ReduceOnly != int64(common.YesNo_YES_NO_YES) ||
		!order.Qty.Equal(decimal.NewFromInt(2)) || !order.FilledQty.Equal(decimal.NewFromInt(1)) ||
		!order.UnfilledQty.Equal(decimal.NewFromInt(1)) ||
		order.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) {
		t.Fatalf("unexpected insurance inventory IOC result: %+v", order)
	}
	listed, err := adminlogic.NewListInsuranceInventoryExitsLogic(
		p0AdminContext(ctx, executorID, p0AssetE2ETenantID), serviceCtx,
	).ListInsuranceInventoryExits(&option.ListInsuranceInventoryExitsReq{
		TenantId: p0AssetE2ETenantID,
		Page:     &common.PageReq{Limit: 10},
	})
	if err != nil || listed == nil || listed.Base == nil || listed.Base.Code != 200 ||
		len(listed.Data) != 1 || listed.Data[0].Id != exitID ||
		listed.Data[0].OrderStatus != option.OrderStatus_ORDER_STATUS_CANCELED ||
		listed.Data[0].FilledQty != "1" || listed.Data[0].UnfilledQty != "1" {
		t.Fatalf("list insurance inventory exit response=%+v err=%v", listed, err)
	}
	remaining, err := serviceCtx.OptionPositionModel.FindOne(ctx, takeover.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !remaining.PositionQty.Equal(decimal.NewFromInt(1)) ||
		!remaining.AvailableQty.Equal(decimal.NewFromInt(1)) || !remaining.FrozenQty.IsZero() ||
		!remaining.MarginAmount.Equal(decimal.NewFromInt(40)) ||
		!remaining.TradeRealizedPnl.IsZero() || !remaining.FeePaid.IsZero() ||
		!remaining.TotalReturn.IsZero() {
		t.Fatalf("unexpected insurance inventory after partial IOC exit: %+v", remaining)
	}
	remainingLot, err := serviceCtx.OptionMarginLotModel.FindOne(ctx, takeoverLot.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !remainingLot.RemainingQuantity.Equal(decimal.NewFromInt(1)) ||
		!remainingLot.RemainingMargin.Equal(decimal.NewFromInt(40)) ||
		remainingLot.Status != int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE) {
		t.Fatalf("unexpected insurance margin lot after exit: %+v", remainingLot)
	}
	assertWalletAmounts(t, ctx, db, insuranceUserID,
		"140.000000000000000000", "100.000000000000000000", "40.000000000000000000")
	assertWalletAmounts(t, ctx, db, longUserID,
		"120.000000000000000000", "120.000000000000000000", "0.000000000000000000")

	var exits, orders, clientKeys, events, instructions, instructionFlows, budgetFlows int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_insurance_inventory_exit WHERE id=? AND status=4 AND order_id=?),
		(SELECT COUNT(*) FROM t_option_order WHERE tenant_id=? AND user_id=? AND client_order_id=?),
		(SELECT COUNT(*) FROM t_option_client_order_key WHERE tenant_id=? AND user_id=? AND client_order_id=?),
		(SELECT COUNT(*) FROM t_option_trading_control_event WHERE tenant_id=? AND event_type LIKE 'INSURANCE_EXIT_%'),
		(SELECT COUNT(*) FROM t_option_asset_instruction WHERE tenant_id=? AND (order_id=? OR margin_lot_id=?)),
		(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction instruction
			JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
			AND flow.biz_no=CASE WHEN instruction.action=1
				THEN instruction.target_biz_no ELSE instruction.instruction_no END
			WHERE instruction.tenant_id=? AND flow.user_id IN (?,?)
			AND (instruction.order_id=? OR instruction.margin_lot_id=?)),
		(SELECT COUNT(*) FROM t_asset_flow WHERE tenant_id=? AND user_id=? AND biz_no=?)`,
		exitID, orderID,
		p0AssetE2ETenantID, insuranceUserID, insuranceInventoryExitClientOrderIDForTest(exitID),
		p0AssetE2ETenantID, insuranceUserID, insuranceInventoryExitClientOrderIDForTest(exitID),
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, orderID, takeoverLot.Id,
		p0AssetE2ETenantID, insuranceUserID, longUserID,
		orderID, takeoverLot.Id,
		p0AssetE2ETenantID, insuranceUserID, "P1-005-INSURANCE-EXIT-PREMIUM-BUDGET",
	).Scan(&exits, &orders, &clientKeys, &events, &instructions, &instructionFlows, &budgetFlows); err != nil {
		t.Fatal(err)
	}
	if exits != 1 || orders != 1 || clientKeys != 1 || events != 3 || instructions != 4 ||
		instructionFlows != 4 || budgetFlows != 1 {
		t.Fatalf("insurance exit evidence exits/orders/keys/events/instructions/instruction-flows/budget-flows=%d/%d/%d/%d/%d/%d/%d",
			exits, orders, clientKeys, events, instructions, instructionFlows, budgetFlows)
	}
	reserved, err := serviceCtx.OptionInsuranceInventoryExitModel.SumReservedQuantity(
		ctx, p0AssetE2ETenantID, contract.Id,
		time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), time.Now().UTC().Day(), 0, 0, 0, 0, time.UTC).Unix(),
	)
	if err != nil || !reserved.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("insurance exit UTC daily reservation=%s err=%v", reserved, err)
	}
	dailyLimitResp, err := adminlogic.NewCreateInsuranceInventoryExitLogic(
		p0AdminContext(ctx, creatorID, p0AssetE2ETenantID), serviceCtx,
	).CreateInsuranceInventoryExit(&option.CreateInsuranceInventoryExitReq{
		TenantId: p0AssetE2ETenantID, PositionId: takeover.Id,
		Quantity: "1", LimitPrice: "40",
		Reason:      "must not exceed the UTC daily insurance exit quantity budget",
		EvidenceRef: "P1-005-INSURANCE-EXIT-DAILY-LIMIT-EVIDENCE",
	})
	if err != nil || dailyLimitResp == nil || dailyLimitResp.Base == nil || dailyLimitResp.Base.Code == 200 {
		t.Fatalf("insurance exit daily-limit response=%+v err=%v", dailyLimitResp, err)
	}
	var exitsAfterLimit, eventsAfterLimit int64
	if err = db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_insurance_inventory_exit WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_trading_control_event WHERE tenant_id=? AND event_type LIKE 'INSURANCE_EXIT_%')`,
		p0AssetE2ETenantID, contract.Id, p0AssetE2ETenantID,
	).Scan(&exitsAfterLimit, &eventsAfterLimit); err != nil {
		t.Fatal(err)
	}
	if exitsAfterLimit != 1 || eventsAfterLimit != 3 {
		t.Fatalf("daily limit created evidence exits/events=%d/%d", exitsAfterLimit, eventsAfterLimit)
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_insurance_inventory_exit SET quantity=1 WHERE id=?`, exitID); err == nil {
		t.Fatal("database allowed insurance exit quantity mutation")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_insurance_inventory_exit WHERE id=?`, exitID); err == nil {
		t.Fatal("database allowed insurance exit deletion")
	}
	t.Logf("P1-005 insurance_exit=%d order=%d filled=1 remaining=1 margin=40 instructions=4 instruction_flows=4 budget_flows=1 concurrent=20",
		exitID, orderID)
}

func insuranceInventoryExitClientOrderIDForTest(exitID int64) string {
	return fmt.Sprintf("INS-EXIT-%d", exitID)
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
	if optionCoverFlows == 1 {
		var flowID int64
		var rawAmount, signedAmount decimal.Decimal
		if err := db.QueryRowContext(ctx, `SELECT id,amount,
			CASE WHEN flow_type IN (2,4) THEN -ABS(amount) ELSE ABS(amount) END
			FROM t_option_insurance_fund_flow
			WHERE tenant_id=? AND liquidation_id=? AND flow_type=?`,
			p0AssetE2ETenantID, liquidationID,
			int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER),
		).Scan(&flowID, &rawAmount, &signedAmount); err != nil {
			t.Fatal(err)
		}
		if !rawAmount.Equal(decimal.NewFromInt(15)) || !signedAmount.Equal(decimal.NewFromInt(-15)) {
			t.Fatalf("insurance flow raw/signed=%s/%s want=15/-15", rawAmount, signedAmount)
		}
		if _, err := db.ExecContext(ctx, `UPDATE t_option_insurance_fund_flow SET amount=amount WHERE id=?`, flowID); err == nil {
			t.Fatal("insurance fund flow update was not rejected")
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM t_option_insurance_fund_flow WHERE id=?`, flowID); err == nil {
			t.Fatal("insurance fund flow delete was not rejected")
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO t_option_insurance_fund_flow
		(tenant_id,flow_no,contract_id,liquidation_id,flow_type,coin,amount,asset_flow_no,create_times)
		VALUES (?,?,?,?,?,'USDT',-1,?,?)`,
		p0AssetE2ETenantID, fmt.Sprintf("P0-INSURANCE-NEGATIVE-%d", liquidationID),
		int64(1), liquidationID,
		int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER),
		fmt.Sprintf("P0-INSURANCE-NEGATIVE-%d", liquidationID), time.Now().Unix(),
	); err == nil {
		t.Fatal("negative insurance fund magnitude was not rejected")
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

func assertP0PartialLiquidationAccounting(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	now int64,
) {
	t.Helper()
	contract := insertP0LiquidationContract(
		t, ctx, serviceCtx, "P0-ISOLATED-PARTIAL-LIQUIDATION-CALL", 156, 157, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "140", "40", now)
	creditAsset(t, ctx, assetClient, 145, "120", "P0-PARTIAL-LIQUIDATION-SEED")
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
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, position, "P0-PARTIAL-LIQUIDATION-MARGIN", "2", "100", now,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, position, lot, "100")
	riskResp, err := NewProcessRiskAccountsLogic(ctx, serviceCtx).ProcessRiskAccounts(&option.OptionTaskReq{
		TenantId: p0AssetE2ETenantID,
	})
	if err != nil {
		t.Fatalf("generate partial liquidation from risk scan: %v", err)
	}
	if riskResp == nil || riskResp.Base == nil || riskResp.Base.Code != 200 {
		t.Fatalf("unexpected partial-liquidation risk response: %+v", riskResp)
	}
	liquidation, err := serviceCtx.OptionLiquidationModel.FindOpenByPosition(
		ctx, p0AssetE2ETenantID, position.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !liquidation.Quantity.Equal(decimal.NewFromInt(1)) ||
		!liquidation.Equity.Equal(decimal.NewFromInt(40)) ||
		!liquidation.MaintenanceMargin.Equal(decimal.NewFromInt(28)) ||
		!liquidation.DeficitAmount.Equal(decimal.NewFromInt(16)) ||
		!liquidation.LiquidationFee.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("unexpected risk-selected partial liquidation: %+v", liquidation)
	}
	processP0Liquidations(t, ctx, serviceCtx)
	reserved, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
	if err != nil {
		t.Fatal(err)
	}
	reservedLot, err := serviceCtx.OptionMarginLotModel.FindOne(ctx, lot.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !reserved.PositionQty.Equal(decimal.NewFromInt(2)) ||
		!reserved.AvailableQty.Equal(decimal.NewFromInt(1)) ||
		!reserved.FrozenQty.Equal(decimal.NewFromInt(1)) ||
		!reserved.MarginAmount.Equal(decimal.NewFromInt(100)) ||
		!reservedLot.RemainingQuantity.Equal(decimal.NewFromInt(1)) ||
		!reservedLot.RemainingMargin.Equal(decimal.NewFromInt(100)) ||
		!reservedLot.PendingMargin.Equal(decimal.NewFromInt(50)) {
		t.Fatalf("partial liquidation was not reserved before funding: position=%+v lot=%+v", reserved, reservedLot)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	stored, err := serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) ||
		stored.RetryCount != 0 || !stored.CollateralAmount.Equal(decimal.NewFromInt(44)) ||
		!stored.LiquidationFee.Equal(decimal.NewFromInt(4)) || stored.TakeoverPositionId <= 0 {
		t.Fatalf("unexpected partial liquidation evidence: %+v", stored)
	}
	var instructions, success, reconciled int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidation.Id).
		Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND liquidation_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, liquidation.Id).Scan(&success, &reconciled); err != nil {
		t.Fatal(err)
	}
	remaining, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, remaining, option.PositionStatus_POSITION_STATUS_HOLDING,
		"1", "50", "28", "40", "-30", "-30", "4", "-34",
	)
	remainingLot, err := serviceCtx.OptionMarginLotModel.FindOne(ctx, lot.Id)
	if err != nil {
		t.Fatal(err)
	}
	if instructions != 5 || success != 5 || reconciled != 5 ||
		!remainingLot.RemainingQuantity.Equal(decimal.NewFromInt(1)) ||
		!remainingLot.RemainingMargin.Equal(decimal.NewFromInt(50)) ||
		!remainingLot.PendingMargin.IsZero() ||
		remainingLot.Status != int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE) {
		t.Fatalf("partial liquidation instructions/lot=%d/%d/%d %+v", instructions, success, reconciled, remainingLot)
	}
	takeover, err := serviceCtx.OptionPositionModel.FindOne(ctx, stored.TakeoverPositionId)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, takeover, option.PositionStatus_POSITION_STATUS_HOLDING,
		"1", "40", "28", "40", "0", "0", "0", "0",
	)
	takeoverLot, err := serviceCtx.OptionMarginLotModel.FindOneByTenantIdTradeId(
		ctx, p0AssetE2ETenantID, -liquidation.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !takeoverLot.RemainingQuantity.Equal(decimal.NewFromInt(1)) ||
		!takeoverLot.RemainingMargin.Equal(decimal.NewFromInt(40)) {
		t.Fatalf("unexpected partial takeover margin lot: %+v", takeoverLot)
	}
	assertWalletAmounts(t, ctx, db, 145, "76.000000000000000000", "26.000000000000000000", "50.000000000000000000")
	assertWalletAmounts(t, ctx, db, contract.InsuranceUserId, "40.000000000000000000", "0.000000000000000000", "40.000000000000000000")
	assertWalletAmounts(t, ctx, db, contract.FeeUserId, "4.000000000000000000", "4.000000000000000000", "0.000000000000000000")
	processP0Liquidations(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	var replayInstructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidation.Id).
		Scan(&replayInstructions); err != nil {
		t.Fatal(err)
	}
	if replayInstructions != instructions {
		t.Fatalf("partial liquidation replay added instructions before/after=%d/%d", instructions, replayInstructions)
	}
}

func testP0PortfolioLiquidationSequential(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		userID          int64 = 151
		insuranceUserID int64 = 149
		feeUserID       int64 = 150
	)
	now := time.Now().Unix()
	contractA := insertP0LiquidationContract(
		t, ctx, serviceCtx, "P0-PORTFOLIO-LIQUIDATION-A-CALL", insuranceUserID, feeUserID, now,
	)
	contractB := insertP0LiquidationContract(
		t, ctx, serviceCtx, "P0-PORTFOLIO-LIQUIDATION-B-CALL", insuranceUserID, feeUserID, now,
	)
	contractB.StrikePrice = decimal.NewFromInt(200)
	for _, contract := range []*models.TOptionContract{contractA, contractB} {
		// Keep the sequential scenario as the explicit full-position fallback:
		// each one-contract source has no smaller valid quantity. The dedicated
		// portfolio-partial scenario below uses two contracts and still selects one.
		contract.MinOrderQty = decimal.NewFromInt(1)
		contract.QtyStep = decimal.NewFromInt(1)
		contract.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)
		contract.LiquidationDeficitPolicy = int64(
			option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP,
		)
		if err := serviceCtx.OptionContractModel.Update(ctx, contract); err != nil {
			t.Fatalf("enable portfolio margin for sequential acceptance: %v", err)
		}
	}
	insertP0ExerciseMarket(t, ctx, serviceCtx, contractA.Id, "140", "40", now)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contractB.Id, "140", "5", now)

	config := &models.TOptionPortfolioRiskConfig{
		TenantId: p0AssetE2ETenantID, SettleCoin: "USDT", Version: 1,
		Status:               int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_PENDING),
		ModelMethod:          int64(option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1),
		InitialShockRate:     decimal.RequireFromString("0.2"),
		MaintenanceShockRate: decimal.RequireFromString("0.1"),
		ScenarioShocks:       "-1,-0.2,0,0.2,4", ConcentrationThreshold: decimal.NewFromInt(1000000),
		EffectiveFrom: now + 1, ChangeReason: "P0 sequential portfolio liquidation",
		EvidenceRef: "P0-PORTFOLIO-LIQUIDATION", CreatedBy: 9001,
		CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionPortfolioRiskConfigModel.Insert(ctx, config)
	if err != nil {
		t.Fatalf("insert approved portfolio risk config: %v", err)
	}
	config.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	config.Status = int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED)
	config.ReviewedBy = 9002
	config.ReviewReason = "P0 approved fixture"
	config.ReviewedAt = now
	if err := serviceCtx.OptionPortfolioRiskConfigModel.Update(ctx, config); err != nil {
		t.Fatalf("approve portfolio risk config fixture: %v", err)
	}
	waitP1PortfolioBoundary(t, config.EffectiveFrom)
	if _, err := db.ExecContext(ctx, `UPDATE t_asset_platform_account
		SET available_amount=100,update_times=?
		WHERE tenant_id=? AND account_type='INSURANCE_FUND' AND coin='USDT'`,
		now*1000, p0AssetE2ETenantID); err != nil {
		t.Fatalf("top up portfolio liquidation insurance fund: %v", err)
	}
	testP0PortfolioPartialLiquidation(
		t, ctx, db, assetClient, serviceCtx, contractA, config.Id, now,
	)
	creditAsset(t, ctx, assetClient, userID, "500", "P0-PORTFOLIO-LIQUIDATION-SEED")

	positionA := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: 8043,
		ContractId: contractA.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(40),
		MarginAmount: decimal.NewFromInt(250), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now, UpdateTimes: now,
	})
	positionB := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: 8044,
		ContractId: contractB.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(2),
		MarkPrice: decimal.NewFromInt(5), PositionValue: decimal.NewFromInt(5),
		MarginAmount: decimal.NewFromInt(250), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now, UpdateTimes: now,
	})
	for i, position := range []*models.TOptionPosition{positionA, positionB} {
		lot := insertP0ExerciseMarginLot(
			t, ctx, serviceCtx, position,
			fmt.Sprintf("P0-PORTFOLIO-LIQUIDATION-MARGIN-%d", i+1), "1", "250", now,
		)
		freezeP0ExerciseMargin(t, ctx, assetClient, position, lot, "250")
	}

	refreshP0PortfolioRiskUser(t, ctx, serviceCtx, userID)
	first := requireP0PortfolioLiquidations(t, ctx, serviceCtx, userID, 1)[0]
	assertP0PortfolioLiquidationSnapshot(t, first, config.Id, 1, "500")
	if first.PositionId != positionA.Id {
		t.Fatalf("portfolio risk-relief selector chose position=%d want=%d", first.PositionId, positionA.Id)
	}
	// A normal market update after trigger must not starve liquidation. The
	// executor re-evaluates the residual portfolio and retains the larger of
	// trigger-time and current initial requirements.
	marketB, err := serviceCtx.OptionMarketModel.FindOneByTenantIdContractId(
		ctx, p0AssetE2ETenantID, contractB.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	marketB.MarkPrice = decimal.RequireFromString("4.9")
	marketB.LastPrice = marketB.MarkPrice
	marketB.TheoreticalPrice = marketB.MarkPrice
	marketB.MarkSnapshotTime = time.Now().Unix()
	marketB.SnapshotTime = marketB.MarkSnapshotTime
	marketB.UpdateTimes = marketB.MarkSnapshotTime
	if err := serviceCtx.OptionMarketModel.Update(ctx, marketB); err != nil {
		t.Fatalf("update residual portfolio market before liquidation: %v", err)
	}
	processP0Liquidations(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	first = requireP0PortfolioLiquidations(t, ctx, serviceCtx, userID, 1)[0]
	if first.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) {
		t.Fatalf("first portfolio liquidation not done: %+v", first)
	}
	plannedCollateralUse := first.PortfolioCollateralBefore.Sub(first.PortfolioCollateralAfter)
	if !first.CollateralAmount.LessThan(plannedCollateralUse) {
		t.Fatalf("current residual requirement did not conservatively reduce collateral use actual/planned=%s/%s",
			first.CollateralAmount, plannedCollateralUse)
	}
	// A second liquidation must not exist until the entire residual wallet has
	// been recalculated after the first takeover completed.
	refreshP0PortfolioRiskUser(t, ctx, serviceCtx, userID)
	items := requireP0PortfolioLiquidations(t, ctx, serviceCtx, userID, 2)
	second := items[0]
	if second.Id == first.Id {
		second = items[1]
	}
	if second.PositionId == first.PositionId {
		t.Fatalf("sequential portfolio liquidation selected the same position twice: %+v %+v", first, second)
	}
	assertP0PortfolioLiquidationSnapshot(t, second, config.Id, 1, second.PortfolioCollateralBefore.String())
	if !second.PortfolioCollateralBefore.GreaterThan(first.PortfolioCollateralAfter) {
		t.Fatalf("second risk scan did not preserve the higher current collateral floor firstPlanned/secondActual=%s/%s",
			first.PortfolioCollateralAfter, second.PortfolioCollateralBefore)
	}
	processP0Liquidations(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	items = requireP0PortfolioLiquidations(t, ctx, serviceCtx, userID, 2)
	for _, liquidation := range items {
		if liquidation.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) {
			t.Fatalf("portfolio liquidation not done: %+v", liquidation)
		}
		assertP0PortfolioLiquidationInstructions(t, ctx, db, liquidation.Id)
	}
	for _, positionID := range []int64{positionA.Id, positionB.Id} {
		position, err := serviceCtx.OptionPositionModel.FindOne(ctx, positionID)
		if err != nil {
			t.Fatal(err)
		}
		if position.Status != int64(option.PositionStatus_POSITION_STATUS_CLOSED) ||
			!position.PositionQty.IsZero() {
			t.Fatalf("portfolio liquidation source position not closed: %+v", position)
		}
	}
	// With no residual user positions, the next risk pass releases the exact
	// collateral remainder rather than silently consuming it.
	if err := NewProcessRiskAccountsLogic(ctx, serviceCtx).rebalancePortfolioCollateral(
		&optionRiskGroup{key: optionRiskKey{
			tenantID: p0AssetE2ETenantID, userID: userID, accountID: 0, coin: "USDT",
		}}, decimal.Zero, time.Now().Unix(),
	); err != nil {
		t.Fatalf("release residual portfolio collateral: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	var total, available, frozen decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT total_amount,available_amount,frozen_amount
		FROM t_user_asset WHERE tenant_id=? AND user_id=? AND wallet_type=? AND coin='USDT'`,
		p0AssetE2ETenantID, userID, int64(common.WalletType_WALLET_TYPE_OPTION)).
		Scan(&total, &available, &frozen); err != nil {
		t.Fatal(err)
	}
	if !total.Equal(second.PortfolioCollateralAfter) || !available.Equal(total) || !frozen.IsZero() {
		t.Fatalf("residual portfolio collateral not released total/available/frozen=%s/%s/%s want=%s/%s/0",
			total, available, frozen, second.PortfolioCollateralAfter, second.PortfolioCollateralAfter)
	}
	assertP0RecoveredPortfolioLiquidationCanceled(
		t, ctx, db, assetClient, serviceCtx, contractA, config.Id,
	)
}

func testP0PortfolioPartialLiquidation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	configID int64,
	now int64,
) {
	t.Helper()
	const userID int64 = 158
	creditAsset(t, ctx, assetClient, userID, "700", "P0-PORTFOLIO-PARTIAL-SEED")
	position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: 8051,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		MarginAmount: decimal.NewFromInt(700), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now, UpdateTimes: now,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, position, "P0-PORTFOLIO-PARTIAL-MARGIN", "2", "700", now,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, position, lot, "700")

	refreshP0PortfolioRiskUser(t, ctx, serviceCtx, userID)
	liquidation := requireP0PortfolioLiquidations(t, ctx, serviceCtx, userID, 1)[0]
	assertP0PortfolioLiquidationSnapshot(t, liquidation, configID, 1, "700")
	if !liquidation.Quantity.Equal(decimal.NewFromInt(1)) ||
		!liquidation.LiquidationFee.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("portfolio partial selector quantity/fee=%s/%s want=1/4: %+v",
			liquidation.Quantity, liquidation.LiquidationFee, liquidation)
	}
	processP0Liquidations(t, ctx, serviceCtx)
	reserved, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !reserved.PositionQty.Equal(decimal.NewFromInt(2)) ||
		!reserved.AvailableQty.Equal(decimal.NewFromInt(1)) ||
		!reserved.FrozenQty.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("portfolio partial quantity was not reserved before funding: %+v", reserved)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	liquidation, err = serviceCtx.OptionLiquidationModel.FindOne(ctx, liquidation.Id)
	if err != nil {
		t.Fatal(err)
	}
	if liquidation.Status != int64(option.LiquidationStatus_LIQUIDATION_STATUS_DONE) ||
		!liquidation.CollateralAmount.Equal(decimal.NewFromInt(44)) ||
		liquidation.TakeoverPositionId <= 0 {
		t.Fatalf("portfolio partial liquidation did not complete: %+v", liquidation)
	}
	var instructions, success, reconciled int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND liquidation_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, liquidation.Id,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if instructions != 4 || success != 4 || reconciled != 4 {
		t.Fatalf("portfolio partial instructions=%d/%d/%d want=4/4/4",
			instructions, success, reconciled)
	}

	// Recalculate the residual wallet. It is healthy after one contract, so no
	// second liquidation may be created and excess pool collateral is released.
	refreshP0PortfolioRiskUser(t, ctx, serviceCtx, userID)
	processAssetInstructions(t, ctx, serviceCtx)
	if items := requireP0PortfolioLiquidations(t, ctx, serviceCtx, userID, 1); items[0].Id != liquidation.Id {
		t.Fatalf("portfolio partial recovery created a replacement liquidation: %+v", items)
	}
	source, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, source, option.PositionStatus_POSITION_STATUS_HOLDING,
		"1", "560", "0", "40", "-30", "-30", "4", "-34",
	)
	takeover, err := serviceCtx.OptionPositionModel.FindOne(ctx, liquidation.TakeoverPositionId)
	if err != nil {
		t.Fatal(err)
	}
	assertP0LiquidationPosition(
		t, takeover, option.PositionStatus_POSITION_STATUS_HOLDING,
		"1", "40", "0", "40", "0", "0", "0", "0",
	)
	remainingLot, err := serviceCtx.OptionMarginLotModel.FindOne(ctx, lot.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !remainingLot.RemainingMargin.Equal(decimal.NewFromInt(560)) ||
		!remainingLot.PendingMargin.IsZero() {
		t.Fatalf("portfolio partial residual collateral lot: %+v", remainingLot)
	}
	assertWalletAmounts(t, ctx, db, userID, "656.000000000000000000", "96.000000000000000000", "560.000000000000000000")
}

func assertP0RecoveredPortfolioLiquidationCanceled(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	configID int64,
) {
	t.Helper()
	const recoveredUserID int64 = 155
	now := time.Now().Unix()
	creditAsset(t, ctx, assetClient, recoveredUserID, "600", "P0-PORTFOLIO-RECOVERY-SEED")
	position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: recoveredUserID, AccountId: 8050,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(40),
		MarginAmount: decimal.NewFromInt(600), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now, UpdateTimes: now,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, position, "P0-PORTFOLIO-RECOVERY-MARGIN", "1", "600", now,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, position, lot, "600")
	refreshP0PortfolioRiskUser(t, ctx, serviceCtx, recoveredUserID)
	liquidation := requireP0PortfolioLiquidations(t, ctx, serviceCtx, recoveredUserID, 1)[0]
	assertP0PortfolioLiquidationSnapshot(t, liquidation, configID, 1, "600")
	creditAsset(t, ctx, assetClient, recoveredUserID, "100", "P0-PORTFOLIO-RECOVERY-TOPUP")
	probe := NewProcessLiquidationsLogic(ctx, serviceCtx)
	waiting, err := probe.cancelRiskOrders(liquidation)
	if err != nil || waiting {
		t.Fatalf("recovered portfolio liquidation cancellation barrier orders waiting/error=%t/%v", waiting, err)
	}
	incomplete, err := probe.hasLiquidationBarrier(liquidation)
	if err != nil || incomplete {
		t.Fatalf("recovered portfolio liquidation cancellation barrier incomplete/error=%t/%v", incomplete, err)
	}
	wantError := "portfolio liquidation snapshot is stale: wallet recovered above current maintenance requirement"
	if _, err := probe.buildLiquidationPlan(liquidation); !errors.Is(err, errPortfolioLiquidationSnapshotStale) ||
		err.Error() != wantError {
		t.Fatalf("recovered portfolio liquidation pre-execution recheck error=%v want=%q", err, wantError)
	}
	processP0Liquidations(t, ctx, serviceCtx)
	var statusValue, retryCount, instructions int64
	var lastError string
	if err := db.QueryRowContext(ctx, `SELECT status,retry_count,last_error_msg
		FROM t_option_liquidation WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, liquidation.Id).Scan(&statusValue, &retryCount, &lastError); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND liquidation_id=?`, p0AssetE2ETenantID, liquidation.Id).
		Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.LiquidationStatus_LIQUIDATION_STATUS_CANCELED) ||
		retryCount != 0 || instructions != 0 || lastError != wantError {
		t.Fatalf("recovered portfolio liquidation status/retry/instructions/error=%d/%d/%d/%q",
			statusValue, retryCount, instructions, lastError)
	}
}

func refreshP0PortfolioRiskUser(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, userID int64,
) {
	t.Helper()
	logic := NewProcessRiskAccountsLogic(ctx, serviceCtx)
	groups, err := logic.collectRiskGroups(p0AssetE2ETenantID)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range groups {
		if group.key.userID == userID && group.key.coin == "USDT" {
			if err := logic.refreshRiskGroup(group); err != nil {
				t.Fatalf("refresh portfolio risk user: %v", err)
			}
			return
		}
	}
	t.Fatalf("portfolio risk group not found for user %d", userID)
}

func requireP0PortfolioLiquidations(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, userID int64, want int,
) []*models.TOptionLiquidation {
	t.Helper()
	items, _, err := serviceCtx.OptionLiquidationModel.FindPage(ctx, models.OptionLiquidationPageFilter{
		TenantId: p0AssetE2ETenantID, UserId: userID,
	}, 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != want {
		t.Fatalf("portfolio liquidation count=%d want=%d items=%+v", len(items), want, items)
	}
	return items
}

func assertP0PortfolioLiquidationSnapshot(
	t *testing.T, item *models.TOptionLiquidation, configID, configVersion int64, collateralBefore string,
) {
	t.Helper()
	if item.LiquidationScope != int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET) ||
		item.AccountId != 0 || item.PortfolioRiskConfigId != configID ||
		item.PortfolioRiskConfigVersion != configVersion ||
		!item.PortfolioMaintenanceBefore.GreaterThan(item.PortfolioMaintenanceAfter) ||
		item.PortfolioCollateralAfter.LessThan(item.PortfolioInitialAfter) ||
		!item.PortfolioCollateralBefore.Equal(decimal.RequireFromString(collateralBefore)) {
		t.Fatalf("invalid portfolio liquidation snapshot: %+v", item)
	}
}

func assertP0PortfolioLiquidationInstructions(
	t *testing.T, ctx context.Context, db *sql.DB, liquidationID int64,
) {
	t.Helper()
	var total, success, matched int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND liquidation_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, liquidationID).Scan(&total, &success, &matched); err != nil {
		t.Fatal(err)
	}
	if total < 3 || success != total || matched != total {
		t.Fatalf("portfolio liquidation instructions total/success/matched=%d/%d/%d", total, success, matched)
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
