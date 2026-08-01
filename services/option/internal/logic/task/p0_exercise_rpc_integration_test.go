package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
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

func testP0AmericanExerciseConcurrencyFIFO(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		longUserID   int64 = 117
		shortAUserID int64 = 118
		shortBUserID int64 = 119
		feeUserID    int64 = 120
	)
	now := time.Now().Unix()
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-AMERICAN-EARLY-CALL",
		option.ExerciseStyle_EXERCISE_STYLE_AMERICAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		now-3600, now+3600, now+7200, now+7200,
		common.YesNo_YES_NO_NO, feeUserID, 9010,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "140", "40", now)
	creditAsset(t, ctx, assetClient, longUserID, "100", "P0-AMERICAN-LONG-SEED")
	creditAsset(t, ctx, assetClient, shortAUserID, "100", "P0-AMERICAN-SHORT-A-SEED")
	creditAsset(t, ctx, assetClient, shortBUserID, "200", "P0-AMERICAN-SHORT-B-SEED")
	transferP0OptionPremium(t, ctx, assetClient, longUserID, shortAUserID, "10", "P0-AMERICAN-PREMIUM-A")
	transferP0OptionPremium(t, ctx, assetClient, longUserID, shortBUserID, "20", "P0-AMERICAN-PREMIUM-B")

	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: 7010,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(3),
		AvailableQty: decimal.NewFromInt(3), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(120),
		ExerciseableQty: decimal.NewFromInt(3),
		Status:          int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes:     now - 300, UpdateTimes: now - 300,
	})
	shortA := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortAUserID, AccountId: 8010,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(40),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	shortB := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortBUserID, AccountId: 8011,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(80),
		MarginAmount: decimal.NewFromInt(100), MaintenanceMargin: decimal.NewFromInt(40),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	lotA := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortA, "P0-AMERICAN-SHORT-A-MARGIN", "1", "50", now-190,
	)
	lotB := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortB, "P0-AMERICAN-SHORT-B-MARGIN", "2", "100", now-190,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, shortA, lotA, "50")
	freezeP0ExerciseMargin(t, ctx, assetClient, shortB, lotB, "100")

	exerciseCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(longUserID),
	))
	req := &option.ExerciseReq{
		AccountId: 7010, ContractId: contract.Id, PositionId: longPosition.Id,
		ExerciseQty: "3", ClientExerciseId: "P0-AMERICAN-EXERCISE-CONCURRENT",
	}
	type exerciseResult struct {
		resp *option.ExerciseResp
		err  error
	}
	results := make(chan exerciseResult, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := applogic.NewExerciseLogic(exerciseCtx, serviceCtx).Exercise(req)
			results <- exerciseResult{resp: resp, err: err}
		}()
	}
	wg.Wait()
	close(results)
	var exerciseID int64
	var exerciseNo string
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent exercise failed: %v", result.err)
		}
		if result.resp == nil || result.resp.GetBase().GetCode() != 200 || result.resp.Data == nil {
			t.Fatalf("concurrent exercise rejected: %+v", result.resp)
		}
		if exerciseID == 0 {
			exerciseID = result.resp.Data.ExerciseId
			exerciseNo = result.resp.Data.ExerciseNo
		}
		if result.resp.Data.ExerciseId != exerciseID || result.resp.Data.ExerciseNo != exerciseNo {
			t.Fatalf("exercise replay identity changed: %+v want=%d/%s", result.resp.Data, exerciseID, exerciseNo)
		}
	}
	assertP0ExerciseReservation(t, ctx, db, contract.Id, longPosition.Id, exerciseID)

	changed := &option.ExerciseReq{
		AccountId: req.AccountId, ContractId: req.ContractId, PositionId: req.PositionId,
		ExerciseQty: "2", ClientExerciseId: req.ClientExerciseId,
	}
	changedResp, err := applogic.NewExerciseLogic(exerciseCtx, serviceCtx).Exercise(changed)
	if err != nil {
		t.Fatal(err)
	}
	if changedResp == nil || changedResp.GetBase().GetCode() == 200 {
		t.Fatalf("same exercise key accepted different quantity: %+v", changedResp)
	}
	assertP0ExerciseReservation(t, ctx, db, contract.Id, longPosition.Id, exerciseID)

	exercise, err := serviceCtx.OptionExerciseModel.FindOne(ctx, exerciseID)
	if err != nil {
		t.Fatal(err)
	}
	clearingErrors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clearingErrors <- NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise)
		}()
	}
	wg.Wait()
	close(clearingErrors)
	for clearingErr := range clearingErrors {
		if clearingErr != nil {
			t.Fatalf("concurrent exercise clearing failed: %v", clearingErr)
		}
	}
	assertP0ExerciseClearingCreated(t, ctx, db, exerciseID, exerciseNo, shortA.Id, shortB.Id)

	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0ExerciseCompleted(t, ctx, db, exerciseID, exerciseNo)
	assertWalletAmounts(t, ctx, db, longUserID, "178.000000000000000000", "178.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, shortAUserID, "70.000000000000000000", "70.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, shortBUserID, "140.000000000000000000", "140.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "12.000000000000000000", "12.000000000000000000", "0.000000000000000000")
	assertP0ExercisePosition(t, ctx, db, longPosition.Id, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", option.PositionStatus_POSITION_STATUS_EXERCISED)
	assertP0ExercisePosition(t, ctx, db, shortA.Id, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", option.PositionStatus_POSITION_STATUS_EXERCISED)
	assertP0ExercisePosition(t, ctx, db, shortB.Id, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", option.PositionStatus_POSITION_STATUS_EXERCISED)
	assertP0ExerciseLot(t, ctx, db, lotA.Id, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED)
	assertP0ExerciseLot(t, ctx, db, lotB.Id, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED)
	assertP0ExerciseReturn(t, ctx, db, longPosition.Id, "90.0000000000000000", "12.0000000000000000", "78.0000000000000000")
	assertP0ExerciseReturn(t, ctx, db, shortA.Id, "-30.0000000000000000", "0.0000000000000000", "-30.0000000000000000")
	assertP0ExerciseReturn(t, ctx, db, shortB.Id, "-60.0000000000000000", "0.0000000000000000", "-60.0000000000000000")
	assertP0WalletTotal(t, ctx, db, []int64{longUserID, shortAUserID, shortBUserID, feeUserID}, "400.000000000000000000", "400.000000000000000000", "0.000000000000000000")

	replayResp, err := applogic.NewExerciseLogic(exerciseCtx, serviceCtx).Exercise(req)
	if err != nil || replayResp.GetBase().GetCode() != 200 || replayResp.Data.ExerciseId != exerciseID {
		t.Fatalf("completed exercise replay failed: resp=%+v err=%v", replayResp, err)
	}
	if err := NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise); err != nil {
		t.Fatalf("completed clearing replay: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0ExerciseCompleted(t, ctx, db, exerciseID, exerciseNo)
}

func testP0ExpiryAutoDNEActualAssignment(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		autoUserID  int64 = 121
		dneUserID   int64 = 122
		shortUserID int64 = 123
		feeUserID   int64 = 124
	)
	now := time.Now().Unix()
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-EXPIRY-AUTO-DNE-CALL",
		option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
		option.ContractStatus_CONTRACT_STATUS_EXPIRED,
		now-3600, now-20, now-10, now-1,
		common.YesNo_YES_NO_YES, feeUserID, 9011,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "120", "20", now)
	creditAsset(t, ctx, assetClient, autoUserID, "100", "P0-EXPIRY-AUTO-SEED")
	creditAsset(t, ctx, assetClient, dneUserID, "100", "P0-EXPIRY-DNE-SEED")
	creditAsset(t, ctx, assetClient, shortUserID, "200", "P0-EXPIRY-SHORT-SEED")
	transferP0OptionPremium(t, ctx, assetClient, autoUserID, shortUserID, "10", "P0-EXPIRY-AUTO-PREMIUM")
	transferP0OptionPremium(t, ctx, assetClient, dneUserID, shortUserID, "10", "P0-EXPIRY-DNE-PREMIUM")

	autoLong := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: autoUserID, AccountId: 7020,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(20), PositionValue: decimal.NewFromInt(20),
		ExerciseableQty: decimal.NewFromInt(1), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 300, UpdateTimes: now - 300,
	})
	dneLong := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: dneUserID, AccountId: 7021,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(20), PositionValue: decimal.NewFromInt(20),
		ExerciseableQty: decimal.NewFromInt(1), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 290, UpdateTimes: now - 290,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: 8020,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(2),
		AvailableQty: decimal.NewFromInt(2), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(20), PositionValue: decimal.NewFromInt(40),
		MarginAmount: decimal.NewFromInt(100), MaintenanceMargin: decimal.NewFromInt(40),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-EXPIRY-SHORT-MARGIN", "2", "100", now-190,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, shortPosition, lot, "100")

	autoInstruction := insertP0ExerciseInstruction(
		t, ctx, serviceCtx, autoLong, "P0-EXPIRY-AUTO", option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
		1, option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE, 0, contract.ExerciseCutoffTime, now-30,
	)
	dnePrior := insertP0ExerciseInstruction(
		t, ctx, serviceCtx, dneLong, "P0-EXPIRY-DNE-PRIOR-AUTO", option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
		1, option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_SUPERSEDED, 0, contract.ExerciseCutoffTime, now-40,
	)
	dneInstruction := insertP0ExerciseInstruction(
		t, ctx, serviceCtx, dneLong, "P0-EXPIRY-DNE", option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
		2, option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE, dnePrior.Id, contract.ExerciseCutoffTime, now-30,
	)
	assertP0ExerciseInstructionImmutable(t, ctx, db, autoInstruction.Id, dnePrior.Id, dneInstruction.Id)

	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contract.Id, contract.ExpireTime, now,
		fmt.Sprintf("P0-EXPIRY-%d", contract.Id), []string{"119", "120", "121"}, "120",
	)
	logic := NewProcessContractLifecycleLogic(ctx, serviceCtx)
	if err := logic.processExpiredContracts(now); err != nil {
		t.Fatalf("process mixed AUTO/DNE expiry: %v", err)
	}
	assertP0ExpiryCreated(t, ctx, db, contract.Id, autoLong.Id, dneLong.Id, shortPosition.Id)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0ExpiryCompleted(t, ctx, db, contract.Id, autoLong.Id, dneLong.Id, shortPosition.Id, lot.Id)
	assertWalletAmounts(t, ctx, db, autoUserID, "108.000000000000000000", "108.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, dneUserID, "90.000000000000000000", "90.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, shortUserID, "200.000000000000000000", "200.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID, "2.000000000000000000", "2.000000000000000000", "0.000000000000000000")
	assertP0ExerciseReturn(t, ctx, db, autoLong.Id, "10.0000000000000000", "2.0000000000000000", "8.0000000000000000")
	assertP0ExerciseReturn(t, ctx, db, dneLong.Id, "-10.0000000000000000", "0.0000000000000000", "-10.0000000000000000")
	assertP0ExerciseReturn(t, ctx, db, shortPosition.Id, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000")
	assertP0WalletTotal(t, ctx, db, []int64{autoUserID, dneUserID, shortUserID, feeUserID}, "400.000000000000000000", "400.000000000000000000", "0.000000000000000000")
	if err := logic.processExpiredContracts(now); err != nil {
		t.Fatalf("replay mixed AUTO/DNE expiry: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0ExpiryCompleted(t, ctx, db, contract.Id, autoLong.Id, dneLong.Id, shortPosition.Id, lot.Id)
}

func insertP0ExerciseContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	code string,
	style option.ExerciseStyle,
	status option.ContractStatus,
	listTime, cutoffTime, expireTime, deliverTime int64,
	autoExercise common.YesNo,
	feeUserID, feeAccountID int64,
) *models.TOptionContract {
	t.Helper()
	now := time.Now().Unix()
	contract := &models.TOptionContract{
		TenantId: p0AssetE2ETenantID, ContractCode: code,
		UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType: int64(option.OptionType_OPTION_TYPE_CALL), ExerciseStyle: int64(style),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_CASH), StrikePrice: decimal.NewFromInt(100),
		ContractUnit: decimal.NewFromInt(1), MinOrderQty: decimal.RequireFromString("0.5"),
		MaxOrderQty: decimal.NewFromInt(1000), PriceTick: decimal.RequireFromString("0.1"),
		QtyStep: decimal.RequireFromString("0.5"), Multiplier: decimal.NewFromInt(1),
		ListTime: listTime, ExerciseCutoffTime: cutoffTime, ExpireTime: expireTime, DeliverTime: deliverTime,
		AutoExerciseThreshold: decimal.NewFromInt(10), MaxUserLongQty: decimal.NewFromInt(10000),
		MaxUserShortQty: decimal.NewFromInt(10000), MaxOpenInterest: decimal.NewFromInt(10000),
		OrderPriceBandRatio: decimal.RequireFromString("0.2"), CircuitBreakerRatio: decimal.RequireFromString("0.5"),
		GreeksMaxAgeSeconds: 60, SettlementPriceSource: "authoritative-market",
		SettlementPriceMethod: "MEDIAN", SettlementWindowSeconds: 60, SettlementMinSamples: 3,
		IsAutoExercise: int64(autoExercise), ExerciseFeeRate: decimal.RequireFromString("0.1"),
		FeeUserId: feeUserID, FeeAccountId: feeAccountID,
		SellerMarginMode:  int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
		InitialMarginRate: decimal.RequireFromString("0.5"), MaintenanceMarginRate: decimal.RequireFromString("0.2"),
		MinMarginRate: decimal.RequireFromString("0.1"), TradingCalendarCode: "CONTINUOUS_24_7",
		Status: int64(status), IsDeleted: int64(common.YesNo_YES_NO_NO),
		CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert exercise contract %s: %v", code, err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func insertP0ExerciseMarket(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractID int64,
	underlyingPrice, markPrice string,
	now int64,
) {
	t.Helper()
	_, err := serviceCtx.OptionMarketModel.Insert(ctx, &models.TOptionMarket{
		TenantId: p0AssetE2ETenantID, ContractId: contractID,
		UnderlyingPrice: decimal.RequireFromString(underlyingPrice), MarkPrice: decimal.RequireFromString(markPrice),
		LastPrice: decimal.RequireFromString(markPrice), BidPrice: decimal.RequireFromString(markPrice),
		AskPrice: decimal.RequireFromString(markPrice), TheoreticalPrice: decimal.RequireFromString(markPrice),
		IntrinsicValue: decimal.RequireFromString(markPrice), Iv: decimal.RequireFromString("0.5"),
		SnapshotTime: now, UnderlyingSnapshotTime: now, MarkSnapshotTime: now, GreeksSnapshotTime: now,
		CreateTimes: now, UpdateTimes: now,
	})
	if err != nil {
		t.Fatalf("insert exercise market: %v", err)
	}
}

func insertP0ExerciseMarginLot(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	position *models.TOptionPosition,
	freezeBizNo, quantity, margin string,
	createTimes int64,
) *models.TOptionMarginLot {
	t.Helper()
	lot := &models.TOptionMarginLot{
		TenantId: position.TenantId, UserId: position.UserId, AccountId: position.AccountId,
		ContractId: position.ContractId, PositionId: position.Id,
		OriginContractId: position.ContractId, OriginPositionId: position.Id,
		TradeId: -position.Id, FreezeBizNo: freezeBizNo, CollateralCoin: "USDT",
		Quantity: decimal.RequireFromString(quantity), RemainingQuantity: decimal.RequireFromString(quantity),
		InitialMargin: decimal.RequireFromString(margin), RemainingMargin: decimal.RequireFromString(margin),
		Status: int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE), CreateTimes: createTimes, UpdateTimes: createTimes,
	}
	result, err := serviceCtx.OptionMarginLotModel.Insert(ctx, lot)
	if err != nil {
		t.Fatalf("insert exercise margin lot: %v", err)
	}
	lot.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return lot
}

func freezeP0ExerciseMargin(
	t *testing.T,
	ctx context.Context,
	assetClient asset.AssetClient,
	position *models.TOptionPosition,
	lot *models.TOptionMarginLot,
	amount string,
) {
	t.Helper()
	resp, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: position.TenantId, UserId: position.UserId,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: lot.Id, BizNo: lot.FreezeBizNo, Remark: "P0 exercise short margin",
	})
	assertAssetOK(t, resp, err)
}

