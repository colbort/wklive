package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	gosqlx "github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestP0AmericanAssignmentCapacityAssetRPC(t *testing.T) {
	gosqlx.DisableLog()
	rawCount := os.Getenv("OPTION_P0_ASSET_CAPACITY_SHORTS")
	if rawCount == "" {
		t.Skip("OPTION_P0_ASSET_CAPACITY_SHORTS is required")
	}
	shortCount, err := strconv.Atoi(rawCount)
	if err != nil || (shortCount != 501 && shortCount != 5000) {
		t.Fatalf("OPTION_P0_ASSET_CAPACITY_SHORTS=%q must be 501 or 5000", rawCount)
	}
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	rpcAddr := os.Getenv("OPTION_P0_ASSET_E2E_RPC_ADDR")
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || rpcAddr == "" || redisAddr == "" {
		t.Fatal("OPTION_P0_ASSET_E2E_DSN, OPTION_P0_ASSET_E2E_RPC_ADDR and OPTION_P0_ASSET_E2E_REDIS_ADDR are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping acceptance database: %v", err)
	}
	grpcConn, err := grpc.NewClient(rpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("connect Asset RPC: %v", err)
	}
	defer grpcConn.Close()
	assetClient := asset.NewAssetClient(grpcConn)
	waitForAssetRPC(t, ctx, assetClient)
	serviceCtx := newP0AssetE2EServiceContext(dsn, redisAddr, assetClient)

	runP0AmericanAssignmentCapacity(t, ctx, db, assetClient, serviceCtx, shortCount)
}

