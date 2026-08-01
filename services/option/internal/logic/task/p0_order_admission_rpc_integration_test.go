package tasklogic

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/metadata"
)

func testP0OrderAdmissionToRiskAccounting(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		buyerUserID  int64 = 161
		sellerUserID int64 = 162
		feeUserID    int64 = 163
	)
	now := time.Now().Unix()
	calendarCode := "P0_ORDER_ADMISSION_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	contract := insertP0OrderAdmissionContract(t, ctx, serviceCtx, calendarCode, feeUserID, now)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, buyerUserID, "100", "P0-ORDER-ADMISSION-BUYER-SEED")
	creditAsset(t, ctx, assetClient, sellerUserID, "100", "P0-ORDER-ADMISSION-SELLER-SEED")

	sellerResp := placeP0Order(t, ctx, serviceCtx, sellerUserID, &option.PlaceOrderReq{
		AccountId: 8050, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-ORDER-ADMISSION-SELLER",
	})
	sellerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, sellerResp.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if sellerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_FUNDING) ||
		!sellerOrder.MarginAmount.Equal(decimal.NewFromInt(50)) || sellerOrder.MarginCoin != "USDT" {
		t.Fatalf("unexpected admitted seller order: %+v", sellerOrder)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	sellerOrder, err = serviceCtx.OptionOrderModel.FindOne(ctx, sellerOrder.Id)
	if err != nil {
		t.Fatal(err)
	}
	if sellerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) {
		t.Fatalf("funded seller was not admitted to book: %+v", sellerOrder)
	}

	buyerResp := placeP0Order(t, ctx, serviceCtx, buyerUserID, &option.PlaceOrderReq{
		AccountId: 7050, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-ORDER-ADMISSION-BUYER",
	})
	buyerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, buyerResp.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if buyerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_FUNDING) ||
		!buyerOrder.MarginAmount.Equal(decimal.RequireFromString("10.4")) || buyerOrder.MarginCoin != "USDT" {
		t.Fatalf("unexpected admitted buyer order: %+v", buyerOrder)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)
	processP0OrderRiskAccounts(
		t, ctx, serviceCtx, contract, buyerUserID, sellerUserID,
	)

	assertP0OrderAdmissionEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyerOrder.Id, sellerOrder.Id,
		buyerUserID, sellerUserID, feeUserID,
	)
	assertWalletAmounts(t, ctx, db, buyerUserID,
		"89.600000000000000000", "89.600000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerUserID,
		"109.800000000000000000", "59.800000000000000000", "50.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID,
		"0.600000000000000000", "0.600000000000000000", "0.000000000000000000")

	// The public API reports the existing economic identity and must not create
	// another order, freeze, trade, or position when the client key is replayed.
	replayCtx := p0OrderUserContext(ctx, buyerUserID)
	replay, err := applogic.NewPlaceOrderLogic(replayCtx, serviceCtx).PlaceOrder(&option.PlaceOrderReq{
		AccountId: 7050, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-ORDER-ADMISSION-BUYER",
	})
	if err != nil || replay == nil || replay.Data == nil || replay.Data.OrderId != buyerOrder.Id ||
		replay.Base == nil || replay.Base.Code == 200 {
		t.Fatalf("unexpected client-order replay response resp=%+v err=%v", replay, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)
	processP0OrderRiskAccounts(
		t, ctx, serviceCtx, contract, buyerUserID, sellerUserID,
	)
	assertP0OrderAdmissionEvidence(
		t, ctx, db, serviceCtx, contract.Id, buyerOrder.Id, sellerOrder.Id,
		buyerUserID, sellerUserID, feeUserID,
	)
}

func placeP0Order(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID int64,
	req *option.PlaceOrderReq,
) *option.PlaceOrderResp {
	t.Helper()
	resp, err := applogic.NewPlaceOrderLogic(p0OrderUserContext(ctx, userID), serviceCtx).PlaceOrder(req)
	if err != nil {
		t.Fatalf("place order user=%d: %v", userID, err)
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 || resp.Data == nil || resp.Data.OrderId <= 0 {
		t.Fatalf("unexpected place-order response user=%d resp=%+v", userID, resp)
	}
	return resp
}

func p0OrderUserContext(ctx context.Context, userID int64) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUid, strconv.FormatInt(userID, 10),
		utils.CtxKeyTenantId, strconv.FormatInt(p0AssetE2ETenantID, 10),
	))
}

func seedP0OpenTradingCalendar(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	calendarCode string,
	now int64,
) {
	t.Helper()
	result, err := db.ExecContext(ctx, `INSERT INTO t_option_trading_calendar
		(tenant_id,calendar_code,version,status,timezone,effective_from,effective_until,
		 change_reason,evidence_ref,created_by,reviewed_by,review_reason,reviewed_at,create_times,update_times)
		VALUES (?,?,1,2,'UTC',?,0,'P0 full order admission','P0-E2E',9001,9002,'approved',?,?,?)`,
		p0AssetE2ETenantID, calendarCode, now-3600, now-1800, now-3600, now-1800,
	)
	if err != nil {
		t.Fatalf("insert P0 trading calendar: %v", err)
	}
	calendarID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	weekday := int64(time.Unix(now, 0).UTC().Weekday())
	if _, err := db.ExecContext(ctx, `INSERT INTO t_option_trading_calendar_session
		(tenant_id,calendar_id,weekday,open_second,close_second,create_times)
		VALUES (?,?,?,0,86400,?)`, p0AssetE2ETenantID, calendarID, weekday, now); err != nil {
		t.Fatalf("insert P0 trading session: %v", err)
	}
}

