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

func testP0MarketAndPostOnlyOrders(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P0_MARKET_POST_ONLY_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)

	marketContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-MARKET-PROTECTION-CALL", calendarCode, 183, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, marketContract.Id, "100", "10", now)
	testP0MarketOrderReleasesUnusedReservation(
		t, ctx, db, assetClient, serviceCtx, marketContract,
	)

	postOnlyContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-POST-ONLY-CALL", calendarCode, 187, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, postOnlyContract.Id, "100", "10", now)
	testP0PostOnlyRestAndWouldTake(
		t, ctx, db, assetClient, serviceCtx, postOnlyContract,
	)
}

func testP0MarketOrderReleasesUnusedReservation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		sellerUserID int64 = 181
		buyerUserID  int64 = 182
		feeUserID    int64 = 183
	)
	creditAsset(t, ctx, assetClient, sellerUserID, "100", "P0-MARKET-SELLER-SEED")
	creditAsset(t, ctx, assetClient, buyerUserID, "100", "P0-MARKET-BUYER-SEED")
	seller := placeP0Order(t, ctx, serviceCtx, sellerUserID, &option.PlaceOrderReq{
		AccountId: 8063, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-MARKET-SELLER",
	})
	processAssetInstructions(t, ctx, serviceCtx)

	buyerReq := &option.PlaceOrderReq{
		AccountId: 7063, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_MARKET, Qty: "1",
		ProtectionPrice: "10", MaxTurnover: "12",
		ClientOrderId: "P0-MARKET-BUYER",
	}
	buyer := placeP0Order(t, ctx, serviceCtx, buyerUserID, buyerReq)
	for i := 0; i < 6; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	processP0TradeEvents(t, ctx, serviceCtx)
	processP0OrderRiskAccountsForAccounts(
		t, ctx, serviceCtx, contract, buyerUserID, 7063, sellerUserID, 8063,
	)
	assertP0MarketEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyer.Data.OrderId, seller.Data.OrderId,
		buyerUserID, sellerUserID, feeUserID,
	)

	replay, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, buyerUserID), serviceCtx,
	).PlaceOrder(buyerReq)
	if err != nil || replay == nil || replay.Base == nil || replay.Base.Code == 200 ||
		replay.Data == nil || replay.Data.OrderId != buyer.Data.OrderId {
		t.Fatalf("unexpected MARKET replay resp=%+v err=%v", replay, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)
	assertP0MarketEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyer.Data.OrderId, seller.Data.OrderId,
		buyerUserID, sellerUserID, feeUserID,
	)
}

func assertP0MarketEvidence(
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
	if buyerOrder.OrderType != int64(option.OrderType_ORDER_TYPE_MARKET) ||
		buyerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_FILLED) ||
		!buyerOrder.Price.Equal(decimal.NewFromInt(10)) || !buyerOrder.MarginAmount.IsZero() ||
		sellerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_FILLED) ||
		!sellerOrder.MarginAmount.IsZero() {
		t.Fatalf("unexpected MARKET orders buyer=%+v seller=%+v", buyerOrder, sellerOrder)
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
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_margin_lot WHERE tenant_id=? AND contract_id=?
		 AND initial_margin=50 AND remaining_margin=50)`,
		p0AssetE2ETenantID, contractID, p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, contractID, p0AssetE2ETenantID, contractID,
	).Scan(&outbox, &inbox, &positions, &marginLots); err != nil {
		t.Fatal(err)
	}
	var releaseAmount decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT amount FROM t_option_asset_instruction
		WHERE tenant_id=? AND instruction_no=?`, p0AssetE2ETenantID,
		buyerOrder.OrderNo+"-RELEASE-REMAINDER",
	).Scan(&releaseAmount); err != nil {
		t.Fatal(err)
	}
	if trades != 1 || instructions != 6 || success != 6 || reconciled != 6 || flows != 6 ||
		outbox != 1 || inbox != 1 || positions != 2 || marginLots != 1 ||
		!releaseAmount.Equal(decimal.RequireFromString("1.6")) {
		t.Fatalf("MARKET evidence trades/instructions/success/reconciled/flows/events/positions/lots/release=%d/%d/%d/%d/%d/%d/%d/%d/%d/%s",
			trades, instructions, success, reconciled, flows, outbox, inbox, positions, marginLots, releaseAmount)
	}
	assertWalletAmounts(t, ctx, db, buyerUserID,
		"89.600000000000000000", "89.600000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerUserID,
		"109.800000000000000000", "59.800000000000000000", "50.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID,
		"0.600000000000000000", "0.600000000000000000", "0.000000000000000000")
}

