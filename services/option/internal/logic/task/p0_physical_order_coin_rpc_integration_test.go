package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func testP0PhysicalOrderCoinLifecycle(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P0_PHYSICAL_ORDER_COIN_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)

	type physicalCase struct {
		label       string
		optionType  option.OptionType
		coin        string
		amount      string
		userBase    int64
		feeUserID   int64
		contract    *models.TOptionContract
		expiryID    int64
		expiryUser  int64
		expiryAcct  int64
		expiryOrder string
	}
	cases := []*physicalCase{
		{
			label: "CALL", optionType: option.OptionType_OPTION_TYPE_CALL,
			coin: "BTC", amount: "1", userBase: 891000, feeUserID: 899001,
			expiryID: 996221, expiryUser: 893001, expiryAcct: 8401,
			expiryOrder: "P0-PHYSICAL-COIN-CALL-EXPIRY",
		},
		{
			label: "PUT", optionType: option.OptionType_OPTION_TYPE_PUT,
			coin: "USDT", amount: "100", userBase: 892000, feeUserID: 899002,
			expiryID: 996222, expiryUser: 894001, expiryAcct: 8402,
			expiryOrder: "P0-PHYSICAL-COIN-PUT-EXPIRY",
		},
	}
	for _, item := range cases {
		item.contract = insertP0TradablePhysicalOrderContract(
			t, ctx, serviceCtx, "P0-PHYSICAL-COIN-"+item.label,
			calendarCode, item.optionType, item.feeUserID, now,
		)
		insertP0ExerciseMarket(t, ctx, serviceCtx, item.contract.Id, "100", "10", now)
		testP0PhysicalOrdinaryOrderTypes(t, ctx, db, assetClient, serviceCtx, item.contract,
			item.label, item.coin, item.amount, item.userBase)
		testP0PhysicalLiquidationOrderRelease(t, ctx, db, assetClient, serviceCtx, item.contract,
			item.label, item.coin, item.amount, item.userBase+10)
		testP0PhysicalAdminOrderRelease(t, ctx, db, assetClient, serviceCtx, item.contract,
			item.label, item.coin, item.amount, item.userBase+11)
	}
	for _, item := range cases {
		testP0PhysicalExpiryOrderRelease(
			t, ctx, db, assetClient, serviceCtx, item.expiryID, item.expiryOrder,
			item.optionType, item.expiryUser, item.expiryAcct, item.coin, item.amount, now,
		)
	}
}

func testP0PhysicalOrdinaryOrderTypes(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	label, coin, amount string,
	userBase int64,
) {
	t.Helper()
	types := []struct {
		name           string
		orderType      option.OrderType
		explicitCancel bool
	}{
		{name: "LIMIT", orderType: option.OrderType_ORDER_TYPE_LIMIT, explicitCancel: true},
		{name: "MARKET", orderType: option.OrderType_ORDER_TYPE_MARKET},
		{name: "POST_ONLY", orderType: option.OrderType_ORDER_TYPE_POST_ONLY, explicitCancel: true},
		{name: "IOC", orderType: option.OrderType_ORDER_TYPE_IOC},
		{name: "FOK", orderType: option.OrderType_ORDER_TYPE_FOK},
	}
	for index, orderType := range types {
		userID := userBase + int64(index+1)
		accountID := userBase + int64(100+index)
		clientOrderID := fmt.Sprintf("P0-PHYSICAL-%s-%s", label, orderType.name)
		creditAssetCoin(t, ctx, assetClient, userID, coin, amount, clientOrderID+"-SEED")
		request := &option.PlaceOrderReq{
			AccountId: accountID, ContractId: contract.Id,
			Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: orderType.orderType, Price: "10", Qty: "1", ClientOrderId: clientOrderID,
		}
		if orderType.orderType == option.OrderType_ORDER_TYPE_MARKET {
			request.Price = ""
			request.ProtectionPrice = "10"
		}
		placed := placeP0Order(t, ctx, serviceCtx, userID, request)
		for attempt := 0; attempt < 4; attempt++ {
			processAssetInstructions(t, ctx, serviceCtx)
		}
		if orderType.explicitCancel {
			current, err := serviceCtx.OptionOrderModel.FindOne(ctx, placed.Data.OrderId)
			if err != nil {
				t.Fatal(err)
			}
			if current.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) {
				t.Fatalf("physical %s %s did not rest before cancel: %+v", label, orderType.name, current)
			}
			assertP0UserCancelOK(t, ctx, serviceCtx, userID, accountID, current.Id)
		}
		for attempt := 0; attempt < 4; attempt++ {
			processAssetInstructions(t, ctx, serviceCtx)
		}
		assertP0PhysicalOrderReleased(
			t, ctx, db, serviceCtx, placed.Data.OrderId, userID, coin, amount, "",
		)
		for attempt := 0; attempt < 2; attempt++ {
			processAssetInstructions(t, ctx, serviceCtx)
		}
		assertP0PhysicalOrderReleased(
			t, ctx, db, serviceCtx, placed.Data.OrderId, userID, coin, amount, "",
		)
	}
}