func insertP0OrderAdmissionContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	feeUserID, now int64,
) *models.TOptionContract {
	t.Helper()
	return insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-FULL-ORDER-ADMISSION-CALL", calendarCode, feeUserID, now,
	)
}

func insertP0OrderTestContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractCode, calendarCode string,
	feeUserID, now int64,
) *models.TOptionContract {
	t.Helper()
	contract := &models.TOptionContract{
		TenantId: p0AssetE2ETenantID, ContractCode: contractCode,
		UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		ExerciseStyle:  int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		StrikePrice:    decimal.NewFromInt(100), ContractUnit: decimal.NewFromInt(1),
		MinOrderQty: decimal.NewFromInt(1), MaxOrderQty: decimal.NewFromInt(1000),
		PriceTick: decimal.RequireFromString("0.1"), QtyStep: decimal.NewFromInt(1),
		Multiplier: decimal.NewFromInt(1), ListTime: now - 3600,
		ExerciseCutoffTime: now + 3600, ExpireTime: now + 7200, DeliverTime: now + 7200,
		AutoExerciseThreshold: decimal.NewFromInt(1), MaxUserLongQty: decimal.NewFromInt(10000),
		MaxUserShortQty: decimal.NewFromInt(10000), MaxOpenInterest: decimal.NewFromInt(10000),
		OrderPriceBandRatio: decimal.RequireFromString("0.2"),
		CircuitBreakerRatio: decimal.RequireFromString("0.5"), GreeksMaxAgeSeconds: 60,
		SettlementPriceSource: "authoritative-market", SettlementPriceMethod: "MEDIAN",
		SettlementWindowSeconds: 60, SettlementMinSamples: 3,
		IsAutoExercise: int64(common.YesNo_YES_NO_NO),
		MakerFeeRate:   decimal.RequireFromString("0.02"), TakerFeeRate: decimal.RequireFromString("0.04"),
		ExerciseFeeRate: decimal.RequireFromString("0.1"), FeeUserId: feeUserID, FeeAccountId: 9050,
		SellerMarginMode:      int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
		InitialMarginRate:     decimal.RequireFromString("0.5"),
		MaintenanceMarginRate: decimal.RequireFromString("0.2"), MinMarginRate: decimal.RequireFromString("0.1"),
		LiquidationFeeRate: decimal.RequireFromString("0.1"), InsuranceUserId: 164, InsuranceAccountId: 9051,
		LiquidationDeficitPolicy: int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW),
		TradingCalendarCode:      calendarCode, Status: int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		IsDeleted: int64(common.YesNo_YES_NO_NO), CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert order-admission contract: %v", err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func processP0OrderRiskAccounts(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	buyerUserID, sellerUserID int64,
) {
	t.Helper()
	processP0OrderRiskAccountsForAccounts(
		t, ctx, serviceCtx, contract, buyerUserID, 7050, sellerUserID, 8050,
	)
}

func processP0OrderRiskAccountsForAccounts(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	buyerUserID, buyerAccountID, sellerUserID, sellerAccountID int64,
) {
	t.Helper()
	market, err := serviceCtx.OptionMarketModel.FindOneByTenantIdContractId(
		ctx, p0AssetE2ETenantID, contract.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	logic := NewProcessRiskAccountsLogic(ctx, serviceCtx)
	for _, item := range []struct {
		userID, accountID int64
		side              common.PositionSide
	}{
		{buyerUserID, buyerAccountID, common.PositionSide_POSITION_SIDE_LONG},
		{sellerUserID, sellerAccountID, common.PositionSide_POSITION_SIDE_SHORT},
	} {
		position, err := serviceCtx.OptionPositionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
			ctx, p0AssetE2ETenantID, item.userID, item.accountID, contract.Id, int64(item.side),
		)
		if err != nil {
			t.Fatal(err)
		}
		group := &optionRiskGroup{
			key: optionRiskKey{
				tenantID: p0AssetE2ETenantID, userID: item.userID, accountID: 0, coin: "USDT",
			},
			positions: []optionRiskPosition{{position: position, contract: contract, market: market}},
		}
		if err := logic.refreshRiskGroup(group); err != nil {
			t.Fatalf("refresh admitted order risk user=%d: %v", item.userID, err)
		}
	}
}

