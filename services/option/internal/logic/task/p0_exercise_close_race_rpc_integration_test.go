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

func testP0AmericanExerciseCloseOrderRace(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	seedP0OpenTradingCalendar(t, ctx, db, "CONTINUOUS_24_7", now)

	accepted, rejected := runP0AmericanCloseRaceScenario(
		t, ctx, db, assetClient, serviceCtx, 0, true,
	)
	if accepted != 1 || rejected != 0 {
		t.Fatalf("pre-existing close order branch=%d/%d want=1/0", accepted, rejected)
	}

	for iteration := 1; iteration <= 10; iteration++ {
		wasAccepted, wasRejected := runP0AmericanCloseRaceScenario(
			t, ctx, db, assetClient, serviceCtx, iteration, false,
		)
		accepted += wasAccepted
		rejected += wasRejected
	}
	if accepted+rejected != 11 {
		t.Fatalf("close/exercise outcomes accepted=%d rejected=%d want=11", accepted, rejected)
	}
	t.Logf(
		"american_exercise_close_race_iterations=10 accepted_then_canceled=%d rejected_after_assignment=%d",
		accepted-1, rejected,
	)
}

func runP0AmericanCloseRaceScenario(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	iteration int,
	placeBeforeClearing bool,
) (int, int) {
	t.Helper()
	now := time.Now().Unix()
	longUserID := int64(700 + iteration*3)
	shortUserID := longUserID + 1
	feeUserID := longUserID + 2
	accountID := shortUserID
	prefix := fmt.Sprintf("P0-AMERICAN-CLOSE-RACE-%02d", iteration)
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, prefix,
		option.ExerciseStyle_EXERCISE_STYLE_AMERICAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		now-3600, now+3600, now+7200, now+7200,
		common.YesNo_YES_NO_NO, feeUserID, feeUserID,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "101", "1", now)
	creditAsset(t, ctx, assetClient, longUserID, "1", prefix+"-LONG-SEED")
	creditAsset(t, ctx, assetClient, shortUserID, "1", prefix+"-SHORT-SEED")
	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: longUserID,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(1),
		MarkPrice: decimal.NewFromInt(1), PositionValue: decimal.NewFromInt(1),
		ExerciseableQty: decimal.NewFromInt(1), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 300, UpdateTimes: now - 300,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: accountID,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(1),
		MarkPrice: decimal.NewFromInt(1), PositionValue: decimal.NewFromInt(1),
		MarginAmount: decimal.NewFromInt(1), MaintenanceMargin: decimal.RequireFromString("0.2"),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortPosition, prefix+"-SHORT-MARGIN", "1", "1", now-190,
	)
	freezeP0ExerciseMargin(t, ctx, assetClient, shortPosition, lot, "1")

	exerciseCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(longUserID),
	))
	exerciseResponse, err := applogic.NewExerciseLogic(exerciseCtx, serviceCtx).Exercise(&option.ExerciseReq{
		AccountId: longUserID, ContractId: contract.Id, PositionId: longPosition.Id,
		ExerciseQty: "1", ClientExerciseId: prefix,
	})
	if err != nil || exerciseResponse == nil || exerciseResponse.GetBase().GetCode() != 200 ||
		exerciseResponse.Data == nil {
		t.Fatalf("submit close-race exercise iteration=%d resp=%+v err=%v", iteration, exerciseResponse, err)
	}
	exercise, err := serviceCtx.OptionExerciseModel.FindOne(ctx, exerciseResponse.Data.ExerciseId)
	if err != nil {
		t.Fatal(err)
	}
	closeRequest := &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_CLOSE,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "0.8", Qty: "1",
		ClientOrderId: prefix + "-CLOSE", ReduceOnly: common.YesNo_YES_NO_YES,
	}

	type placeResult struct {
		response *option.PlaceOrderResp
		err      error
	}
	var result placeResult
	if placeBeforeClearing {
		result.response, result.err = applogic.NewPlaceOrderLogic(
			p0OrderUserContext(ctx, shortUserID), serviceCtx,
		).PlaceOrder(closeRequest)
		if result.err != nil || result.response == nil || result.response.GetBase().GetCode() != 200 ||
			result.response.Data == nil {
			t.Fatalf("place pre-existing close order resp=%+v err=%v", result.response, result.err)
		}
		if err := NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise); err != nil {
			t.Fatalf("clear exercise with pre-existing close order: %v", err)
		}
	} else {
		start := make(chan struct{})
		placeDone := make(chan placeResult, 1)
		clearingDone := make(chan error, 1)
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			<-start
			response, placeErr := applogic.NewPlaceOrderLogic(
				p0OrderUserContext(ctx, shortUserID), serviceCtx,
			).PlaceOrder(closeRequest)
			placeDone <- placeResult{response: response, err: placeErr}
		}()
		go func() {
			defer waitGroup.Done()
			<-start
			clearingDone <- NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise)
		}()
		close(start)
		waitGroup.Wait()
		result = <-placeDone
		if clearingErr := <-clearingDone; clearingErr != nil {
			t.Fatalf("concurrent clearing iteration=%d: %v", iteration, clearingErr)
		}
		if result.err != nil {
			t.Fatalf("concurrent close placement iteration=%d: %v", iteration, result.err)
		}
	}

	accepted, rejected := 0, 0
	if result.response != nil && result.response.GetBase().GetCode() == 200 && result.response.Data != nil {
		accepted = 1
	} else {
		rejected = 1
	}
	assertP0AmericanCloseRaceReserved(
		t, ctx, db, contract.Id, exercise.Id, shortPosition.Id,
		prefix+"-CLOSE", accepted,
	)
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0AmericanCloseRaceCompleted(
		t, ctx, db, exercise.Id, exercise.ExerciseNo, contract.Id,
		longUserID, shortUserID, feeUserID, lot.Id,
	)
	return accepted, rejected
}

