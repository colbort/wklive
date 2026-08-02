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

type p0ExerciseCallResult struct {
	resp *option.ExerciseResp
	err  error
}

func testP0ExerciseCutoffAndLifecycleRace(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		cutoffUserID int64 = 125
		stateUserID  int64 = 126
		pausedUserID int64 = 127
	)

	// The public American exercise passes its optimistic validation, then waits
	// on the contract row until the exact published cutoff has elapsed. The
	// transactional recheck must reject it without freezing quantity or writing
	// an exercise row.
	now := time.Now().Unix()
	cutoff := now + 2
	cutoffContract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-EXERCISE-CUTOFF-CROSS",
		option.ExerciseStyle_EXERCISE_STYLE_AMERICAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		now-60, cutoff, cutoff+3600, cutoff+3600,
		common.YesNo_YES_NO_NO, 9201, 9201,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, cutoffContract.Id, "140", "40", now)
	cutoffPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: cutoffUserID, AccountId: 7125,
		ContractId: cutoffContract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), ExerciseableQty: decimal.NewFromInt(1),
		OpenAvgPrice: decimal.NewFromInt(10), MarkPrice: decimal.NewFromInt(40),
		PositionValue: decimal.NewFromInt(40),
		Status:        int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes:   now - 30, UpdateTimes: now - 30,
	})
	cutoffLock := lockP0ExerciseContract(t, ctx, db, cutoffContract.Id)
	cutoffCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(cutoffUserID),
	))
	cutoffReq := &option.ExerciseReq{
		AccountId: 7125, ContractId: cutoffContract.Id, PositionId: cutoffPosition.Id,
		ExerciseQty: "1", ClientExerciseId: "P0-EXERCISE-CUTOFF-CROSS",
	}
	cutoffResult := make(chan p0ExerciseCallResult, 1)
	startedAtMillis := utils.NowMillis()
	go func() {
		resp, err := applogic.NewExerciseLogic(cutoffCtx, serviceCtx).Exercise(cutoffReq)
		cutoffResult <- p0ExerciseCallResult{resp: resp, err: err}
	}()
	assertP0ExerciseCallBlocked(t, cutoffResult, 150*time.Millisecond)
	cutoffMillis := cutoff * 1000
	if wait := time.Until(time.UnixMilli(cutoffMillis)); wait > 0 {
		time.Sleep(wait + 2*time.Millisecond)
	}
	releasedAtMillis := utils.NowMillis()
	if err := cutoffLock.Commit(); err != nil {
		t.Fatalf("release cutoff contract lock: %v", err)
	}
	result := waitP0ExerciseResult(t, cutoffResult)
	if result.err == nil || result.resp != nil {
		t.Fatalf("cutoff-crossing exercise was not rejected: resp=%+v err=%v", result.resp, result.err)
	}
	if startedAtMillis >= cutoffMillis || releasedAtMillis < cutoffMillis {
		t.Fatalf("cutoff race did not cross boundary: start=%d cutoff=%d release=%d",
			startedAtMillis, cutoffMillis, releasedAtMillis)
	}
	assertP0RejectedExerciseMutation(
		t, ctx, db, cutoffContract.Id, cutoffPosition.Id, cutoffReq.ClientExerciseId,
	)

	// The independent last-trade boundary pauses the contract and cancels the
	// ordinary book, but must not mark the contract expired or start exercise /
	// settlement before the separately approved expiry time.
	lastTradeBoundary := time.Now().Unix()
	lastTradeContract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-LAST-TRADE-INDEPENDENT",
		option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		lastTradeBoundary-60, lastTradeBoundary+120,
		lastTradeBoundary+240, lastTradeBoundary+300,
		common.YesNo_YES_NO_YES, 9204, 9204,
	)
	if _, err := db.ExecContext(ctx, `UPDATE t_option_contract
		SET last_trade_time=?,update_times=? WHERE id=?`,
		lastTradeBoundary, lastTradeBoundary, lastTradeContract.Id,
	); err != nil {
		t.Fatalf("set independent last-trade boundary: %v", err)
	}
	lastTradeContract.LastTradeTime = lastTradeBoundary
	lastTradeOrder := insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P0-LAST-TRADE-PENDING-ORDER",
		UserId: 128, AccountId: 7128, ContractId: lastTradeContract.Id,
		UnderlyingSymbol: "BTCUSDT", Side: int64(common.Side_SIDE_BUY),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType:      int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(10),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1),
		FeeCoin: "USDT", MarginCoin: "USDT",
		Source:     int64(option.OrderSource_ORDER_SOURCE_APP),
		ReduceOnly: int64(common.YesNo_YES_NO_NO), Mmp: int64(common.YesNo_YES_NO_NO),
		Status:      int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: lastTradeBoundary - 1, UpdateTimes: lastTradeBoundary - 1,
	})
	if err := NewProcessContractLifecycleLogic(ctx, serviceCtx).closeContractTrading(
		lastTradeContract.Id, lastTradeContract.TenantId,
		lastTradeContract.LastTradeTime, lastTradeBoundary,
	); err != nil {
		t.Fatalf("close contract at last-trade boundary: %v", err)
	}
	var contractStatus, orderStatus int64
	var cancelReason string
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_contract WHERE id=?`,
		lastTradeContract.Id).Scan(&contractStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT status,cancel_reason FROM t_option_order WHERE id=?`,
		lastTradeOrder.Id).Scan(&orderStatus, &cancelReason); err != nil {
		t.Fatal(err)
	}
	if contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) ||
		orderStatus != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		cancelReason != "CONTRACT_LAST_TRADE_ENDED" {
		t.Fatalf("last-trade close state contract/order/reason=%d/%d/%q",
			contractStatus, orderStatus, cancelReason)
	}
	var prematureExercises, prematureSettlements int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_exercise WHERE contract_id=?),
		(SELECT COUNT(*) FROM t_option_settlement WHERE contract_id=?)`,
		lastTradeContract.Id, lastTradeContract.Id,
	).Scan(&prematureExercises, &prematureSettlements); err != nil {
		t.Fatal(err)
	}
	if prematureExercises != 0 || prematureSettlements != 0 {
		t.Fatalf("last trade started expiry work early: exercises=%d settlements=%d",
			prematureExercises, prematureSettlements)
	}

	// The public expiry instruction also passes its optimistic TRADING check,
	// then observes a committed TRADING -> EXPIRED transition after acquiring
	// the same contract lock. No stale instruction may be appended.
	stateNow := time.Now().Unix()
	stateContract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-INSTRUCTION-STATE-CROSS",
		option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		stateNow-60, stateNow+3600, stateNow+7200, stateNow+7200,
		common.YesNo_YES_NO_YES, 9202, 9202,
	)
	statePosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: stateUserID, AccountId: 7126,
		ContractId: stateContract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), ExerciseableQty: decimal.NewFromInt(1),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: stateNow - 30, UpdateTimes: stateNow - 30,
	})
	stateLock := lockP0ExerciseContract(t, ctx, db, stateContract.Id)
	stateCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(stateUserID),
	))
	stateReq := &option.SetExerciseInstructionReq{
		AccountId: 7126, ContractId: stateContract.Id, PositionId: statePosition.Id,
		ClientInstructionId: "P0-INSTRUCTION-STATE-CROSS",
		InstructionType:     option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
	}
	type instructionCallResult struct {
		resp *option.GetExerciseInstructionResp
		err  error
	}
	stateResult := make(chan instructionCallResult, 1)
	go func() {
		resp, err := applogic.NewSetExerciseInstructionLogic(stateCtx, serviceCtx).
			SetExerciseInstruction(stateReq)
		stateResult <- instructionCallResult{resp: resp, err: err}
	}()
	select {
	case early := <-stateResult:
		t.Fatalf("state-transition instruction did not wait for contract lock: resp=%+v err=%v", early.resp, early.err)
	case <-time.After(150 * time.Millisecond):
	}
	if _, err := stateLock.ExecContext(ctx, `UPDATE t_option_contract SET status=?,update_times=? WHERE id=?`,
		int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED), time.Now().Unix(), stateContract.Id,
	); err != nil {
		t.Fatalf("transition locked contract to EXPIRED: %v", err)
	}
	if err := stateLock.Commit(); err != nil {
		t.Fatalf("commit locked contract transition: %v", err)
	}
	select {
	case got := <-stateResult:
		if got.err == nil || got.resp != nil {
			t.Fatalf("state-crossing instruction was not rejected: resp=%+v err=%v", got.resp, got.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("state-crossing instruction did not finish after lock release")
	}
	assertP0RejectedInstructionMutation(t, ctx, db, stateContract.Id, stateReq.ClientInstructionId)

	// PAUSED preserves the holder's expiry-instruction right when no concurrent
	// lifecycle transition occurs.
	pausedNow := time.Now().Unix()
	pausedContract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-INSTRUCTION-PAUSED-ACCEPT",
		option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
		option.ContractStatus_CONTRACT_STATUS_PAUSED,
		pausedNow-60, pausedNow+3600, pausedNow+7200, pausedNow+7200,
		common.YesNo_YES_NO_YES, 9203, 9203,
	)
	if _, err := db.ExecContext(ctx, `UPDATE t_option_contract
		SET last_trade_time=?,update_times=? WHERE id=?`,
		pausedNow-1, pausedNow, pausedContract.Id,
	); err != nil {
		t.Fatalf("set paused contract last-trade boundary: %v", err)
	}
	pausedContract.LastTradeTime = pausedNow - 1
	pausedPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: pausedUserID, AccountId: 7127,
		ContractId: pausedContract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), ExerciseableQty: decimal.NewFromInt(1),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: pausedNow - 30, UpdateTimes: pausedNow - 30,
	})
	pausedCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(pausedUserID),
	))
	pausedResp, err := applogic.NewSetExerciseInstructionLogic(pausedCtx, serviceCtx).
		SetExerciseInstruction(&option.SetExerciseInstructionReq{
			AccountId: 7127, ContractId: pausedContract.Id, PositionId: pausedPosition.Id,
			ClientInstructionId: "P0-INSTRUCTION-PAUSED-ACCEPT",
			InstructionType:     option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
		})
	if err != nil || pausedResp == nil || pausedResp.GetBase().GetCode() != 200 || pausedResp.Data == nil {
		t.Fatalf("PAUSED expiry instruction was rejected: resp=%+v err=%v", pausedResp, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_position SET status=?,update_times=? WHERE id IN (?,?,?)`,
		int64(option.PositionStatus_POSITION_STATUS_EXPIRED), time.Now().Unix(),
		cutoffPosition.Id, statePosition.Id, pausedPosition.Id,
	); err != nil {
		t.Fatalf("retire exercise race evidence positions: %v", err)
	}

	t.Logf("exercise_cutoff_boundary=start:%d cutoff:%d release:%d last_trade_status=%d last_trade_order=%d premature_expiry=0 paused_post_trade_instruction=1",
		startedAtMillis, cutoffMillis, releasedAtMillis, contractStatus, orderStatus)
}