func transferP0OptionPremium(
	t *testing.T,
	ctx context.Context,
	assetClient asset.AssetClient,
	payerUserID, payeeUserID int64,
	amount, bizPrefix string,
) {
	t.Helper()
	debitResp, err := assetClient.SubAvailable(ctx, &asset.SubAvailableReq{
		TenantId: p0AssetE2ETenantID, UserId: payerUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
		BizNo: bizPrefix + "-DEBIT", Remark: "P0 option opening premium debit",
	})
	assertAssetOK(t, debitResp, err)
	creditResp, err := assetClient.AddAvailable(ctx, &asset.AddAvailableReq{
		TenantId: p0AssetE2ETenantID, UserId: payeeUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
		BizNo: bizPrefix + "-CREDIT", Remark: "P0 option opening premium credit",
	})
	assertAssetOK(t, creditResp, err)
}

func insertP0ExerciseInstruction(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	position *models.TOptionPosition,
	clientID string,
	instructionType option.ExerciseInstructionType,
	version int64,
	status option.ExerciseInstructionStatus,
	supersedesID, cutoffTime, createTimes int64,
) *models.TOptionExerciseInstruction {
	t.Helper()
	item := &models.TOptionExerciseInstruction{
		TenantId: position.TenantId, UserId: position.UserId, AccountId: position.AccountId,
		ContractId: position.ContractId, PositionId: position.Id, ClientInstructionId: clientID,
		InstructionType: int64(instructionType), Version: version, Status: int64(status),
		SupersedesId: supersedesID, CutoffTime: cutoffTime, CreateTimes: createTimes, UpdateTimes: createTimes,
	}
	result, err := serviceCtx.OptionExerciseInstructionModel.Insert(ctx, item)
	if err != nil {
		t.Fatalf("insert exercise instruction %s: %v", clientID, err)
	}
	item.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func assertP0ExerciseReservation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, positionID, exerciseID int64,
) {
	t.Helper()
	var count, minID, maxID int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(MIN(id),0),COALESCE(MAX(id),0)
		FROM t_option_exercise WHERE tenant_id=? AND contract_id=? AND client_exercise_id=?`,
		p0AssetE2ETenantID, contractID, "P0-AMERICAN-EXERCISE-CONCURRENT",
	).Scan(&count, &minID, &maxID); err != nil {
		t.Fatal(err)
	}
	if count != 1 || minID != exerciseID || maxID != exerciseID {
		t.Fatalf("exercise reservation count/id=%d/%d/%d want=1/%d/%d", count, minID, maxID, exerciseID, exerciseID)
	}
	var available, frozen string
	if err := db.QueryRowContext(ctx, `SELECT CAST(available_qty AS CHAR),CAST(frozen_qty AS CHAR)
		FROM t_option_position WHERE id=?`, positionID).Scan(&available, &frozen); err != nil {
		t.Fatal(err)
	}
	if available != "0.0000000000000000" || frozen != "3.0000000000000000" {
		t.Fatalf("exercise reservation available/frozen=%s/%s want=0/3", available, frozen)
	}
}

func assertP0ExerciseClearingCreated(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
	shortAID, shortBID int64,
) {
	t.Helper()
	rows, err := db.QueryContext(ctx, `SELECT short_position_id,CAST(quantity AS CHAR),CAST(payoff AS CHAR),status
		FROM t_option_exercise_assignment WHERE tenant_id=? AND exercise_id=? ORDER BY id`,
		p0AssetE2ETenantID, exerciseID,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	type assignment struct {
		positionID int64
		quantity   string
		payoff     string
		status     int64
	}
	var got []assignment
	for rows.Next() {
		var item assignment
		if err := rows.Scan(&item.positionID, &item.quantity, &item.payoff, &item.status); err != nil {
			t.Fatal(err)
		}
		got = append(got, item)
	}
	want := []assignment{
		{shortAID, "1.0000000000000000", "40.0000000000000000", int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING)},
		{shortBID, "2.0000000000000000", "80.0000000000000000", int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING)},
	}
	if len(got) != len(want) {
		t.Fatalf("exercise assignments=%+v want=%+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("exercise assignment[%d]=%+v want=%+v", i, got[i], want[i])
		}
	}
	var count, step1, step2 int64
	var amount string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(step_no=1),SUM(step_no=2),CAST(SUM(amount) AS CHAR)
		FROM t_option_asset_instruction WHERE tenant_id=? AND biz_no=?`,
		p0AssetE2ETenantID, exerciseNo,
	).Scan(&count, &step1, &step2, &amount); err != nil {
		t.Fatal(err)
	}
	if count != 6 || step1 != 4 || step2 != 2 || amount != "270.0000000000000000" {
		t.Fatalf("exercise instruction count/steps/amount=%d/%d/%d/%s want=6/4/2/270", count, step1, step2, amount)
	}
}

