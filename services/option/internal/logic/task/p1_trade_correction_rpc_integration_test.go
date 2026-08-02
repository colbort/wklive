package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
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

func testP1TradeCorrectionAssetRPC(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		buyerUserID      int64 = 219
		buyerAccountID   int64 = 7219
		sellerUserID     int64 = 220
		sellerAccountID  int64 = 8220
		feeUserID        int64 = 221
		feeAccountID     int64 = 9221
		activeUserID     int64 = 222
		activeAccountID  int64 = 7222
		requesterID      int64 = 880021
		reviewerIDBase   int64 = 880100
		invalidUserID    int64 = 999991
		invalidAccountID int64 = 999992
	)
	now := time.Now().Unix()
	contract := insertP1TradeCorrectionContract(
		t, ctx, serviceCtx, feeUserID, feeAccountID, now,
	)
	trade := insertP1TradeCorrectionTrade(
		t, ctx, serviceCtx, contract.Id,
		buyerUserID, buyerAccountID, sellerUserID, sellerAccountID, now,
	)
	originalTradeHash := p1TradeCorrectionTradeHash(t, ctx, db, trade.Id)

	creditAsset(t, ctx, assetClient, buyerUserID, "20", "P1-TRADE-CORRECTION-BUYER-SEED")
	creditAsset(t, ctx, assetClient, sellerUserID, "2", "P1-TRADE-CORRECTION-SELLER-SEED")
	creditAsset(t, ctx, assetClient, feeUserID, "1", "P1-TRADE-CORRECTION-FEE-SEED")
	creditAsset(t, ctx, assetClient, activeUserID, "20", "P1-TRADE-CORRECTION-ACTIVE-SEED")
	activeOrder := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P1-TRADE-CORRECTION-ACTIVE",
		UserId: activeUserID, AccountId: activeAccountID, ContractId: contract.Id,
		UnderlyingSymbol: "BTCUSDT", Side: int64(common.Side_SIDE_BUY),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType:      int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(5),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1),
		FeeCoin: "USDT", MarginAmount: decimal.NewFromInt(5), MarginCoin: "USDT",
		Source:     int64(option.OrderSource_ORDER_SOURCE_APP),
		ReduceOnly: int64(common.YesNo_YES_NO_NO), Mmp: int64(common.YesNo_YES_NO_NO),
		Status:      int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now, UpdateTimes: now,
	})
	freezeResp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: activeUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "5",
		BizType:   asset.BizType_BIZ_TYPE_OPTION,
		SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId:     activeOrder.Id, BizNo: activeOrder.OrderNo,
		Remark: "P1 trade correction active order",
	})
	assertAssetOK(t, freezeResp, err)

	requesterCtx := p0AdminContext(ctx, requesterID, p0AssetE2ETenantID)
	invalid, err := adminlogic.NewCreateTradeCorrectionLogic(requesterCtx, serviceCtx).
		CreateTradeCorrection(&option.CreateTradeCorrectionReq{
			TenantId: p0AssetE2ETenantID, TradeId: trade.Id,
			Action:      option.TradeCorrectionAction_TRADE_CORRECTION_ACTION_CASH_ADJUSTMENT,
			Reason:      "invalid outsider ledger must be rejected",
			EvidenceRef: "s3://option-evidence/p1-trade-correction-invalid.json",
			Legs: []*option.TradeCorrectionLegInput{
				{UserId: buyerUserID, AccountId: buyerAccountID, Coin: "USDT", Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT, Amount: "1"},
				{UserId: invalidUserID, AccountId: invalidAccountID, Coin: "USDT", Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT, Amount: "1"},
			},
		})
	if err != nil {
		t.Fatalf("reject invalid trade correction ledger: %v", err)
	}
	if invalid == nil || invalid.Base == nil || invalid.Base.Code == 200 {
		t.Fatalf("invalid participant ledger accepted: %+v", invalid)
	}
	assertP1TradeCorrectionContractStatus(t, ctx, db, contract.Id, option.ContractStatus_CONTRACT_STATUS_TRADING)

	created, err := adminlogic.NewCreateTradeCorrectionLogic(requesterCtx, serviceCtx).
		CreateTradeCorrection(&option.CreateTradeCorrectionReq{
			TenantId: p0AssetE2ETenantID, TradeId: trade.Id,
			Action:      option.TradeCorrectionAction_TRADE_CORRECTION_ACTION_CASH_ADJUSTMENT,
			Reason:      "confirmed erroneous premium execution",
			EvidenceRef: "s3://option-evidence/p1-trade-correction.json#sha256=test",
			Legs: []*option.TradeCorrectionLegInput{
				{UserId: buyerUserID, AccountId: buyerAccountID, Coin: "USDT", Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT, Amount: "7"},
				{UserId: sellerUserID, AccountId: sellerAccountID, Coin: "USDT", Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_DEBIT, Amount: "3"},
				{UserId: feeUserID, AccountId: feeAccountID, Coin: "USDT", Direction: option.TradeCorrectionLegDirection_TRADE_CORRECTION_LEG_DIRECTION_CREDIT, Amount: "10"},
			},
		})
	if err != nil {
		t.Fatalf("create trade correction: %v", err)
	}
	if created == nil || created.Base == nil || created.Base.Code != 200 || created.Data == nil {
		t.Fatalf("unexpected create trade correction response: %+v", created)
	}
	correctionID, caseNo := created.Data.Id, created.Data.CaseNo
	assertP1TradeCorrectionContractStatus(t, ctx, db, contract.Id, option.ContractStatus_CONTRACT_STATUS_PAUSED)
	assertP1TradeCorrectionOrderStatus(t, ctx, db, activeOrder.Id, option.OrderStatus_ORDER_STATUS_CANCELING)
	if got := p1TradeCorrectionTradeHash(t, ctx, db, trade.Id); got != originalTradeHash {
		t.Fatalf("original trade changed while opening correction: got=%s want=%s", got, originalTradeHash)
	}

	selfReview, err := adminlogic.NewReviewTradeCorrectionLogic(requesterCtx, serviceCtx).
		ReviewTradeCorrection(&option.ReviewTradeCorrectionReq{
			TenantId: p0AssetE2ETenantID, CorrectionId: correctionID,
			Approve: true, Reason: "self review must fail",
		})
	if err != nil {
		t.Fatalf("self review request: %v", err)
	}
	if selfReview == nil || selfReview.Base == nil || selfReview.Base.Code == 200 {
		t.Fatalf("requester self review accepted: %+v", selfReview)
	}

	processAssetInstructions(t, ctx, serviceCtx)
	assertP1TradeCorrectionOrderStatus(t, ctx, db, activeOrder.Id, option.OrderStatus_ORDER_STATUS_CANCELED)
	assertWalletAmounts(t, ctx, db, activeUserID, "20.000000000000000000", "20.000000000000000000", "0.000000000000000000")

	var wg sync.WaitGroup
	results := make(chan bool, 20)
	for index := int64(0); index < 20; index++ {
		wg.Add(1)
		go func(operatorID int64) {
			defer wg.Done()
			response, reviewErr := adminlogic.NewReviewTradeCorrectionLogic(
				p0AdminContext(ctx, operatorID, p0AssetE2ETenantID), serviceCtx,
			).ReviewTradeCorrection(&option.ReviewTradeCorrectionReq{
				TenantId: p0AssetE2ETenantID, CorrectionId: correctionID,
				Approve: true, Reason: fmt.Sprintf("independent review %d", operatorID),
			})
			results <- reviewErr == nil && response != nil && response.Base != nil && response.Base.Code == 200
		}(reviewerIDBase + index)
	}
	wg.Wait()
	close(results)
	winners := 0
	for won := range results {
		if won {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("concurrent independent review winners=%d want=1", winners)
	}
	assertP1TradeCorrectionInstructionState(t, ctx, db, caseNo, 3, 0, 0, 0, 3)

	legs, err := serviceCtx.OptionTradeCorrectionLegModel.FindByCorrection(
		ctx, p0AssetE2ETenantID, correctionID,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(legs) != 3 {
		t.Fatalf("correction legs=%d want=3", len(legs))
	}
	faultClient := &failOnceSubAvailableClient{AssetClient: assetClient}
	faultClient.setTarget(legs[0].InstructionNo)
	serviceCtx.AssetClient = faultClient
	processAssetInstructions(t, ctx, serviceCtx)
	serviceCtx.AssetClient = assetClient
	if faultClient.failureCount() != 1 {
		t.Fatalf("committed debit response losses=%d want=1", faultClient.failureCount())
	}
	assertWalletAmounts(t, ctx, db, buyerUserID, "13.000000000000000000", "13.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerUserID, "2.000000000000000000", "2.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	// The first debit committed in Asset before its response was lost, so the
	// authoritative flow already exists even though Option still marks it failed.
	assertP1TradeCorrectionInstructionState(t, ctx, db, caseNo, 3, 0, 2, 1, 1)

	resetP1TradeCorrectionFailedInstructions(t, ctx, db, caseNo)
	processAssetInstructions(t, ctx, serviceCtx)
	assertWalletAmounts(t, ctx, db, buyerUserID, "13.000000000000000000", "13.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	assertP1TradeCorrectionInstructionState(t, ctx, db, caseNo, 3, 1, 1, 1, 1)

	creditAsset(t, ctx, assetClient, sellerUserID, "3", "P1-TRADE-CORRECTION-SELLER-CURE")
	resetP1TradeCorrectionFailedInstructions(t, ctx, db, caseNo)
	processAssetInstructions(t, ctx, serviceCtx)
	assertWalletAmounts(t, ctx, db, feeUserID, "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)

	assertWalletAmounts(t, ctx, db, buyerUserID, "13.000000000000000000", "13.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerUserID, "2.000000000000000000", "2.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "11.000000000000000000", "11.000000000000000000", "0.000000000000000000")
	assertP1TradeCorrectionCompleted(t, ctx, db, correctionID, caseNo, requesterID)
	if got := p1TradeCorrectionTradeHash(t, ctx, db, trade.Id); got != originalTradeHash {
		t.Fatalf("original trade changed after correction completion: got=%s want=%s", got, originalTradeHash)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP1TradeCorrectionCompleted(t, ctx, db, correctionID, caseNo, requesterID)
}

func insertP1TradeCorrectionContract(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext,
	feeUserID, feeAccountID, now int64,
) *models.TOptionContract {
	t.Helper()
	contract := &models.TOptionContract{
		TenantId: p0AssetE2ETenantID, ContractCode: "P1-TRADE-CORRECTION-CALL",
		UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		ExerciseStyle:  int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		StrikePrice:    decimal.NewFromInt(100), ContractUnit: decimal.NewFromInt(1),
		MinOrderQty: decimal.NewFromInt(1), MaxOrderQty: decimal.NewFromInt(1000),
		PriceTick: decimal.RequireFromString("0.1"), QtyStep: decimal.NewFromInt(1),
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
		ExerciseFeeRate: decimal.RequireFromString("0.1"), FeeUserId: feeUserID, FeeAccountId: feeAccountID,
		SellerMarginMode:      int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
		InitialMarginRate:     decimal.RequireFromString("0.5"),
		MaintenanceMarginRate: decimal.RequireFromString("0.2"), MinMarginRate: decimal.RequireFromString("0.1"),
		TradingCalendarCode: "CONTINUOUS_24_7", Status: int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		IsDeleted: int64(common.YesNo_YES_NO_NO), CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert trade correction contract: %v", err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func insertP1TradeCorrectionTrade(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, contractID,
	buyerUserID, buyerAccountID, sellerUserID, sellerAccountID, now int64,
) *models.TOptionTrade {
	t.Helper()
	trade := &models.TOptionTrade{
		TenantId: p0AssetE2ETenantID, TradeNo: "P1-TRADE-CORRECTION-ORIGINAL",
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT",
		BuyOrderId: 990021, BuyOrderNo: "P1-TRADE-CORRECTION-BUY",
		BuyUserId: buyerUserID, BuyAccountId: buyerAccountID,
		SellOrderId: 990022, SellOrderNo: "P1-TRADE-CORRECTION-SELL",
		SellUserId: sellerUserID, SellAccountId: sellerAccountID,
		Price: decimal.NewFromInt(10), Qty: decimal.NewFromInt(1), Turnover: decimal.NewFromInt(10),
		BuyFee: decimal.RequireFromString("0.4"), SellFee: decimal.RequireFromString("0.2"),
		FeeCoin: "USDT", MakerSide: int64(common.Side_SIDE_SELL), MatchSequence: 1,
		TradeTime: now - 60, CreateTimes: now - 60,
	}
	result, err := serviceCtx.OptionTradeModel.Insert(ctx, trade)
	if err != nil {
		t.Fatalf("insert original trade: %v", err)
	}
	trade.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return trade
}

func p1TradeCorrectionTradeHash(t *testing.T, ctx context.Context, db *sql.DB, tradeID int64) string {
	t.Helper()
	var hash string
	if err := db.QueryRowContext(ctx, `SELECT SHA2(CONCAT_WS('|',
		tenant_id,trade_no,combo_match_no,combo_leg_no,contract_id,underlying_symbol,
		buy_order_id,buy_order_no,buy_user_id,buy_account_id,
		sell_order_id,sell_order_no,sell_user_id,sell_account_id,
		price,qty,turnover,buy_fee,sell_fee,fee_coin,maker_side,match_sequence,trade_time,create_times
	),256) FROM t_option_trade WHERE id=?`, tradeID).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	return hash
}

func assertP1TradeCorrectionContractStatus(
	t *testing.T, ctx context.Context, db *sql.DB, contractID int64, want option.ContractStatus,
) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_contract WHERE id=?`, contractID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != int64(want) {
		t.Fatalf("trade correction contract status=%d want=%d", got, want)
	}
}

func assertP1TradeCorrectionOrderStatus(
	t *testing.T, ctx context.Context, db *sql.DB, orderID int64, want option.OrderStatus,
) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_order WHERE id=?`, orderID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != int64(want) {
		t.Fatalf("trade correction order status=%d want=%d", got, want)
	}
}

func resetP1TradeCorrectionFailedInstructions(
	t *testing.T, ctx context.Context, db *sql.DB, caseNo string,
) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET next_retry_at=0 WHERE tenant_id=? AND biz_no=? AND status=?`,
		p0AssetE2ETenantID, caseNo,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
	); err != nil {
		t.Fatal(err)
	}
}

func assertP1TradeCorrectionInstructionState(
	t *testing.T, ctx context.Context, db *sql.DB, caseNo string,
	total, success, failed, flows, pending int64,
) {
	t.Helper()
	var gotTotal, gotSuccess, gotFailed, gotFlows, gotPending int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		COALESCE(SUM(status=?),0),COALESCE(SUM(status=?),0),COALESCE(SUM(status=?),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		p0AssetE2ETenantID, caseNo,
	).Scan(&gotTotal, &gotSuccess, &gotFailed, &gotPending); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		p0AssetE2ETenantID, caseNo,
	).Scan(&gotFlows); err != nil {
		t.Fatal(err)
	}
	if gotTotal != total || gotSuccess != success || gotFailed != failed || gotFlows != flows || gotPending != pending {
		t.Fatalf("trade correction instructions total/success/failed/flows/pending=%d/%d/%d/%d/%d want=%d/%d/%d/%d/%d",
			gotTotal, gotSuccess, gotFailed, gotFlows, gotPending,
			total, success, failed, flows, pending)
	}
}

func assertP1TradeCorrectionCompleted(
	t *testing.T, ctx context.Context, db *sql.DB, correctionID int64, caseNo string, requesterID int64,
) {
	t.Helper()
	var status, completedAt, requestedBy, reviewedBy, instructions, success, reconciled, flows, duplicateFlows int64
	if err := db.QueryRowContext(ctx, `SELECT status,completed_at,requested_by,reviewed_by
		FROM t_option_trade_correction WHERE tenant_id=? AND id=? AND case_no=?`,
		p0AssetE2ETenantID, correctionID, caseNo,
	).Scan(&status, &completedAt, &requestedBy, &reviewedBy); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=?),0),
		COALESCE(SUM(reconciliation_status=?),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND biz_no=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, caseNo,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		p0AssetE2ETenantID, caseNo,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (
		SELECT flow.biz_no FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.biz_no=?
		GROUP BY flow.biz_no HAVING COUNT(*)>1
	) duplicate_flow`, p0AssetE2ETenantID, caseNo).Scan(&duplicateFlows); err != nil {
		t.Fatal(err)
	}
	var correctionNet string
	if err := db.QueryRowContext(ctx, `SELECT CAST(COALESCE(SUM(CASE
		WHEN flow.op_type=? THEN flow.change_amount
		WHEN flow.op_type=? THEN -flow.change_amount ELSE 0 END),0) AS CHAR)
		FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		int64(asset.AssetOpType_ASSET_OP_TYPE_ADD), int64(asset.AssetOpType_ASSET_OP_TYPE_SUB),
		p0AssetE2ETenantID, caseNo,
	).Scan(&correctionNet); err != nil {
		t.Fatal(err)
	}
	var createdEvents, approvedEvents, completedEvents int64
	if err := db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(event_type='TRADE_CORRECTION_CREATED'),0),
		COALESCE(SUM(event_type='TRADE_CORRECTION_APPROVED'),0),
		COALESCE(SUM(event_type='TRADE_CORRECTION_COMPLETED'),0)
		FROM t_option_trading_control_event
		WHERE tenant_id=? AND contract_id=? AND detail LIKE ?`,
		p0AssetE2ETenantID, p1TradeCorrectionContractID(t, ctx, db, correctionID), "%case_no="+caseNo+"%",
	).Scan(&createdEvents, &approvedEvents, &completedEvents); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_COMPLETED) ||
		completedAt <= 0 || requestedBy != requesterID || reviewedBy == 0 || reviewedBy == requesterID ||
		instructions != 3 || success != 3 || reconciled != 3 || flows != 3 || duplicateFlows != 0 ||
		correctionNet != "0.000000000000000000" || createdEvents != 1 || approvedEvents != 1 || completedEvents != 1 {
		t.Fatalf("trade correction completion status/completed/requested/reviewed/instructions/success/reconciled/flows/duplicates/net/events=%d/%d/%d/%d/%d/%d/%d/%d/%d/%s/%d,%d,%d",
			status, completedAt, requestedBy, reviewedBy, instructions, success, reconciled, flows,
			duplicateFlows, correctionNet, createdEvents, approvedEvents, completedEvents)
	}
}

func p1TradeCorrectionContractID(
	t *testing.T, ctx context.Context, db *sql.DB, correctionID int64,
) int64 {
	t.Helper()
	var contractID int64
	if err := db.QueryRowContext(ctx, `SELECT contract_id FROM t_option_trade_correction WHERE id=?`, correctionID).
		Scan(&contractID); err != nil {
		t.Fatal(err)
	}
	return contractID
}