func lockP0ExerciseContract(t *testing.T, ctx context.Context, db *sql.DB, contractID int64) *sql.Tx {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var lockedID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM t_option_contract WHERE id=? FOR UPDATE`, contractID).
		Scan(&lockedID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("lock exercise contract %d: %v", contractID, err)
	}
	if lockedID != contractID {
		_ = tx.Rollback()
		t.Fatalf("locked exercise contract=%d want=%d", lockedID, contractID)
	}
	return tx
}

func assertP0ExerciseCallBlocked(
	t *testing.T,
	results <-chan p0ExerciseCallResult,
	wait time.Duration,
) {
	t.Helper()
	select {
	case early := <-results:
		t.Fatalf("cutoff exercise did not wait for contract lock: resp=%+v err=%v", early.resp, early.err)
	case <-time.After(wait):
	}
}

func waitP0ExerciseResult(
	t *testing.T,
	results <-chan p0ExerciseCallResult,
) p0ExerciseCallResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("cutoff exercise did not finish after lock release")
		return p0ExerciseCallResult{}
	}
}

func assertP0RejectedExerciseMutation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, positionID int64,
	clientID string,
) {
	t.Helper()
	var rows int64
	var available, frozen string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_exercise
		WHERE tenant_id=? AND contract_id=? AND client_exercise_id=?`,
		p0AssetE2ETenantID, contractID, clientID,
	).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT CAST(available_qty AS CHAR),CAST(frozen_qty AS CHAR)
		FROM t_option_position WHERE id=?`, positionID).Scan(&available, &frozen); err != nil {
		t.Fatal(err)
	}
	if rows != 0 || available != "1.0000000000000000" || frozen != "0.0000000000000000" {
		t.Fatalf("rejected cutoff mutation rows/available/frozen=%d/%s/%s", rows, available, frozen)
	}
}

func assertP0RejectedInstructionMutation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	clientID string,
) {
	t.Helper()
	var rows, status int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_exercise_instruction WHERE tenant_id=? AND contract_id=? AND client_instruction_id=?),
		(SELECT status FROM t_option_contract WHERE id=?)`,
		p0AssetE2ETenantID, contractID, clientID, contractID,
	).Scan(&rows, &status); err != nil {
		t.Fatal(err)
	}
	if rows != 0 || status != int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED) {
		t.Fatalf("rejected state mutation rows/status=%d/%d", rows, status)
	}
}