func testP0PhysicalLiquidationOrderRelease(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	label, coin, amount string,
	userID int64,
) {
	t.Helper()
	accountID := userID + 100
	clientOrderID := "P0-PHYSICAL-" + label + "-LIQUIDATION-CANCEL"
	creditAssetCoin(t, ctx, assetClient, userID, coin, amount, clientOrderID+"-SEED")
	placed := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: clientOrderID,
	})
	processAssetInstructions(t, ctx, serviceCtx)
	if err := cancelOptionSystemOrder(ctx, serviceCtx, placed.Data.OrderId, "LIQUIDATION"); err != nil {
		t.Fatalf("physical %s liquidation cancel: %v", label, err)
	}
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0PhysicalOrderReleased(
		t, ctx, db, serviceCtx, placed.Data.OrderId, userID, coin, amount, "LIQUIDATION",
	)
}

func testP0PhysicalAdminOrderRelease(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	label, coin, amount string,
	userID int64,
) {
	t.Helper()
	const operatorID int64 = 9010
	accountID := userID + 100
	clientOrderID := "P0-PHYSICAL-" + label + "-ADMIN-CANCEL"
	reason := "P0_PHYSICAL_" + label + "_ADMIN_CANCEL"
	creditAssetCoin(t, ctx, assetClient, userID, coin, amount, clientOrderID+"-SEED")
	placed := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: clientOrderID,
	})
	processAssetInstructions(t, ctx, serviceCtx)
	contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
	contract.UpdateTimes = time.Now().Unix()
	if err := serviceCtx.OptionContractModel.Update(ctx, contract); err != nil {
		t.Fatalf("pause physical %s admin-cancel contract: %v", label, err)
	}
	resp, err := adminlogic.NewForceCancelContractOrdersLogic(
		p0AdminContext(ctx, operatorID, p0AssetE2ETenantID), serviceCtx,
	).ForceCancelContractOrders(&option.ForceCancelContractOrdersReq{
		TenantId: p0AssetE2ETenantID, ContractId: contract.Id, Reason: reason,
	})
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("physical %s admin cancel resp=%+v err=%v", label, resp, err)
	}
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0PhysicalOrderReleased(
		t, ctx, db, serviceCtx, placed.Data.OrderId, userID, coin, amount, reason,
	)
	var events, operators, reasons int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(operator_id=?),SUM(reason=?)
		FROM t_option_trading_control_event
		WHERE tenant_id=? AND contract_id=? AND order_id=? AND event_type='ADMIN_FORCE_CANCEL_ORDER'`,
		operatorID, reason, p0AssetE2ETenantID, contract.Id, placed.Data.OrderId,
	).Scan(&events, &operators, &reasons); err != nil {
		t.Fatal(err)
	}
	if events != 1 || operators != 1 || reasons != 1 {
		t.Fatalf("physical %s admin audit events/operators/reasons=%d/%d/%d", label, events, operators, reasons)
	}
}

func testP0PhysicalExpiryOrderRelease(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contractID int64,
	orderNo string,
	optionType option.OptionType,
	userID, accountID int64,
	coin, amount string,
	now int64,
) {
	t.Helper()
	seedP1PhysicalContract(
		t, ctx, db, contractID, orderNo+"-CONTRACT", optionType, now-10, now+3600,
	)
	creditAssetCoin(t, ctx, assetClient, userID, coin, amount, orderNo+"-SEED")
	order := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: orderNo, UserId: userID, AccountId: accountID,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT", Side: int64(common.Side_SIDE_SELL),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType:      int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(10),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1), FeeCoin: "USDT",
		MarginAmount: decimal.RequireFromString(amount), MarginCoin: coin,
		Source: int64(option.OrderSource_ORDER_SOURCE_APP), ReduceOnly: int64(common.YesNo_YES_NO_NO),
		Mmp: int64(common.YesNo_YES_NO_NO), Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now - 5, UpdateTimes: now - 5,
	})
	freeze, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: coin, Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: order.Id, BizNo: order.OrderNo, Remark: "P0 physical expiry order freeze",
	})
	assertAssetOK(t, freeze, err)
	if _, err := serviceCtx.OptionAssetInstructionModel.Insert(ctx, &models.TOptionAssetInstruction{
		TenantId: p0AssetE2ETenantID, InstructionNo: order.OrderNo + "-FREEZE", BizNo: order.OrderNo,
		OrderId: order.Id, UserId: userID, AccountId: accountID,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		TargetBizNo: order.OrderNo, Coin: coin, Amount: decimal.RequireFromString(amount), StepNo: 1,
		Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		ReconciledAt:         now, CreateTimes: now - 5, UpdateTimes: now - 5,
	}); err != nil {
		t.Fatalf("insert physical expiry freeze evidence: %v", err)
	}
	contract, err := serviceCtx.OptionContractModel.FindOne(ctx, contractID)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewProcessContractLifecycleLogic(ctx, serviceCtx).expireContractOrders(contract, now); err != nil {
		t.Fatalf("expire physical %s order: %v", optionType, err)
	}
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0PhysicalOrderReleased(
		t, ctx, db, serviceCtx, order.Id, userID, coin, amount, "CONTRACT_EXPIRED",
	)
}

func insertP0TradablePhysicalOrderContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractCode, calendarCode string,
	optionType option.OptionType,
	feeUserID, now int64,
) *models.TOptionContract {
	t.Helper()
	contract := &models.TOptionContract{
		TenantId: p0AssetE2ETenantID, ContractCode: contractCode,
		UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType: int64(optionType), ExerciseStyle: int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL),
		StrikePrice:    decimal.NewFromInt(100), ContractUnit: decimal.NewFromInt(1),
		MinOrderQty: decimal.NewFromInt(1), MaxOrderQty: decimal.NewFromInt(1000),
		PriceTick: decimal.RequireFromString("0.1"), QtyStep: decimal.NewFromInt(1), Multiplier: decimal.NewFromInt(1),
		ListTime: now - 3600, ExerciseCutoffTime: now + 3600, ExpireTime: now + 7200, DeliverTime: now + 7200,
		AutoExerciseThreshold: decimal.NewFromInt(1), MaxUserLongQty: decimal.NewFromInt(10000),
		MaxUserShortQty: decimal.NewFromInt(10000), MaxOpenInterest: decimal.NewFromInt(10000),
		OrderPriceBandRatio: decimal.RequireFromString("0.2"), CircuitBreakerRatio: decimal.RequireFromString("0.5"),
		GreeksMaxAgeSeconds: 60, SettlementPriceSource: "authoritative-market",
		SettlementPriceMethod: "MEDIAN", SettlementWindowSeconds: 60, SettlementMinSamples: 3,
		IsAutoExercise: int64(common.YesNo_YES_NO_NO), MakerFeeRate: decimal.RequireFromString("0.02"),
		TakerFeeRate: decimal.RequireFromString("0.04"), ExerciseFeeRate: decimal.RequireFromString("0.1"),
		FeeUserId: feeUserID, FeeAccountId: feeUserID + 100,
		SellerMarginMode:  int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY),
		InitialMarginRate: decimal.NewFromInt(1), MaintenanceMarginRate: decimal.RequireFromString("0.1"),
		MinMarginRate: decimal.NewFromInt(1), LiquidationFeeRate: decimal.RequireFromString("0.1"),
		InsuranceUserId: feeUserID + 10, InsuranceAccountId: feeUserID + 110,
		LiquidationDeficitPolicy:    int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW),
		PhysicalDeliveryPolicy:      int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_STRICT),
		PhysicalDeliveryCureSeconds: 3600, TradingCalendarCode: calendarCode,
		Status:    int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		IsDeleted: int64(common.YesNo_YES_NO_NO), CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert tradable physical contract %s: %v", contractCode, err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func assertP0PhysicalOrderReleased(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	orderID, userID int64,
	coin, amount, reason string,
) {
	t.Helper()
	order, err := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	wantStatus := int64(option.OrderStatus_ORDER_STATUS_CANCELED)
	if reason == "CONTRACT_EXPIRED" {
		wantStatus = int64(option.OrderStatus_ORDER_STATUS_EXPIRED)
	}
	if order.Status != wantStatus ||
		!order.MarginAmount.IsZero() || order.MarginCoin != coin ||
		(reason != "" && order.CancelReason != reason) {
		t.Fatalf("physical released order=%+v expected coin/reason=%s/%s", order, coin, reason)
	}
	var instructions, success, reconciled, freezes, releases, wrongCoins, flows int64
	var frozenAmount, releasedAmount decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?),
		SUM(action=?),SUM(action=?),SUM(coin<>?),
		SUM(CASE WHEN action=? THEN amount ELSE 0 END),
		SUM(CASE WHEN action=? THEN amount ELSE 0 END)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN), coin,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
		p0AssetE2ETenantID, orderID,
	).Scan(&instructions, &success, &reconciled, &freezes, &releases, &wrongCoins, &frozenAmount, &releasedAmount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id=?`,
		p0AssetE2ETenantID, orderID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	wantAmount := decimal.RequireFromString(amount)
	if instructions != 2 || success != 2 || reconciled != 2 || freezes != 1 || releases != 1 ||
		wrongCoins != 0 || flows != 2 || !frozenAmount.Equal(wantAmount) || !releasedAmount.Equal(wantAmount) {
		t.Fatalf("physical order %d instructions/success/reconciled/freeze/release/wrong/flows/amounts=%d/%d/%d/%d/%d/%d/%d/%s/%s want=%s",
			orderID, instructions, success, reconciled, freezes, releases, wrongCoins, flows,
			frozenAmount, releasedAmount, wantAmount)
	}
	assertWalletCoinAmounts(t, ctx, db, userID, coin,
		wantAmount.StringFixed(18), wantAmount.StringFixed(18), "0.000000000000000000")
}
