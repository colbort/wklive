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

func testP0PartialCloseTradeAccounting(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		longUserID          int64 = 131
		newBuyerUserID      int64 = 132
		originalShortUserID int64 = 133
		feeUserID           int64 = 134
	)
	now := time.Now().Unix()
	contract := insertP0PartialCloseContract(t, ctx, serviceCtx, feeUserID, now)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "115", "15", now)

	creditAsset(t, ctx, assetClient, longUserID, "100", "P0-PARTIAL-CLOSE-LONG-SEED")
	creditAsset(t, ctx, assetClient, newBuyerUserID, "100", "P0-PARTIAL-CLOSE-BUYER-SEED")
	creditAsset(t, ctx, assetClient, originalShortUserID, "100", "P0-PARTIAL-CLOSE-SHORT-SEED")
	transferP0OptionPremium(t, ctx, assetClient, longUserID, originalShortUserID, "20", "P0-PARTIAL-CLOSE-OPEN-PREMIUM")
	transferP0OptionPremium(t, ctx, assetClient, longUserID, feeUserID, "1", "P0-PARTIAL-CLOSE-OPEN-FEE")

	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: 7030,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(1), FrozenQty: decimal.NewFromInt(1),
		OpenAvgPrice: decimal.NewFromInt(10), MarkPrice: decimal.NewFromInt(10),
		PositionValue: decimal.NewFromInt(20), ExerciseableQty: decimal.NewFromInt(2),
		FeePaid: decimal.NewFromInt(1), TotalReturn: decimal.NewFromInt(-1),
		RealizedPnl: decimal.NewFromInt(-1),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: originalShortUserID, AccountId: 8030,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(10), PositionValue: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})

	closeOrder := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P0-PARTIAL-CLOSE-MAKER", UserId: longUserID,
		AccountId: 7030, ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.Side_SIDE_SELL), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_CLOSE),
		OrderType: int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(15),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1), FeeCoin: "USDT", MarginCoin: "USDT",
		Source: int64(option.OrderSource_ORDER_SOURCE_APP), ReduceOnly: int64(common.YesNo_YES_NO_YES),
		Mmp: int64(common.YesNo_YES_NO_NO), Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now - 10, UpdateTimes: now - 10,
	})
	buyOrder := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P0-PARTIAL-CLOSE-TAKER", UserId: newBuyerUserID,
		AccountId: 7031, ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.Side_SIDE_BUY), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType: int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(15),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1), FeeCoin: "USDT",
		MarginAmount: decimal.RequireFromString("15.6"), MarginCoin: "USDT",
		Source: int64(option.OrderSource_ORDER_SOURCE_APP), ReduceOnly: int64(common.YesNo_YES_NO_NO),
		Mmp: int64(common.YesNo_YES_NO_NO), Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now, UpdateTimes: now,
	})
	freezeResp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: newBuyerUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "15.6",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: buyOrder.Id, BizNo: buyOrder.OrderNo, Remark: "P0 partial close buyer reservation",
	})
	assertAssetOK(t, freezeResp, err)

	if err := applogic.MatchFundedOrder(ctx, serviceCtx, buyOrder); err != nil {
		t.Fatalf("match partial close: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)

	var tradeID int64
	var turnover, buyFee, sellFee string
	if err := db.QueryRowContext(ctx, `SELECT id,CAST(turnover AS CHAR),CAST(buy_fee AS CHAR),CAST(sell_fee AS CHAR)
		FROM t_option_trade WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contract.Id).
		Scan(&tradeID, &turnover, &buyFee, &sellFee); err != nil {
		t.Fatal(err)
	}
	if turnover != "15.0000000000000000" || buyFee != "0.6000000000000000" || sellFee != "0.3000000000000000" {
		t.Fatalf("partial close trade turnover/buyFee/sellFee=%s/%s/%s", turnover, buyFee, sellFee)
	}

	closedLong, err := serviceCtx.OptionPositionModel.FindOne(ctx, longPosition.Id)
	if err != nil {
		t.Fatal(err)
	}
	assertP0PartialClosePosition(t, closedLong, "1", "1", "0", "10", "15", "15", "5", "5", "0", "1.3", "3.7")
	newLong, err := serviceCtx.OptionPositionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
		ctx, p0AssetE2ETenantID, newBuyerUserID, 7031, contract.Id,
		int64(common.PositionSide_POSITION_SIDE_LONG),
	)
	if err != nil {
		t.Fatal(err)
	}
	assertP0PartialClosePosition(t, newLong, "1", "1", "0", "15", "15", "15", "0", "0", "0", "0.6", "-0.6")

	assertWalletAmounts(t, ctx, db, longUserID, "93.700000000000000000", "93.700000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, newBuyerUserID, "84.400000000000000000", "84.400000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, originalShortUserID, "120.000000000000000000", "120.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "1.900000000000000000", "1.900000000000000000", "0.000000000000000000")

	assertP0PartialCloseEvidence(t, ctx, db, contract.Id, tradeID, closeOrder.Id, buyOrder.Id)
	processAssetInstructions(t, ctx, serviceCtx)
	processP0TradeEvents(t, ctx, serviceCtx)
	assertP0PartialCloseEvidence(t, ctx, db, contract.Id, tradeID, closeOrder.Id, buyOrder.Id)
}

func insertP0PartialCloseContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	feeUserID, now int64,
) *models.TOptionContract {
	t.Helper()
	contract := &models.TOptionContract{
		TenantId: p0AssetE2ETenantID, ContractCode: "P0-PARTIAL-CLOSE-CALL",
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
		IsAutoExercise: int64(common.YesNo_YES_NO_NO),
		MakerFeeRate:   decimal.RequireFromString("0.02"), TakerFeeRate: decimal.RequireFromString("0.04"),
		ExerciseFeeRate: decimal.RequireFromString("0.1"), FeeUserId: feeUserID, FeeAccountId: 9030,
		SellerMarginMode:      int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
		InitialMarginRate:     decimal.RequireFromString("0.5"),
		MaintenanceMarginRate: decimal.RequireFromString("0.2"), MinMarginRate: decimal.RequireFromString("0.1"),
		TradingCalendarCode: "CONTINUOUS_24_7", Status: int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		IsDeleted: int64(common.YesNo_YES_NO_NO), CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert partial close contract: %v", err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func processP0TradeEvents(t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext) {
	t.Helper()
	resp, err := NewProcessTradeEventsLogic(ctx, serviceCtx).ProcessTradeEvents(&option.OptionTaskReq{
		TenantId: p0AssetE2ETenantID,
	})
	if err != nil {
		t.Fatalf("process trade events: %v", err)
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("unexpected trade event task response: %+v", resp)
	}
}

func assertP0PartialClosePosition(
	t *testing.T,
	position *models.TOptionPosition,
	qty, available, frozen, openAvg, mark, value, unrealized, tradeRealized,
	settlementRealized, fee, total string,
) {
	t.Helper()
	wants := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"qty", position.PositionQty, qty}, {"available", position.AvailableQty, available},
		{"frozen", position.FrozenQty, frozen}, {"open_avg", position.OpenAvgPrice, openAvg},
		{"mark", position.MarkPrice, mark}, {"value", position.PositionValue, value},
		{"unrealized", position.UnrealizedPnl, unrealized},
		{"trade_realized", position.TradeRealizedPnl, tradeRealized},
		{"settlement_realized", position.SettlementRealizedPnl, settlementRealized},
		{"fee", position.FeePaid, fee}, {"total", position.TotalReturn, total},
		{"realized", position.RealizedPnl, total},
	}
	for _, item := range wants {
		if !item.got.Equal(decimal.RequireFromString(item.want)) {
			t.Fatalf("position %d %s=%s want=%s", position.Id, item.name, item.got, item.want)
		}
	}
}

func assertP0PartialCloseEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, tradeID, closeOrderID, buyOrderID int64,
) {
	t.Helper()
	var instructions, success, reconciled, flows, outboxSuccess int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND trade_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, tradeID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow AS flow
		JOIN t_option_asset_instruction AS instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.trade_id=?`, p0AssetE2ETenantID, tradeID).
		Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_outbox
		WHERE tenant_id=? AND contract_id=? AND trade_id=? AND status=?`,
		p0AssetE2ETenantID, contractID, tradeID,
		int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
	).Scan(&outboxSuccess); err != nil {
		t.Fatal(err)
	}
	if instructions != 3 || success != 3 || reconciled != 3 || flows != 3 || outboxSuccess != 1 {
		t.Fatalf("partial close evidence instructions/success/reconciled/flows/outbox=%d/%d/%d/%d/%d",
			instructions, success, reconciled, flows, outboxSuccess)
	}
	var closeStatus, buyStatus int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_order WHERE id=?`, closeOrderID).Scan(&closeStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_order WHERE id=?`, buyOrderID).Scan(&buyStatus); err != nil {
		t.Fatal(err)
	}
	wantFilled := int64(option.OrderStatus_ORDER_STATUS_FILLED)
	if closeStatus != wantFilled || buyStatus != wantFilled {
		t.Fatalf("partial close order statuses close/buy=%d/%d", closeStatus, buyStatus)
	}
}