func testP0ExerciseInstructionReplacementRace(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		userID      int64 = 128
		accountID   int64 = 7128
		concurrency       = 20
	)
	now := time.Now().Unix()
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-INSTRUCTION-REPLACE-RACE",
		option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		now-60, now+3600, now+7200, now+7200,
		common.YesNo_YES_NO_YES, 9204, 9204,
	)
	position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: accountID,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), ExerciseableQty: decimal.NewFromInt(1),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 30, UpdateTimes: now - 30,
	})
	instructionCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(userID),
	))
	types := []option.ExerciseInstructionType{
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE,
	}
	type requestEvidence struct {
		clientID        string
		instructionType option.ExerciseInstructionType
		id              int64
		version         int64
	}
	type callResult struct {
		index int
		resp  *option.GetExerciseInstructionResp
		err   error
	}
	requests := make([]requestEvidence, concurrency)
	results := make(chan callResult, concurrency)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		requests[i] = requestEvidence{
			clientID:        fmt.Sprintf("P0-INSTRUCTION-REPLACE-%02d", i+1),
			instructionType: types[i%len(types)],
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			resp, err := applogic.NewSetExerciseInstructionLogic(instructionCtx, serviceCtx).
				SetExerciseInstruction(&option.SetExerciseInstructionReq{
					AccountId: accountID, ContractId: contract.Id, PositionId: position.Id,
					ClientInstructionId: requests[index].clientID,
					InstructionType:     requests[index].instructionType,
				})
			results <- callResult{index: index, resp: resp, err: err}
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	for result := range results {
		if result.err != nil || result.resp == nil || result.resp.GetBase().GetCode() != 200 || result.resp.Data == nil {
			t.Fatalf("different-key replacement request %d failed: resp=%+v err=%v",
				result.index, result.resp, result.err)
		}
		item := result.resp.Data
		if item.ClientInstructionId != requests[result.index].clientID ||
			item.InstructionType != requests[result.index].instructionType ||
			item.Status != option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE {
			t.Fatalf("different-key replacement response %d changed: %+v", result.index, item)
		}
		requests[result.index].id = item.Id
		requests[result.index].version = item.Version
	}
	assertP0ExerciseInstructionVersionChain(t, ctx, db, contract.Id, position.Id, concurrency)

	// Every original key remains an immutable idempotency identity. Replaying an
	// old, now-superseded key returns that exact historical version and never
	// reactivates it or appends a new version.
	for i := range requests {
		resp, err := applogic.NewSetExerciseInstructionLogic(instructionCtx, serviceCtx).
			SetExerciseInstruction(&option.SetExerciseInstructionReq{
				AccountId: accountID, ContractId: contract.Id, PositionId: position.Id,
				ClientInstructionId: requests[i].clientID,
				InstructionType:     requests[i].instructionType,
			})
		if err != nil || resp == nil || resp.GetBase().GetCode() != 200 || resp.Data == nil ||
			resp.Data.Id != requests[i].id || resp.Data.Version != requests[i].version {
			t.Fatalf("replacement replay %s changed identity: resp=%+v err=%v", requests[i].clientID, resp, err)
		}
		expectedStatus := option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_SUPERSEDED
		if requests[i].version == concurrency {
			expectedStatus = option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE
		}
		if resp.Data.Status != expectedStatus {
			t.Fatalf("replacement replay %s status=%s want=%s",
				requests[i].clientID, resp.Data.Status, expectedStatus)
		}
	}

	changedType := option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO
	if requests[0].instructionType == changedType {
		changedType = option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE
	}
	changedResp, err := applogic.NewSetExerciseInstructionLogic(instructionCtx, serviceCtx).
		SetExerciseInstruction(&option.SetExerciseInstructionReq{
			AccountId: accountID, ContractId: contract.Id, PositionId: position.Id,
			ClientInstructionId: requests[0].clientID, InstructionType: changedType,
		})
	if err != nil || changedResp == nil || changedResp.GetBase().GetCode() == 200 {
		t.Fatalf("replacement key accepted changed economics: resp=%+v err=%v", changedResp, err)
	}
	assertP0ExerciseInstructionVersionChain(t, ctx, db, contract.Id, position.Id, concurrency)

	if _, err := db.ExecContext(ctx, `UPDATE t_option_position SET status=?,update_times=? WHERE id=?`,
		int64(option.PositionStatus_POSITION_STATUS_EXPIRED), time.Now().Unix(), position.Id,
	); err != nil {
		t.Fatalf("retire replacement race position: %v", err)
	}
	t.Logf("exercise_instruction_replacement_race=requests=%d versions=%d active=1 superseded=%d replays=%d",
		concurrency, concurrency, concurrency-1, concurrency)
}

