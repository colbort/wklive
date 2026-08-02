package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// TestP0ExerciseProcessKillTakeover proves the American-exercise process
// boundary independently of ordinary order funding and physical delivery:
// the short debit/release legs finish, the long credit commits in Asset, and
// the Option worker is SIGKILLed before receiving the response. Two fresh
// processes then compete after the real Redis lease expiry; exactly one must
// reconcile the original credit and finish the fee credit and exercise state.
func TestP0ExerciseProcessKillTakeover(t *testing.T) {
	if os.Getenv(p0MultiInstanceWorkerEnv) == "1" {
		t.Skip("parent-only American exercise process-kill acceptance")
	}
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	directRPCAddr := os.Getenv(p0MultiInstanceRPCEnv)
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || directRPCAddr == "" || redisAddr == "" {
		t.Skip("Option P0 Asset E2E environment is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 85*time.Second)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping acceptance database: %v", err)
	}
	directConn, err := grpc.NewClient(
		directRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("connect direct Asset RPC: %v", err)
	}
	defer directConn.Close()
	directAsset := asset.NewAssetClient(directConn)
	waitForAssetRPC(t, ctx, directAsset)
	serviceCtx := newP0AssetE2EServiceContext(dsn, redisAddr, directAsset)
	assertP0ExerciseTaskLockAbsent(t, ctx, serviceCtx.Redis)

	const (
		longUserID   int64 = 93151
		shortUserID  int64 = 93152
		feeUserID    int64 = 93153
		longAccount  int64 = 97151
		shortAccount int64 = 98152
	)
	now := time.Now().Unix()
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-AMERICAN-EXERCISE-PROCESS-KILL",
		option.ExerciseStyle_EXERCISE_STYLE_AMERICAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		now-3600, now+3600, now+7200, now+7200,
		common.YesNo_YES_NO_NO, feeUserID, 9153,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "140", "40", now)
	creditAsset(t, ctx, directAsset, longUserID, "100", "P0-EXERCISE-KILL-LONG-SEED")
	creditAsset(t, ctx, directAsset, shortUserID, "100", "P0-EXERCISE-KILL-SHORT-SEED")
	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: longAccount,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), ExerciseableQty: decimal.NewFromInt(1),
		OpenAvgPrice: decimal.NewFromInt(10), MarkPrice: decimal.NewFromInt(40),
		PositionValue: decimal.NewFromInt(40),
		Status:        int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes:   now - 300, UpdateTimes: now - 300,
	})
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUserID, AccountId: shortAccount,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		MarkPrice: decimal.NewFromInt(40), PositionValue: decimal.NewFromInt(40),
		MarginAmount: decimal.NewFromInt(50), MaintenanceMargin: decimal.NewFromInt(20),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 200, UpdateTimes: now - 200,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-EXERCISE-KILL-SHORT-MARGIN", "1", "50", now-190,
	)
	freezeP0ExerciseMargin(t, ctx, directAsset, shortPosition, lot, "50")

	exerciseCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(longUserID),
	))
	resp, err := applogic.NewExerciseLogic(exerciseCtx, serviceCtx).Exercise(&option.ExerciseReq{
		AccountId: longAccount, ContractId: contract.Id, PositionId: longPosition.Id,
		ExerciseQty: "1", ClientExerciseId: "P0-AMERICAN-EXERCISE-PROCESS-KILL",
	})
	if err != nil || resp == nil || resp.GetBase().GetCode() != 200 || resp.Data == nil {
		t.Fatalf("create process-kill exercise: resp=%+v err=%v", resp, err)
	}
	exercise, err := serviceCtx.OptionExerciseModel.FindOne(ctx, resp.Data.ExerciseId)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise); err != nil {
		t.Fatalf("create process-kill exercise clearing: %v", err)
	}
	instructions, err := serviceCtx.OptionAssetInstructionModel.FindByBizNo(
		ctx, p0AssetE2ETenantID, exercise.ExerciseNo,
	)
	if err != nil {
		t.Fatal(err)
	}
	var targetCredit, feeCredit *models.TOptionAssetInstruction
	for _, instruction := range instructions {
		if instruction.Action != int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE) {
			continue
		}
		if instruction.UserId == longUserID {
			targetCredit = instruction
		} else if instruction.UserId == feeUserID {
			feeCredit = instruction
		}
	}
	if len(instructions) != 4 || targetCredit == nil || feeCredit == nil {
		t.Fatalf("exercise process-kill instructions=%d target=%+v fee=%+v",
			len(instructions), targetCredit, feeCredit)
	}

	// The first worker sees only step 1 as runnable and completes the short
	// frozen debit and release. Step 2 remains pending until a later task pass.
	firstStep := runP0AssetWorker(t, directRPCAddr)
	if firstStep.err != nil || firstStep.code != 200 {
		t.Fatalf("exercise step-1 worker code=%d err=%v output=%s",
			firstStep.code, firstStep.err, firstStep.output)
	}
	assertP0ExerciseStepOneBoundary(t, ctx, db, exercise.Id, exercise.ExerciseNo)

	victimProxy := newP1PhysicalCreditCommitProxy(t, directAsset, targetCredit.InstructionNo)
	defer victimProxy.stop()
	victim := startP0AssetWorker(t, victimProxy.address())
	victimResults := make(chan p0WorkerResult, 1)
	go func() { victimResults <- waitP0AssetWorker(victim) }()
	select {
	case <-victimProxy.committed:
	case exited := <-victimResults:
		t.Fatalf("exercise victim exited before committed credit, code=%d err=%v output=%s",
			exited.code, exited.err, exited.output)
	case <-ctx.Done():
		t.Fatalf("exercise victim did not reach committed credit: %v output=%s", ctx.Err(), victim.output.String())
	}
	assertP0ExerciseCommittedCreditBeforeKill(
		t, ctx, db, exercise.Id, exercise.ExerciseNo, targetCredit.Id, feeCredit.Id,
	)
	if err := victim.cmd.Process.Kill(); err != nil {
		t.Fatalf("SIGKILL exercise victim worker: %v", err)
	}
	if killed := <-victimResults; killed.err == nil {
		t.Fatal("SIGKILLed exercise victim worker exited successfully")
	}
	victimProxy.stop()
	if victimProxy.calls.Load() != 1 {
		t.Fatalf("exercise victim credit calls=%d want=1", victimProxy.calls.Load())
	}

	blocked := runP0AssetWorker(t, directRPCAddr)
	if blocked.code == 200 {
		t.Fatalf("fresh exercise worker bypassed killed lease: %s", blocked.output)
	}
	startedWaiting := time.Now()
	waitP0TaskLeaseExpiry(t, ctx, serviceCtx.Redis)
	t.Logf("exercise killed worker lease expired naturally after %s",
		time.Since(startedWaiting).Round(time.Millisecond))

	result, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET update_times=? WHERE id=? AND status=?`,
		time.Now().Unix()-61, targetCredit.Id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING),
	)
	if err != nil {
		t.Fatalf("age killed exercise instruction for takeover: %v", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		t.Fatalf("age killed exercise instruction rows=%d err=%v", affected, rowsErr)
	}

	takeoverProxy := newP1PhysicalCreditCommitProxy(t, directAsset, targetCredit.InstructionNo)
	defer takeoverProxy.stop()
	first := startP0AssetWorker(t, takeoverProxy.address())
	second := startP0AssetWorker(t, takeoverProxy.address())
	results := make(chan p0WorkerResult, 2)
	go func() { results <- waitP0AssetWorker(first) }()
	go func() { results <- waitP0AssetWorker(second) }()
	select {
	case <-takeoverProxy.committed:
	case <-ctx.Done():
		t.Fatalf("exercise takeover did not replay committed credit: %v", ctx.Err())
	}
	select {
	case loser := <-results:
		if loser.err != nil || loser.code == 200 {
			t.Fatalf("expected one clean exercise lease rejection, code=%d err=%v output=%s",
				loser.code, loser.err, loser.output)
		}
	case <-ctx.Done():
		t.Fatalf("competing exercise worker did not observe lease: %v", ctx.Err())
	}
	close(takeoverProxy.release)
	winner := <-results
	if winner.err != nil || winner.code != 200 {
		t.Fatalf("exercise takeover worker code=%d err=%v output=%s", winner.code, winner.err, winner.output)
	}
	takeoverProxy.stop()
	if takeoverProxy.calls.Load() != 1 {
		t.Fatalf("competing exercise takeover credit calls=%d want=1", takeoverProxy.calls.Load())
	}

	assertP0ExerciseProcessKillCompleted(
		t, ctx, db, exercise.Id, exercise.ExerciseNo, targetCredit.Id, feeCredit.Id,
	)
	assertWalletAmounts(t, ctx, db, longUserID,
		"136.000000000000000000", "136.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, shortUserID,
		"60.000000000000000000", "60.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID,
		"4.000000000000000000", "4.000000000000000000", "0.000000000000000000")
	assertP0ExerciseThreeWalletTotal(t, ctx, db, longUserID, shortUserID, feeUserID,
		"200.000000000000000000", "200.000000000000000000", "0.000000000000000000")
	assertP0ExercisePosition(t, ctx, db, longPosition.Id,
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		option.PositionStatus_POSITION_STATUS_EXERCISED)
	assertP0ExercisePosition(t, ctx, db, shortPosition.Id,
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		option.PositionStatus_POSITION_STATUS_EXERCISED)
	assertP0ExerciseLot(t, ctx, db, lot.Id,
		"0.0000000000000000", "0.0000000000000000", "0.0000000000000000",
		option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED)
	assertP0ExerciseReturn(t, ctx, db, longPosition.Id,
		"30.0000000000000000", "4.0000000000000000", "26.0000000000000000")
	assertP0ExerciseReturn(t, ctx, db, shortPosition.Id,
		"-30.0000000000000000", "0.0000000000000000", "-30.0000000000000000")
	assertP0ExerciseTaskLockAbsent(t, ctx, serviceCtx.Redis)
}

func assertP0ExerciseThreeWalletTotal(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	longUserID, shortUserID, feeUserID int64,
	total, available, frozen string,
) {
	t.Helper()
	var gotTotal, gotAvailable, gotFrozen string
	if err := db.QueryRowContext(ctx, `SELECT CAST(SUM(total_amount) AS CHAR),
		CAST(SUM(available_amount) AS CHAR),CAST(SUM(frozen_amount) AS CHAR)
		FROM t_user_asset
		WHERE tenant_id=? AND wallet_type=? AND coin='USDT' AND user_id IN (?,?,?)`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION),
		longUserID, shortUserID, feeUserID,
	).Scan(&gotTotal, &gotAvailable, &gotFrozen); err != nil {
		t.Fatal(err)
	}
	if gotTotal != total || gotAvailable != available || gotFrozen != frozen {
		t.Fatalf("exercise wallet conservation=%s/%s/%s want=%s/%s/%s",
			gotTotal, gotAvailable, gotFrozen, total, available, frozen)
	}
}

func assertP0ExerciseStepOneBoundary(
	t *testing.T, ctx context.Context, db *sql.DB, exerciseID int64, exerciseNo string,
) {
	t.Helper()
	var exerciseStatus, assignmentStatus, stepOneSuccess, stepOneFlows, stepTwoSuccess, stepTwoFlows int64
	if err := db.QueryRowContext(ctx, `SELECT exercise.status,assignment.status,
		(SELECT COUNT(*) FROM t_option_asset_instruction instruction
		 WHERE instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
		   AND instruction.step_no=1 AND instruction.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow JOIN t_option_asset_instruction instruction
		   ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		 WHERE instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
		   AND instruction.step_no=1),
		(SELECT COUNT(*) FROM t_option_asset_instruction instruction
		 WHERE instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
		   AND instruction.step_no=2 AND instruction.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow JOIN t_option_asset_instruction instruction
		   ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		 WHERE instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
		   AND instruction.step_no=2)
		FROM t_option_exercise exercise
		JOIN t_option_exercise_assignment assignment
		  ON assignment.tenant_id=exercise.tenant_id AND assignment.exercise_id=exercise.id
		WHERE exercise.id=? AND exercise.exercise_no=?`, exerciseID, exerciseNo,
	).Scan(
		&exerciseStatus, &assignmentStatus, &stepOneSuccess, &stepOneFlows, &stepTwoSuccess, &stepTwoFlows,
	); err != nil {
		t.Fatal(err)
	}
	if exerciseStatus != int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING) ||
		assignmentStatus != int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING) ||
		stepOneSuccess != 2 || stepOneFlows != 2 || stepTwoSuccess != 0 || stepTwoFlows != 0 {
		t.Fatalf("exercise step-1 boundary status=%d/%d step1=%d/%d step2=%d/%d",
			exerciseStatus, assignmentStatus, stepOneSuccess, stepOneFlows, stepTwoSuccess, stepTwoFlows)
	}
}

func assertP0ExerciseCommittedCreditBeforeKill(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
	targetCreditID, feeCreditID int64,
) {
	t.Helper()
	var exerciseStatus, assignmentStatus, stepOneSuccess, stepOneFlows int64
	var targetStatus, targetRetry, targetFlows, feeStatus, feeFlows int64
	if err := db.QueryRowContext(ctx, `SELECT exercise.status,assignment.status,
		(SELECT COUNT(*) FROM t_option_asset_instruction instruction
		 WHERE instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
		   AND instruction.step_no=1 AND instruction.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow JOIN t_option_asset_instruction instruction
		   ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
		 WHERE instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
		   AND instruction.step_no=1),
		target.status,target.retry_count,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=target.tenant_id AND flow.biz_no=target.instruction_no),
		fee.status,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=fee.tenant_id AND flow.biz_no=fee.instruction_no)
		FROM t_option_exercise exercise
		JOIN t_option_exercise_assignment assignment
		  ON assignment.tenant_id=exercise.tenant_id AND assignment.exercise_id=exercise.id
		JOIN t_option_asset_instruction target ON target.id=? AND target.biz_no=exercise.exercise_no
		JOIN t_option_asset_instruction fee ON fee.id=? AND fee.biz_no=exercise.exercise_no
		WHERE exercise.id=? AND exercise.exercise_no=?`,
		targetCreditID, feeCreditID, exerciseID, exerciseNo,
	).Scan(
		&exerciseStatus, &assignmentStatus, &stepOneSuccess, &stepOneFlows,
		&targetStatus, &targetRetry, &targetFlows, &feeStatus, &feeFlows,
	); err != nil {
		t.Fatal(err)
	}
	if exerciseStatus != int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING) ||
		assignmentStatus != int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING) ||
		stepOneSuccess != 2 || stepOneFlows != 2 ||
		targetStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING) ||
		targetRetry != 0 || targetFlows != 1 ||
		feeStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) || feeFlows != 0 {
		t.Fatalf("exercise before kill status=%d/%d step1=%d/%d target=%d/%d/%d fee=%d/%d",
			exerciseStatus, assignmentStatus, stepOneSuccess, stepOneFlows,
			targetStatus, targetRetry, targetFlows, feeStatus, feeFlows)
	}
}

func assertP0ExerciseProcessKillCompleted(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
	targetCreditID, feeCreditID int64,
) {
	t.Helper()
	var exerciseStatus, assignments, assignmentDone, instructions, success, reconciled, flows int64
	var targetRetry, targetFlows, feeFlows, duplicateFlows int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_exercise WHERE id=?`, exerciseID).
		Scan(&exerciseStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=2)
		FROM t_option_exercise_assignment WHERE tenant_id=? AND exercise_id=?`,
		p0AssetE2ETenantID, exerciseID,
	).Scan(&assignments, &assignmentDone); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(instruction.status=3),
		SUM(instruction.reconciliation_status=2),COUNT(DISTINCT flow.id),
		COUNT(flow.id)-COUNT(DISTINCT flow.flow_no)
		FROM t_option_asset_instruction instruction
		LEFT JOIN t_asset_flow flow
		  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
		WHERE instruction.tenant_id=? AND instruction.biz_no=?`,
		p0AssetE2ETenantID, exerciseNo,
	).Scan(&instructions, &success, &reconciled, &flows, &duplicateFlows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT target.retry_count,
		(SELECT COUNT(*) FROM t_asset_flow WHERE tenant_id=target.tenant_id AND biz_no=target.instruction_no),
		(SELECT COUNT(*) FROM t_asset_flow WHERE tenant_id=fee.tenant_id AND biz_no=fee.instruction_no)
		FROM t_option_asset_instruction target JOIN t_option_asset_instruction fee ON fee.id=?
		WHERE target.id=?`, feeCreditID, targetCreditID,
	).Scan(&targetRetry, &targetFlows, &feeFlows); err != nil {
		t.Fatal(err)
	}
	if exerciseStatus != int64(option.ExerciseStatus_EXERCISE_STATUS_DONE) ||
		assignments != 1 || assignmentDone != 1 || instructions != 4 || success != 4 || reconciled != 4 ||
		flows != 4 || duplicateFlows != 0 || targetRetry != 0 || targetFlows != 1 || feeFlows != 1 {
		t.Fatalf("exercise takeover status=%d assignments=%d/%d instructions=%d/%d/%d flows=%d duplicate=%d target=%d/%d fee=%d",
			exerciseStatus, assignmentDone, assignments, success, reconciled, instructions,
			flows, duplicateFlows, targetRetry, targetFlows, feeFlows)
	}
}

func assertP0ExerciseTaskLockAbsent(
	t *testing.T,
	ctx context.Context,
	redis interface {
		ExistsCtx(context.Context, string) (bool, error)
	},
) {
	t.Helper()
	locked, err := redis.ExistsCtx(ctx, p0MultiInstanceLockKey)
	if err != nil {
		t.Fatal(err)
	}
	if locked {
		t.Fatalf("task lock %s remained after exercise process takeover", p0MultiInstanceLockKey)
	}
}