func runP0AmericanAssignmentCapacity(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	shortCount int,
) {
	t.Helper()
	now := time.Now().Unix()
	longUserID := int64(300000 + shortCount)
	feeUserID := int64(320000 + shortCount)
	shortUserBase := int64(400000 + shortCount*10)
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, fmt.Sprintf("P0-AMERICAN-CAPACITY-%d", shortCount),
		option.ExerciseStyle_EXERCISE_STYLE_AMERICAN,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		now-3600, now+3600, now+7200, now+7200,
		common.YesNo_YES_NO_NO, feeUserID, feeUserID,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "101", "1", now)
	creditAsset(
		t, ctx, assetClient, longUserID, "1",
		fmt.Sprintf("P0-AMERICAN-CAPACITY-%d-LONG-SEED", shortCount),
	)
	quantity := decimal.NewFromInt(int64(shortCount))
	longPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: longUserID, AccountId: longUserID,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: quantity,
		AvailableQty: quantity, OpenAvgPrice: decimal.NewFromInt(1),
		MarkPrice: decimal.NewFromInt(1), PositionValue: quantity,
		ExerciseableQty: quantity, Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 20000, UpdateTimes: now - 20000,
	})

	seedStarted := time.Now()
	firstShortID, lastShortID := seedP0AmericanCapacityShorts(
		t, ctx, db, contract.Id, shortUserBase, shortCount, now,
	)
	seedElapsed := time.Since(seedStarted)

	exerciseCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyTenantId, fmt.Sprint(p0AssetE2ETenantID),
		utils.CtxKeyUid, fmt.Sprint(longUserID),
	))
	clientExerciseID := fmt.Sprintf("P0-AMERICAN-CAPACITY-%d", shortCount)
	response, err := applogic.NewExerciseLogic(exerciseCtx, serviceCtx).Exercise(&option.ExerciseReq{
		AccountId: longUserID, ContractId: contract.Id, PositionId: longPosition.Id,
		ExerciseQty: quantity.String(), ClientExerciseId: clientExerciseID,
	})
	if err != nil {
		t.Fatalf("submit capacity exercise: %v", err)
	}
	if response == nil || response.GetBase().GetCode() != 200 || response.Data == nil {
		t.Fatalf("capacity exercise rejected: %+v", response)
	}
	exercise, err := serviceCtx.OptionExerciseModel.FindOne(ctx, response.Data.ExerciseId)
	if err != nil {
		t.Fatal(err)
	}

	clearingStarted := time.Now()
	if err := NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise); err != nil {
		t.Fatalf("create %d-short clearing: %v", shortCount, err)
	}
	clearingElapsed := time.Since(clearingStarted)
	assertP0AmericanCapacityClearing(
		t, ctx, db, exercise.Id, exercise.ExerciseNo, shortCount, firstShortID, lastShortID,
	)

	assetStarted := time.Now()
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assetElapsed := time.Since(assetStarted)
	assertP0AmericanCapacityCompleted(
		t, ctx, db, exercise.Id, exercise.ExerciseNo, contract.Id,
		longUserID, feeUserID, shortUserBase, shortCount,
	)

	if err := NewProcessExercisesLogic(ctx, serviceCtx).createExerciseClearing(exercise); err != nil {
		t.Fatalf("replay %d-short clearing: %v", shortCount, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0AmericanCapacityCompleted(
		t, ctx, db, exercise.Id, exercise.ExerciseNo, contract.Id,
		longUserID, feeUserID, shortUserBase, shortCount,
	)
	t.Logf(
		"american_assignment_capacity_shorts=%d seed=%s clearing=%s asset_rpc=%s instructions=%d",
		shortCount, seedElapsed.Round(time.Millisecond), clearingElapsed.Round(time.Millisecond),
		assetElapsed.Round(time.Millisecond), shortCount+2,
	)
}

func seedP0AmericanCapacityShorts(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, shortUserBase int64,
	shortCount int,
	now int64,
) (int64, int64) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	positionStmt, err := tx.PrepareContext(ctx, `INSERT INTO t_option_position (
		tenant_id,user_id,account_id,contract_id,underlying_symbol,side,
		position_qty,available_qty,open_avg_price,mark_price,position_value,
		margin_amount,maintenance_margin,status,create_times,update_times
	) VALUES (?,?,?,?,?,2,1,1,1,1,1,1,0.2,1,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer positionStmt.Close()
	lotStmt, err := tx.PrepareContext(ctx, `INSERT INTO t_option_margin_lot (
		tenant_id,user_id,account_id,contract_id,position_id,origin_contract_id,
		origin_position_id,trade_id,freeze_biz_no,collateral_coin,quantity,
		remaining_quantity,initial_margin,remaining_margin,status,create_times,update_times
	) VALUES (?,?,?,?,?,?,?,? ,?,'USDT',1,1,1,1,1,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer lotStmt.Close()
	walletStmt, err := tx.PrepareContext(ctx, `INSERT INTO t_user_asset (
		tenant_id,user_id,wallet_type,coin,total_amount,available_amount,frozen_amount,
		locked_amount,enabled,version,remark,create_times,update_times
	) VALUES (?,?,5,'USDT',1,0,1,0,1,0,'P0 assignment capacity fixture',?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer walletStmt.Close()
	freezeStmt, err := tx.PrepareContext(ctx, `INSERT INTO t_asset_freeze (
		freeze_no,tenant_id,user_id,wallet_type,coin,biz_type,scene_type,biz_id,biz_no,
		amount,used_amount,unfreeze_amount,remain_amount,status,expire_time,remark,
		create_times,update_times
	) VALUES (?,?,?,5,'USDT','option','place_order',?,?,1,0,0,1,1,0,
		'P0 assignment capacity fixture',?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer freezeStmt.Close()

	var firstPositionID, lastPositionID int64
	for index := 0; index < shortCount; index++ {
		userID := shortUserBase + int64(index)
		createTimes := now - int64(shortCount) + int64(index)
		result, err := positionStmt.ExecContext(
			ctx, p0AssetE2ETenantID, userID, userID, contractID,
			"BTCUSDT", createTimes, createTimes,
		)
		if err != nil {
			t.Fatalf("insert capacity short %d: %v", index, err)
		}
		positionID, err := result.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			firstPositionID = positionID
		}
		lastPositionID = positionID
		freezeBizNo := fmt.Sprintf("P0-ASN-CAP-%d-F-%d", shortCount, index+1)
		lotResult, err := lotStmt.ExecContext(
			ctx, p0AssetE2ETenantID, userID, userID, contractID, positionID,
			contractID, positionID, -positionID, freezeBizNo, createTimes, createTimes,
		)
		if err != nil {
			t.Fatalf("insert capacity margin lot %d: %v", index, err)
		}
		lotID, err := lotResult.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		millis := createTimes * 1000
		if _, err := walletStmt.ExecContext(ctx, p0AssetE2ETenantID, userID, millis, millis); err != nil {
			t.Fatalf("insert capacity wallet %d: %v", index, err)
		}
		freezeNo := fmt.Sprintf("P0-ASN-CAP-%d-N-%d", shortCount, index+1)
		if _, err := freezeStmt.ExecContext(
			ctx, freezeNo, p0AssetE2ETenantID, userID, lotID, freezeBizNo, millis, millis,
		); err != nil {
			t.Fatalf("insert capacity freeze %d: %v", index, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit capacity fixtures: %v", err)
	}
	return firstPositionID, lastPositionID
}

func assertP0AmericanCapacityClearing(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
	shortCount int,
	firstShortID, lastShortID int64,
) {
	t.Helper()
	var assignments, minShort, maxShort, instructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),MIN(short_position_id),MAX(short_position_id)
		FROM t_option_exercise_assignment WHERE tenant_id=? AND exercise_id=?`,
		p0AssetE2ETenantID, exerciseID,
	).Scan(&assignments, &minShort, &maxShort); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND biz_no=?`, p0AssetE2ETenantID, exerciseNo,
	).Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if assignments != int64(shortCount) || minShort != firstShortID || maxShort != lastShortID ||
		instructions != int64(shortCount+2) {
		t.Fatalf(
			"capacity clearing assignments/min/max/instructions=%d/%d/%d/%d want=%d/%d/%d/%d",
			assignments, minShort, maxShort, instructions,
			shortCount, firstShortID, lastShortID, shortCount+2,
		)
	}
}

func assertP0AmericanCapacityCompleted(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	exerciseID int64,
	exerciseNo string,
	contractID, longUserID, feeUserID, shortUserBase int64,
	shortCount int,
) {
	t.Helper()
	var exerciseStatus, assignments, assignmentDone, instructions, success, reconciled, flows int64
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
	wantInstructions := int64(shortCount + 2)
	if exerciseStatus != int64(option.ExerciseStatus_EXERCISE_STATUS_DONE) ||
		assignments != int64(shortCount) || assignmentDone != int64(shortCount) ||
		instructions != wantInstructions || success != wantInstructions ||
		reconciled != wantInstructions || flows != wantInstructions {
		t.Fatalf(
			"capacity completion status=%d assignments=%d/%d instructions=%d/%d/%d flows=%d",
			exerciseStatus, assignmentDone, assignments, success, reconciled, instructions, flows,
		)
	}

	var shortPositions, shortLots, closedFreezes, shortWallets int64
	var shortTotal, shortFrozen, longTotal, feeTotal string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_position
		WHERE tenant_id=? AND contract_id=? AND side=2 AND status=3
		  AND position_qty=0 AND available_qty=0 AND frozen_qty=0 AND margin_amount=0`,
		p0AssetE2ETenantID, contractID,
	).Scan(&shortPositions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_margin_lot
		WHERE tenant_id=? AND contract_id=? AND status=?
		  AND remaining_quantity=0 AND remaining_margin=0 AND pending_margin=0`,
		p0AssetE2ETenantID, contractID,
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED),
	).Scan(&shortLots); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_freeze
		WHERE tenant_id=? AND user_id>=? AND user_id<? AND status=4
		  AND used_amount=1 AND remain_amount=0`,
		p0AssetE2ETenantID, shortUserBase, shortUserBase+int64(shortCount),
	).Scan(&closedFreezes); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),CAST(SUM(total_amount) AS CHAR),CAST(SUM(frozen_amount) AS CHAR)
		FROM t_user_asset WHERE tenant_id=? AND user_id>=? AND user_id<?
		  AND wallet_type=5 AND coin='USDT'`,
		p0AssetE2ETenantID, shortUserBase, shortUserBase+int64(shortCount),
	).Scan(&shortWallets, &shortTotal, &shortFrozen); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT CAST(total_amount AS CHAR) FROM t_user_asset
		WHERE tenant_id=? AND user_id=? AND wallet_type=5 AND coin='USDT'`,
		p0AssetE2ETenantID, longUserID,
	).Scan(&longTotal); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT CAST(total_amount AS CHAR) FROM t_user_asset
		WHERE tenant_id=? AND user_id=? AND wallet_type=5 AND coin='USDT'`,
		p0AssetE2ETenantID, feeUserID,
	).Scan(&feeTotal); err != nil {
		t.Fatal(err)
	}
	wantLong := decimal.NewFromInt(1).Add(decimal.NewFromInt(int64(shortCount)).Mul(decimal.RequireFromString("0.9")))
	wantFee := decimal.NewFromInt(int64(shortCount)).Mul(decimal.RequireFromString("0.1"))
	gotShortTotal := decimal.RequireFromString(shortTotal)
	gotShortFrozen := decimal.RequireFromString(shortFrozen)
	gotLong := decimal.RequireFromString(longTotal)
	gotFee := decimal.RequireFromString(feeTotal)
	if shortPositions != int64(shortCount) || shortLots != int64(shortCount) ||
		closedFreezes != int64(shortCount) || shortWallets != int64(shortCount) ||
		!gotShortTotal.IsZero() || !gotShortFrozen.IsZero() ||
		!gotLong.Equal(wantLong) || !gotFee.Equal(wantFee) ||
		!gotLong.Add(gotFee).Equal(decimal.NewFromInt(int64(shortCount+1))) {
		t.Fatalf(
			"capacity conservation positions/lots/freezes/wallets=%d/%d/%d/%d short=%s/%s long=%s fee=%s",
			shortPositions, shortLots, closedFreezes, shortWallets,
			shortTotal, shortFrozen, longTotal, feeTotal,
		)
	}
}