func assertP0ExerciseInstructionVersionChain(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, positionID int64,
	want int,
) {
	t.Helper()
	rows, err := db.QueryContext(ctx, `SELECT id,client_instruction_id,version,status,supersedes_id,cutoff_time
		FROM t_option_exercise_instruction
		WHERE tenant_id=? AND contract_id=? AND position_id=? ORDER BY version`,
		p0AssetE2ETenantID, contractID, positionID,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	var previousID int64
	clients := make(map[string]struct{}, want)
	for rows.Next() {
		var id, version, status, supersedesID, cutoffTime int64
		var clientID string
		if err := rows.Scan(&id, &clientID, &version, &status, &supersedesID, &cutoffTime); err != nil {
			t.Fatal(err)
		}
		count++
		if version != int64(count) || cutoffTime <= 0 {
			t.Fatalf("replacement version/cutoff at row %d = %d/%d", count, version, cutoffTime)
		}
		if count == 1 && supersedesID != 0 || count > 1 && supersedesID != previousID {
			t.Fatalf("replacement chain row %d supersedes=%d want=%d", count, supersedesID, previousID)
		}
		expectedStatus := int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_SUPERSEDED)
		if count == want {
			expectedStatus = int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE)
		}
		if status != expectedStatus {
			t.Fatalf("replacement chain row %d status=%d want=%d", count, status, expectedStatus)
		}
		if _, exists := clients[clientID]; exists {
			t.Fatalf("replacement chain duplicated client key %q", clientID)
		}
		clients[clientID] = struct{}{}
		previousID = id
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != want || len(clients) != want {
		t.Fatalf("replacement chain rows/clients=%d/%d want=%d", count, len(clients), want)
	}
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
		// Keep manually seeded trade identities in a namespace disjoint from
		// liquidation takeover lots, which use -liquidation_id.
		TradeId: -1_000_000_000 - position.Id, FreezeBizNo: freezeBizNo, CollateralCoin: "USDT",
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