func assertP0AmericanCloseRaceReserved(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, exerciseID, shortPositionID int64,
	clientOrderID string,
	accepted int,
) {
	t.Helper()
	var assignments, activeClose, orderCount, canceledByAssignment int64
	var positionQty, availableQty, frozenQty string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_exercise_assignment
		WHERE tenant_id=? AND exercise_id=? AND short_position_id=? AND quantity=1`,
		p0AssetE2ETenantID, exerciseID, shortPositionID,
	).Scan(&assignments); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_order
		WHERE tenant_id=? AND contract_id=? AND position_effect=2 AND status IN (7,1,2)`,
		p0AssetE2ETenantID, contractID,
	).Scan(&activeClose); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		COALESCE(SUM(status=? AND cancel_reason='AMERICAN_EXERCISE_ASSIGNMENT'),0)
		FROM t_option_order WHERE tenant_id=? AND client_order_id=?`,
		int64(option.OrderStatus_ORDER_STATUS_CANCELED), p0AssetE2ETenantID, clientOrderID,
	).Scan(&orderCount, &canceledByAssignment); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT CAST(position_qty AS CHAR),CAST(available_qty AS CHAR),
		CAST(frozen_qty AS CHAR) FROM t_option_position WHERE id=?`, shortPositionID,
	).Scan(&positionQty, &availableQty, &frozenQty); err != nil {
		t.Fatal(err)
	}
	if assignments != 1 || activeClose != 0 || orderCount != int64(accepted) ||
		canceledByAssignment != int64(accepted) || positionQty != "0.0000000000000000" ||
		availableQty != "0.0000000000000000" || frozenQty != "0.0000000000000000" {
		t.Fatalf(
			"close-race reserved assignments/active/orders/canceled=%d/%d/%d/%d position=%s/%s/%s accepted=%d",
			assignments, activeClose, orderCount, canceledByAssignment,
			positionQty, availableQty, frozenQty, accepted,
		)
	}
}

func assertP0AmericanCloseRaceCompleted(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
	contractID, longUserID, shortUserID, feeUserID, lotID int64,
) {
	t.Helper()
	var exerciseStatus, assignmentDone, instructions, success, reconciled, flows int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_exercise WHERE id=?`, exerciseID).
		Scan(&exerciseStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_exercise_assignment
		WHERE tenant_id=? AND exercise_id=? AND status=2`, p0AssetE2ETenantID, exerciseID,
	).Scan(&assignmentDone); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=3),SUM(reconciliation_status=2)
		FROM t_option_asset_instruction WHERE tenant_id=? AND biz_no=?`,
		p0AssetE2ETenantID, exerciseNo,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow flow
		JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		p0AssetE2ETenantID, exerciseNo,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if exerciseStatus != int64(option.ExerciseStatus_EXERCISE_STATUS_DONE) || assignmentDone != 1 ||
		instructions != 3 || success != 3 || reconciled != 3 || flows != 3 {
		t.Fatalf(
			"close-race completion status=%d assignment=%d instructions=%d/%d/%d flows=%d",
			exerciseStatus, assignmentDone, success, reconciled, instructions, flows,
		)
	}
	assertP0ExercisePosition(
		t, ctx, db, lotPositionID(t, ctx, db, lotID),
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		option.PositionStatus_POSITION_STATUS_EXERCISED,
	)
	assertP0ExerciseLot(
		t, ctx, db, lotID,
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED,
	)
	assertWalletAmounts(t, ctx, db, longUserID,
		"1.900000000000000000", "1.900000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, shortUserID,
		"0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID,
		"0.100000000000000000", "0.100000000000000000", "0.000000000000000000")
	var shortCount int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_position
		WHERE tenant_id=? AND contract_id=? AND side=2`, p0AssetE2ETenantID, contractID).
		Scan(&shortCount); err != nil {
		t.Fatal(err)
	}
	if shortCount != 1 {
		t.Fatalf("close-race short positions=%d want=1", shortCount)
	}
}

func lotPositionID(t *testing.T, ctx context.Context, db *sql.DB, lotID int64) int64 {
	t.Helper()
	var positionID int64
	if err := db.QueryRowContext(ctx, `SELECT position_id FROM t_option_margin_lot WHERE id=?`, lotID).
		Scan(&positionID); err != nil {
		t.Fatal(err)
	}
	return positionID
}