func assertP0OrderAdmissionEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID, buyerOrderID, sellerOrderID, buyerUserID, sellerUserID, feeUserID int64,
) {
	t.Helper()
	var orders, clientKeys, trades, instructions, success, reconciled, flows, outbox, inbox, lots int64
	var turnover, buyFee, sellFee decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=3) FROM t_option_order
		WHERE tenant_id=? AND id IN (?,?)`, p0AssetE2ETenantID, buyerOrderID, sellerOrderID).
		Scan(&orders, &success); err != nil {
		t.Fatal(err)
	}
	if orders != 2 || success != 2 {
		t.Fatalf("order admission order count/filled=%d/%d", orders, success)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_client_order_key
		WHERE tenant_id=? AND order_id IN (?,?)`, p0AssetE2ETenantID, buyerOrderID, sellerOrderID).
		Scan(&clientKeys); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),MAX(turnover),MAX(buy_fee),MAX(sell_fee)
		FROM t_option_trade WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).
		Scan(&trades, &turnover, &buyFee, &sellFee); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=3),SUM(reconciliation_status=2)
		FROM t_option_asset_instruction WHERE tenant_id=? AND
		(order_id IN (?,?) OR trade_id IN (
		  SELECT id FROM t_option_trade WHERE tenant_id=? AND contract_id=?
		))`, p0AssetE2ETenantID, buyerOrderID, sellerOrderID, p0AssetE2ETenantID, contractID).
		Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction JOIN t_asset_flow flow
		 ON flow.tenant_id=instruction.tenant_id
		AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND
		(instruction.order_id IN (?,?) OR instruction.trade_id IN (
		  SELECT id FROM t_option_trade WHERE tenant_id=? AND contract_id=?
		))`, p0AssetE2ETenantID, buyerOrderID, sellerOrderID, p0AssetE2ETenantID, contractID).
		Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT SUM(status=3) FROM t_option_outbox
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&outbox); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT SUM(status=2) FROM t_option_inbox
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&inbox); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_margin_lot
		WHERE tenant_id=? AND contract_id=? AND initial_margin=50 AND remaining_margin=50`,
		p0AssetE2ETenantID, contractID).Scan(&lots); err != nil {
		t.Fatal(err)
	}
	if clientKeys != 2 || trades != 1 || !turnover.Equal(decimal.NewFromInt(10)) ||
		!buyFee.Equal(decimal.RequireFromString("0.4")) || !sellFee.Equal(decimal.RequireFromString("0.2")) ||
		instructions != 5 || success != 5 || reconciled != 5 || flows != 5 || outbox != 1 || inbox != 1 || lots != 1 {
		t.Fatalf("order admission evidence keys/trades/value/fees/instructions/flows/events/lots=%d/%d/%s/%s/%s/%d/%d/%d/%d/%d/%d/%d",
			clientKeys, trades, turnover, buyFee, sellFee, instructions, success, reconciled, flows, outbox, inbox, lots)
	}
	buyer, err := serviceCtx.OptionPositionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
		ctx, p0AssetE2ETenantID, buyerUserID, 7050, contractID,
		int64(common.PositionSide_POSITION_SIDE_LONG),
	)
	if err != nil {
		t.Fatal(err)
	}
	seller, err := serviceCtx.OptionPositionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
		ctx, p0AssetE2ETenantID, sellerUserID, 8050, contractID,
		int64(common.PositionSide_POSITION_SIDE_SHORT),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !buyer.PositionQty.Equal(decimal.NewFromInt(1)) || !buyer.FeePaid.Equal(decimal.RequireFromString("0.4")) ||
		!buyer.TotalReturn.Equal(decimal.RequireFromString("-0.4")) ||
		!seller.PositionQty.Equal(decimal.NewFromInt(1)) || !seller.MarginAmount.Equal(decimal.NewFromInt(50)) ||
		!seller.MaintenanceMargin.Equal(decimal.NewFromInt(20)) || !seller.FeePaid.Equal(decimal.RequireFromString("0.2")) ||
		!seller.TotalReturn.Equal(decimal.RequireFromString("-0.2")) {
		t.Fatalf("unexpected admitted positions buyer=%+v seller=%+v", buyer, seller)
	}
	var riskAccounts, normalAccounts int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=1) FROM t_option_risk_account
		WHERE tenant_id=? AND user_id IN (?,?) AND account_id=0 AND settle_coin='USDT'`,
		p0AssetE2ETenantID, buyerUserID, sellerUserID).Scan(&riskAccounts, &normalAccounts); err != nil {
		t.Fatal(err)
	}
	var walletTotal decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT SUM(total_amount) FROM t_user_asset
		WHERE tenant_id=? AND wallet_type=5 AND coin='USDT' AND user_id IN (?,?,?)`,
		p0AssetE2ETenantID, buyerUserID, sellerUserID, feeUserID).Scan(&walletTotal); err != nil {
		t.Fatal(err)
	}
	if riskAccounts != 2 || normalAccounts != 2 || !walletTotal.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("risk/wallet conservation accounts/normal/wallet=%d/%d/%s",
			riskAccounts, normalAccounts, walletTotal)
	}
}
