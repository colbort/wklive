package tasklogic

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func testP0OrderCancellationAndImmediateTypes(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P0_ORDER_CANCEL_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)

	cancelContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-USER-CANCEL-CALL", calendarCode, 179, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, cancelContract.Id, "100", "10", now)
	testP0CancelBeforeAndAfterFunding(
		t, ctx, db, assetClient, serviceCtx, cancelContract,
	)

	iocContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-IOC-PARTIAL-CALL", calendarCode, 175, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, iocContract.Id, "100", "10", now)
	testP0IOCPartialFillReleasesRemainder(
		t, ctx, db, assetClient, serviceCtx, iocContract,
	)

	fokContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-FOK-INSUFFICIENT-CALL", calendarCode, 178, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, fokContract.Id, "100", "10", now)
	testP0FOKInsufficientLiquidityIsAllOrNone(
		t, ctx, db, assetClient, serviceCtx, fokContract,
	)
}

func testP0CancelBeforeAndAfterFunding(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		beforeUserID int64 = 171
		afterUserID  int64 = 172
		accountID    int64 = 7060
	)
	creditAsset(t, ctx, assetClient, beforeUserID, "100", "P0-CANCEL-BEFORE-FUNDING-SEED")
	creditAsset(t, ctx, assetClient, afterUserID, "100", "P0-CANCEL-AFTER-FUNDING-SEED")

	before := placeP0Order(t, ctx, serviceCtx, beforeUserID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-CANCEL-BEFORE-FUNDING",
	})
	assertP0UserCancelOK(t, ctx, serviceCtx, beforeUserID, accountID, before.Data.OrderId)
	processAssetInstructions(t, ctx, serviceCtx)
	beforeOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, before.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if beforeOrder.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		beforeOrder.CancelReason != "USER_CANCEL" || !beforeOrder.MarginAmount.IsZero() {
		t.Fatalf("unexpected pre-funding cancel order: %+v", beforeOrder)
	}
	var beforeInstructions, canceledInstructions, beforeFlows int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=?),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED),
		p0AssetE2ETenantID, beforeOrder.Id,
	).Scan(&beforeInstructions, &canceledInstructions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		 ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.order_id=?`,
		p0AssetE2ETenantID, beforeOrder.Id,
	).Scan(&beforeFlows); err != nil {
		t.Fatal(err)
	}
	if beforeInstructions != 1 || canceledInstructions != 1 || beforeFlows != 0 {
		t.Fatalf("pre-funding cancel instructions/canceled/flows=%d/%d/%d",
			beforeInstructions, canceledInstructions, beforeFlows)
	}
	assertWalletAmounts(t, ctx, db, beforeUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertP0UserCancelRejected(t, ctx, serviceCtx, beforeUserID, accountID, beforeOrder.Id)

	after := placeP0Order(t, ctx, serviceCtx, afterUserID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-CANCEL-AFTER-FUNDING",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	afterOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, after.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if afterOrder.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) {
		t.Fatalf("expected funded order on book before user cancel: %+v", afterOrder)
	}
	assertWalletAmounts(t, ctx, db, afterUserID,
		"100.000000000000000000", "89.600000000000000000", "10.400000000000000000")
	assertP0UserCancelOK(t, ctx, serviceCtx, afterUserID, accountID, afterOrder.Id)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0FundedCancelEvidence(t, ctx, db, serviceCtx, afterOrder.Id, afterUserID)
	assertP0UserCancelRejected(t, ctx, serviceCtx, afterUserID, accountID, afterOrder.Id)
	assertP0FundedCancelEvidence(t, ctx, db, serviceCtx, afterOrder.Id, afterUserID)
}

func assertP0FundedCancelEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	orderID, userID int64,
) {
	t.Helper()
	order, err := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		order.CancelReason != "USER_CANCEL" || !order.MarginAmount.IsZero() {
		t.Fatalf("unexpected funded cancel order: %+v", order)
	}
	var instructions, success, reconciled, flows int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=3),0),
		COALESCE(SUM(reconciliation_status=2),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id=?`,
		p0AssetE2ETenantID, orderID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction JOIN t_asset_flow flow
		 ON flow.tenant_id=instruction.tenant_id
		AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id=?`,
		p0AssetE2ETenantID, orderID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if instructions != 2 || success != 2 || reconciled != 2 || flows != 2 {
		t.Fatalf("funded cancel instructions/success/reconciled/flows=%d/%d/%d/%d",
			instructions, success, reconciled, flows)
	}
	assertWalletAmounts(t, ctx, db, userID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
}

func testP0IOCPartialFillReleasesRemainder(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		sellerUserID int64 = 173
		buyerUserID  int64 = 174
		feeUserID    int64 = 175
	)
	creditAsset(t, ctx, assetClient, sellerUserID, "100", "P0-IOC-SELLER-SEED")
	creditAsset(t, ctx, assetClient, buyerUserID, "100", "P0-IOC-BUYER-SEED")
	seller := placeP0Order(t, ctx, serviceCtx, sellerUserID, &option.PlaceOrderReq{
		AccountId: 8061, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-IOC-SELLER",
	})
	processAssetInstructions(t, ctx, serviceCtx)

	buyerReq := &option.PlaceOrderReq{
		AccountId: 7061, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_IOC, Price: "10", Qty: "2",
		ClientOrderId: "P0-IOC-BUYER-PARTIAL",
	}
	buyer := placeP0Order(t, ctx, serviceCtx, buyerUserID, buyerReq)
	for i := 0; i < 6; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	processP0TradeEvents(t, ctx, serviceCtx)
	processP0OrderRiskAccountsForAccounts(
		t, ctx, serviceCtx, contract, buyerUserID, 7061, sellerUserID, 8061,
	)
	assertP0IOCPartialEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyer.Data.OrderId, seller.Data.OrderId,
		buyerUserID, sellerUserID, feeUserID,
	)

	replay, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, buyerUserID), serviceCtx,
	).PlaceOrder(buyerReq)
	if err != nil || replay == nil || replay.Base == nil || replay.Base.Code == 200 ||
		replay.Data == nil || replay.Data.OrderId != buyer.Data.OrderId {
		t.Fatalf("unexpected IOC replay resp=%+v err=%v", replay, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)
	assertP0IOCPartialEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyer.Data.OrderId, seller.Data.OrderId,
		buyerUserID, sellerUserID, feeUserID,
	)
}

func assertP0IOCPartialEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID, buyerOrderID, sellerOrderID, buyerUserID, sellerUserID, feeUserID int64,
) {
	t.Helper()
	buyerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, buyerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	sellerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, sellerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	if buyerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		buyerOrder.CancelReason != "IMMEDIATE_REMAINDER_CANCELED" ||
		!buyerOrder.FilledQty.Equal(decimal.NewFromInt(1)) ||
		!buyerOrder.UnfilledQty.Equal(decimal.NewFromInt(1)) || !buyerOrder.MarginAmount.IsZero() ||
		sellerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_FILLED) ||
		!sellerOrder.MarginAmount.IsZero() {
		t.Fatalf("unexpected IOC orders buyer=%+v seller=%+v", buyerOrder, sellerOrder)
	}
	var trades, instructions, success, reconciled, flows, outbox, inbox, positions, marginLots int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trade
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&trades); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=3),0),
		COALESCE(SUM(reconciliation_status=2),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND
		(order_id IN (?,?) OR trade_id IN (SELECT id FROM t_option_trade WHERE tenant_id=? AND contract_id=?))`,
		p0AssetE2ETenantID, buyerOrderID, sellerOrderID, p0AssetE2ETenantID, contractID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction JOIN t_asset_flow flow
		 ON flow.tenant_id=instruction.tenant_id
		AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND
		(instruction.order_id IN (?,?) OR instruction.trade_id IN
		 (SELECT id FROM t_option_trade WHERE tenant_id=? AND contract_id=?))`,
		p0AssetE2ETenantID, buyerOrderID, sellerOrderID, p0AssetE2ETenantID, contractID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_outbox WHERE tenant_id=? AND contract_id=? AND status=3),
		(SELECT COUNT(*) FROM t_option_inbox WHERE tenant_id=? AND contract_id=? AND status=2),
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND contract_id=?)`,
		p0AssetE2ETenantID, contractID, p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, contractID,
	).Scan(&outbox, &inbox, &positions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_margin_lot
		WHERE tenant_id=? AND contract_id=? AND initial_margin=50 AND remaining_margin=50`,
		p0AssetE2ETenantID, contractID,
	).Scan(&marginLots); err != nil {
		t.Fatal(err)
	}
	var releaseAmount decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT amount FROM t_option_asset_instruction
		WHERE tenant_id=? AND instruction_no=?`, p0AssetE2ETenantID,
		buyerOrder.OrderNo+"-IMMEDIATE-RELEASE",
	).Scan(&releaseAmount); err != nil {
		t.Fatal(err)
	}
	if trades != 1 || instructions != 6 || success != 6 || reconciled != 6 || flows != 6 ||
		outbox != 1 || inbox != 1 || positions != 2 || marginLots != 1 ||
		!releaseAmount.Equal(decimal.RequireFromString("10.4")) {
		t.Fatalf("IOC evidence trades/instructions/success/reconciled/flows/events/positions/lots/release=%d/%d/%d/%d/%d/%d/%d/%d/%d/%s",
			trades, instructions, success, reconciled, flows, outbox, inbox, positions, marginLots, releaseAmount)
	}
	assertWalletAmounts(t, ctx, db, buyerUserID,
		"89.600000000000000000", "89.600000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerUserID,
		"109.800000000000000000", "59.800000000000000000", "50.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID,
		"0.600000000000000000", "0.600000000000000000", "0.000000000000000000")
}