func assertP0ExerciseCompleted(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
) {
	t.Helper()
	var status, assignments, assignmentDone, instructions, success, reconciled, flows int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_exercise WHERE id=?`, exerciseID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=2) FROM t_option_exercise_assignment
		WHERE tenant_id=? AND exercise_id=?`, p0AssetE2ETenantID, exerciseID,
	).Scan(&assignments, &assignmentDone); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(i.status=3),SUM(i.reconciliation_status=2),COUNT(DISTINCT f.id)
		FROM t_option_asset_instruction i LEFT JOIN t_asset_flow f
		  ON f.tenant_id=i.tenant_id AND f.biz_no=i.instruction_no
		WHERE i.tenant_id=? AND i.biz_no=?`, p0AssetE2ETenantID, exerciseNo,
	).Scan(&instructions, &success, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.ExerciseStatus_EXERCISE_STATUS_DONE) || assignments != 2 || assignmentDone != 2 ||
		instructions != 6 || success != 6 || reconciled != 6 || flows != 6 {
		t.Fatalf("exercise completion status=%d assignments=%d/%d instructions=%d/%d/%d flows=%d",
			status, assignmentDone, assignments, success, reconciled, instructions, flows)
	}
}

func assertP0ExercisePosition(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	positionID int64,
	qty, available, frozen, margin, maintenance, exerciseable string,
	status option.PositionStatus,
) {
	t.Helper()
	var gotQty, gotAvailable, gotFrozen, gotMargin, gotMaintenance, gotExerciseable string
	var gotStatus int64
	if err := db.QueryRowContext(ctx, `SELECT CAST(position_qty AS CHAR),CAST(available_qty AS CHAR),
		CAST(frozen_qty AS CHAR),CAST(margin_amount AS CHAR),CAST(maintenance_margin AS CHAR),
		CAST(exerciseable_qty AS CHAR),status FROM t_option_position WHERE id=?`, positionID,
	).Scan(&gotQty, &gotAvailable, &gotFrozen, &gotMargin, &gotMaintenance, &gotExerciseable, &gotStatus); err != nil {
		t.Fatal(err)
	}
	if gotQty != qty || gotAvailable != available || gotFrozen != frozen || gotMargin != margin ||
		gotMaintenance != maintenance || gotExerciseable != exerciseable || gotStatus != int64(status) {
		t.Fatalf("position %d=%s/%s/%s/%s/%s/%s/%d want=%s/%s/%s/%s/%s/%s/%d",
			positionID, gotQty, gotAvailable, gotFrozen, gotMargin, gotMaintenance, gotExerciseable, gotStatus,
			qty, available, frozen, margin, maintenance, exerciseable, status)
	}
}

func assertP0ExerciseReturn(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	positionID int64,
	settlementPnL, feePaid, totalReturn string,
) {
	t.Helper()
	var gotTrade, gotSettlement, gotFee, gotTotal, gotRealized string
	if err := db.QueryRowContext(ctx, `SELECT CAST(trade_realized_pnl AS CHAR),
		CAST(settlement_realized_pnl AS CHAR),CAST(fee_paid AS CHAR),
		CAST(total_return AS CHAR),CAST(realized_pnl AS CHAR)
		FROM t_option_position WHERE id=?`, positionID,
	).Scan(&gotTrade, &gotSettlement, &gotFee, &gotTotal, &gotRealized); err != nil {
		t.Fatal(err)
	}
	if gotTrade != "0.0000000000000000" || gotSettlement != settlementPnL || gotFee != feePaid ||
		gotTotal != totalReturn || gotRealized != totalReturn {
		t.Fatalf("position return %d trade/settlement/fee/total/realized=%s/%s/%s/%s/%s want=0/%s/%s/%s/%s",
			positionID, gotTrade, gotSettlement, gotFee, gotTotal, gotRealized,
			settlementPnL, feePaid, totalReturn, totalReturn)
	}
}

func assertP0ExerciseLot(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	lotID int64,
	quantity, margin, pending string,
	status option.MarginLotStatus,
) {
	t.Helper()
	var gotQuantity, gotMargin, gotPending string
	var gotStatus int64
	if err := db.QueryRowContext(ctx, `SELECT CAST(remaining_quantity AS CHAR),CAST(remaining_margin AS CHAR),
		CAST(pending_margin AS CHAR),status FROM t_option_margin_lot WHERE id=?`, lotID,
	).Scan(&gotQuantity, &gotMargin, &gotPending, &gotStatus); err != nil {
		t.Fatal(err)
	}
	if gotQuantity != quantity || gotMargin != margin || gotPending != pending || gotStatus != int64(status) {
		t.Fatalf("margin lot %d=%s/%s/%s/%d want=%s/%s/%s/%d",
			lotID, gotQuantity, gotMargin, gotPending, gotStatus, quantity, margin, pending, status)
	}
}

func assertP0WalletTotal(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userIDs []int64,
	total, available, frozen string,
) {
	t.Helper()
	if len(userIDs) != 4 {
		t.Fatalf("wallet conservation helper requires four users")
	}
	var gotTotal, gotAvailable, gotFrozen string
	if err := db.QueryRowContext(ctx, `SELECT CAST(SUM(total_amount) AS CHAR),CAST(SUM(available_amount) AS CHAR),
		CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset
		WHERE tenant_id=? AND wallet_type=? AND coin='USDT' AND user_id IN (?,?,?,?)`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION),
		userIDs[0], userIDs[1], userIDs[2], userIDs[3],
	).Scan(&gotTotal, &gotAvailable, &gotFrozen); err != nil {
		t.Fatal(err)
	}
	if gotTotal != total || gotAvailable != available || gotFrozen != frozen {
		t.Fatalf("wallet conservation=%s/%s/%s want=%s/%s/%s", gotTotal, gotAvailable, gotFrozen, total, available, frozen)
	}
}

func assertP0ExerciseInstructionImmutable(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	autoID, supersededID, dneID int64,
) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_exercise_instruction SET instruction_type=2 WHERE id=?`, autoID); err == nil {
		t.Fatal("database accepted in-place exercise instruction economic change")
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_exercise_instruction SET status=1 WHERE id=?`, supersededID); err == nil {
		t.Fatal("database accepted superseded-to-active exercise instruction reversal")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_exercise_instruction WHERE id=?`, dneID); err == nil {
		t.Fatal("database accepted exercise instruction history deletion")
	}
	var count int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_exercise_instruction WHERE id IN (?,?,?)`,
		autoID, supersededID, dneID,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("exercise instruction history count=%d want=3", count)
	}
}

func assertP0ExpiryCreated(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, autoPositionID, dnePositionID, shortPositionID int64,
) {
	t.Helper()
	var exercises, autoExercises, dneExercises int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(position_id=?),SUM(position_id=?)
		FROM t_option_exercise WHERE tenant_id=? AND contract_id=?`,
		autoPositionID, dnePositionID, p0AssetE2ETenantID, contractID,
	).Scan(&exercises, &autoExercises, &dneExercises); err != nil {
		t.Fatal(err)
	}
	if exercises != 1 || autoExercises != 1 || dneExercises != 0 {
		t.Fatalf("expiry exercise rows=%d auto=%d dne=%d want=1/1/0", exercises, autoExercises, dneExercises)
	}
	var autoQty, autoPayoff, dneQty, dnePayoff, shortQty, shortPayoff string
	var autoDirection, dneDirection, shortDirection int64
	query := `SELECT CAST(quantity AS CHAR),CAST(payoff AS CHAR),direction
		FROM t_option_settlement_detail WHERE tenant_id=? AND contract_id=? AND position_id=?`
	if err := db.QueryRowContext(ctx, query, p0AssetE2ETenantID, contractID, autoPositionID).Scan(&autoQty, &autoPayoff, &autoDirection); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, query, p0AssetE2ETenantID, contractID, dnePositionID).Scan(&dneQty, &dnePayoff, &dneDirection); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, query, p0AssetE2ETenantID, contractID, shortPositionID).Scan(&shortQty, &shortPayoff, &shortDirection); err != nil {
		t.Fatal(err)
	}
	if autoQty != "1.0000000000000000" || autoPayoff != "20.0000000000000000" ||
		autoDirection != int64(option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_CREDIT) ||
		dneQty != "0.0000000000000000" || dnePayoff != "0.0000000000000000" ||
		dneDirection != int64(option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_ABANDON) ||
		shortQty != "1.0000000000000000" || shortPayoff != "20.0000000000000000" ||
		shortDirection != int64(option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_DEBIT) {
		t.Fatalf("expiry allocation auto=%s/%s/%d dne=%s/%s/%d short=%s/%s/%d",
			autoQty, autoPayoff, autoDirection, dneQty, dnePayoff, dneDirection, shortQty, shortPayoff, shortDirection)
	}
	var instructionCount, step1, step2 int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(step_no=1),SUM(step_no=2)
		FROM t_option_asset_instruction WHERE tenant_id=? AND biz_no=(
			SELECT settlement_no FROM t_option_settlement WHERE tenant_id=? AND contract_id=?)`,
		p0AssetE2ETenantID, p0AssetE2ETenantID, contractID,
	).Scan(&instructionCount, &step1, &step2); err != nil {
		t.Fatal(err)
	}
	if instructionCount != 4 || step1 != 1 || step2 != 3 {
		t.Fatalf("expiry instruction count/steps=%d/%d/%d want=4/1/3", instructionCount, step1, step2)
	}
}

func assertP0ExpiryCompleted(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, autoPositionID, dnePositionID, shortPositionID, lotID int64,
) {
	t.Helper()
	var settlementStatus, batchStatus, contractStatus, instructionCount, success, reconciled, flows int64
	var totalCredit, totalDebit string
	if err := db.QueryRowContext(ctx, `SELECT s.status,b.status,c.status,b.instruction_count,
		CAST(b.total_credit AS CHAR),CAST(b.total_debit AS CHAR),
		SUM(i.status=3),SUM(i.reconciliation_status=2),COUNT(DISTINCT f.id)
		FROM t_option_settlement s
		JOIN t_option_settlement_batch b ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
		JOIN t_option_contract c ON c.id=s.contract_id
		JOIN t_option_asset_instruction i ON i.tenant_id=s.tenant_id AND i.biz_no=s.settlement_no
		LEFT JOIN t_asset_flow f ON f.tenant_id=i.tenant_id AND f.biz_no=i.instruction_no
		WHERE s.tenant_id=? AND s.contract_id=? GROUP BY s.id,b.id,c.id`,
		p0AssetE2ETenantID, contractID,
	).Scan(&settlementStatus, &batchStatus, &contractStatus, &instructionCount,
		&totalCredit, &totalDebit, &success, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE) ||
		contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) ||
		instructionCount != 4 || success != 4 || reconciled != 4 || flows != 4 ||
		totalCredit != "20.0000000000000000" || totalDebit != "20.0000000000000000" {
		t.Fatalf("expiry completion settlement/batch/contract=%d/%d/%d instructions=%d/%d/%d flows=%d credit/debit=%s/%s",
			settlementStatus, batchStatus, contractStatus, success, reconciled, instructionCount, flows, totalCredit, totalDebit)
	}
	for _, positionID := range []int64{autoPositionID, dnePositionID, shortPositionID} {
		var qty, available, frozen, exerciseable, margin, maintenance string
		var status int64
		if err := db.QueryRowContext(ctx, `SELECT CAST(position_qty AS CHAR),CAST(available_qty AS CHAR),
			CAST(frozen_qty AS CHAR),CAST(exerciseable_qty AS CHAR),CAST(margin_amount AS CHAR),
			CAST(maintenance_margin AS CHAR),status FROM t_option_position WHERE id=?`, positionID,
		).Scan(&qty, &available, &frozen, &exerciseable, &margin, &maintenance, &status); err != nil {
			t.Fatal(err)
		}
		if qty != "0.0000000000000000" || available != "0.0000000000000000" ||
			frozen != "0.0000000000000000" || exerciseable != "0.0000000000000000" ||
			margin != "0.0000000000000000" || maintenance != "0.0000000000000000" ||
			status != int64(option.PositionStatus_POSITION_STATUS_SETTLED) {
			t.Fatalf("expiry position %d=%s/%s/%s/%s/%s/%s/%d", positionID, qty, available, frozen, exerciseable, margin, maintenance, status)
		}
	}
	assertP0ExerciseLot(t, ctx, db, lotID, "0.0000000000000000", "0.0000000000000000", "0.0000000000000000", option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED)
}