func testP0PostOnlyRestAndWouldTake(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		sellerUserID        int64 = 184
		crossingBuyerUserID int64 = 185
		restingBuyerUserID  int64 = 186
	)
	creditAsset(t, ctx, assetClient, sellerUserID, "100", "P0-POST-ONLY-SELLER-SEED")
	creditAsset(t, ctx, assetClient, crossingBuyerUserID, "100", "P0-POST-ONLY-CROSS-SEED")
	creditAsset(t, ctx, assetClient, restingBuyerUserID, "100", "P0-POST-ONLY-REST-SEED")
	seller := placeP0Order(t, ctx, serviceCtx, sellerUserID, &option.PlaceOrderReq{
		AccountId: 8064, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-POST-ONLY-SELLER",
	})
	processAssetInstructions(t, ctx, serviceCtx)

	crossingReq := &option.PlaceOrderReq{
		AccountId: 7064, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_POST_ONLY, Price: "10", Qty: "1",
		ClientOrderId: "P0-POST-ONLY-WOULD-TAKE",
	}
	crossing := placeP0Order(t, ctx, serviceCtx, crossingBuyerUserID, crossingReq)
	for i := 0; i < 3; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}

	resting := placeP0Order(t, ctx, serviceCtx, restingBuyerUserID, &option.PlaceOrderReq{
		AccountId: 7065, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_POST_ONLY, Price: "8", Qty: "1",
		ClientOrderId: "P0-POST-ONLY-RESTING",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	restingOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, resting.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if restingOrder.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) ||
		!restingOrder.MarginAmount.Equal(decimal.RequireFromString("8.32")) {
		t.Fatalf("POST_ONLY non-crossing order did not rest: %+v", restingOrder)
	}
	assertP0UserCancelOK(t, ctx, serviceCtx, restingBuyerUserID, 7065, resting.Data.OrderId)
	assertP0UserCancelOK(t, ctx, serviceCtx, sellerUserID, 8064, seller.Data.OrderId)
	for i := 0; i < 3; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0PostOnlyEvidence(
		t, ctx, db, serviceCtx, contract.Id, seller.Data.OrderId,
		crossing.Data.OrderId, resting.Data.OrderId,
		sellerUserID, crossingBuyerUserID, restingBuyerUserID,
	)

	replay, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, crossingBuyerUserID), serviceCtx,
	).PlaceOrder(crossingReq)
	if err != nil || replay == nil || replay.Base == nil || replay.Base.Code == 200 ||
		replay.Data == nil || replay.Data.OrderId != crossing.Data.OrderId {
		t.Fatalf("unexpected POST_ONLY replay resp=%+v err=%v", replay, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0PostOnlyEvidence(
		t, ctx, db, serviceCtx, contract.Id, seller.Data.OrderId,
		crossing.Data.OrderId, resting.Data.OrderId,
		sellerUserID, crossingBuyerUserID, restingBuyerUserID,
	)
}

func assertP0PostOnlyEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID, sellerOrderID, crossingOrderID, restingOrderID,
	sellerUserID, crossingBuyerUserID, restingBuyerUserID int64,
) {
	t.Helper()
	crossing, err := serviceCtx.OptionOrderModel.FindOne(ctx, crossingOrderID)
	if err != nil {
		t.Fatal(err)
	}
	resting, err := serviceCtx.OptionOrderModel.FindOne(ctx, restingOrderID)
	if err != nil {
		t.Fatal(err)
	}
	seller, err := serviceCtx.OptionOrderModel.FindOne(ctx, sellerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	if crossing.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		crossing.CancelReason != "POST_ONLY_WOULD_TAKE" || !crossing.FilledQty.IsZero() ||
		!crossing.MarginAmount.IsZero() ||
		resting.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		resting.CancelReason != "USER_CANCEL" || !resting.MarginAmount.IsZero() ||
		seller.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		seller.CancelReason != "USER_CANCEL" || !seller.MarginAmount.IsZero() {
		t.Fatalf("unexpected POST_ONLY terminal orders cross=%+v rest=%+v seller=%+v",
			crossing, resting, seller)
	}
	var orders, clientKeys, trades, instructions, success, reconciled, flows, positions, outbox int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_order
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_client_order_key key_item
		JOIN t_option_order o ON o.tenant_id=key_item.tenant_id AND o.id=key_item.order_id
		WHERE o.tenant_id=? AND o.contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&clientKeys); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_trade WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_outbox WHERE tenant_id=? AND contract_id=?)`,
		p0AssetE2ETenantID, contractID, p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, contractID,
	).Scan(&trades, &positions, &outbox); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=3),0),
		COALESCE(SUM(reconciliation_status=2),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (?,?,?)`,
		p0AssetE2ETenantID, sellerOrderID, crossingOrderID, restingOrderID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction JOIN t_asset_flow flow
		 ON flow.tenant_id=instruction.tenant_id
		AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id IN (?,?,?)`,
		p0AssetE2ETenantID, sellerOrderID, crossingOrderID, restingOrderID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if orders != 3 || clientKeys != 3 || trades != 0 || instructions != 6 || success != 6 ||
		reconciled != 6 || flows != 6 || positions != 0 || outbox != 0 {
		t.Fatalf("POST_ONLY evidence orders/keys/trades/instructions/success/reconciled/flows/positions/outbox=%d/%d/%d/%d/%d/%d/%d/%d/%d",
			orders, clientKeys, trades, instructions, success, reconciled, flows, positions, outbox)
	}
	assertWalletAmounts(t, ctx, db, sellerUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, crossingBuyerUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, restingBuyerUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
}