func testP0FOKInsufficientLiquidityIsAllOrNone(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		sellerUserID int64 = 176
		buyerUserID  int64 = 177
	)
	creditAsset(t, ctx, assetClient, sellerUserID, "100", "P0-FOK-SELLER-SEED")
	creditAsset(t, ctx, assetClient, buyerUserID, "100", "P0-FOK-BUYER-SEED")
	seller := placeP0Order(t, ctx, serviceCtx, sellerUserID, &option.PlaceOrderReq{
		AccountId: 8062, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-FOK-SELLER-ONE",
	})
	processAssetInstructions(t, ctx, serviceCtx)

	buyerReq := &option.PlaceOrderReq{
		AccountId: 7062, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_FOK, Price: "10", Qty: "2",
		ClientOrderId: "P0-FOK-BUYER-TWO",
	}
	buyer := placeP0Order(t, ctx, serviceCtx, buyerUserID, buyerReq)
	for i := 0; i < 4; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0UserCancelOK(t, ctx, serviceCtx, sellerUserID, 8062, seller.Data.OrderId)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0FOKEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyer.Data.OrderId, seller.Data.OrderId,
		buyerUserID, sellerUserID,
	)

	replay, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, buyerUserID), serviceCtx,
	).PlaceOrder(buyerReq)
	if err != nil || replay == nil || replay.Base == nil || replay.Base.Code == 200 ||
		replay.Data == nil || replay.Data.OrderId != buyer.Data.OrderId {
		t.Fatalf("unexpected FOK replay resp=%+v err=%v", replay, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0FOKEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyer.Data.OrderId, seller.Data.OrderId,
		buyerUserID, sellerUserID,
	)
}

func assertP0FOKEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID, buyerOrderID, sellerOrderID, buyerUserID, sellerUserID int64,
) {
	t.Helper()
	buyerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, buyerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	sellerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, sellerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	if buyerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		buyerOrder.CancelReason != "FOK_NOT_FILLED" || !buyerOrder.FilledQty.IsZero() ||
		!buyerOrder.UnfilledQty.Equal(decimal.NewFromInt(2)) || !buyerOrder.MarginAmount.IsZero() ||
		sellerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		sellerOrder.CancelReason != "USER_CANCEL" || !sellerOrder.MarginAmount.IsZero() {
		t.Fatalf("unexpected FOK orders buyer=%+v seller=%+v", buyerOrder, sellerOrder)
	}
	var trades, instructions, success, reconciled, flows, positions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trade
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&trades); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=3),0),
		COALESCE(SUM(reconciliation_status=2),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (?,?)`,
		p0AssetE2ETenantID, buyerOrderID, sellerOrderID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction JOIN t_asset_flow flow
		 ON flow.tenant_id=instruction.tenant_id
		AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id IN (?,?)`,
		p0AssetE2ETenantID, buyerOrderID, sellerOrderID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_position
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&positions); err != nil {
		t.Fatal(err)
	}
	if trades != 0 || instructions != 4 || success != 4 || reconciled != 4 || flows != 4 || positions != 0 {
		t.Fatalf("FOK evidence trades/instructions/success/reconciled/flows/positions=%d/%d/%d/%d/%d/%d",
			trades, instructions, success, reconciled, flows, positions)
	}
	assertWalletAmounts(t, ctx, db, buyerUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
}

func assertP0UserCancelOK(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID, accountID, orderID int64,
) {
	t.Helper()
	resp, err := applogic.NewCancelOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).CancelOrder(&option.CancelOrderReq{AccountId: accountID, OrderId: orderID})
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("cancel order user=%d order=%d resp=%+v err=%v", userID, orderID, resp, err)
	}
}

func assertP0UserCancelRejected(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID, accountID, orderID int64,
) {
	t.Helper()
	resp, err := applogic.NewCancelOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).CancelOrder(&option.CancelOrderReq{AccountId: accountID, OrderId: orderID})
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code == 200 {
		t.Fatalf("expected cancel replay rejection user=%d order=%d resp=%+v err=%v",
			userID, orderID, resp, err)
	}
}
