package models

import (
	"context"
	"testing"

	"wklive/proto/option"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestResetFailedAssetInstructionsByBizNo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const now = int64(1234)
	mock.ExpectExec(`(?s)UPDATE t_option_asset_instruction.*WHERE tenant_id = \? AND biz_no = \? AND status IN \(\?, \?\)`).
		WithArgs(
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			now,
			int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			now,
			int64(9),
			"EX-1",
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW),
		).
		WillReturnResult(sqlmock.NewResult(0, 2))

	model := &customTOptionAssetInstructionModel{
		defaultTOptionAssetInstructionModel: &defaultTOptionAssetInstructionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_asset_instruction`",
		},
	}
	affected, err := model.ResetFailedByBizNo(context.Background(), 9, "EX-1", now)
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	if affected != 2 {
		t.Fatalf("affected=%d want=2", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindRunnableAssetInstructionsEnforcesBizStepBarrier(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .* AS current.*NOT EXISTS \(.*biz_previous\.biz_no = current\.biz_no.*biz_previous\.step_no < current\.step_no.*biz_previous\.status <> \?.*\).*ORDER BY id ASC LIMIT \?`).
		WithArgs(
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
			int64(1234),
			int64(0),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(9),
			int64(100),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionAssetInstructionModel{
		defaultTOptionAssetInstructionModel: &defaultTOptionAssetInstructionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_asset_instruction`",
		},
	}
	items, err := model.FindRunnable(context.Background(), 9, 1234, 0, 100)
	if err != nil {
		t.Fatalf("find runnable failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d want=0", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestResetExerciseAssignmentsForRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const now = int64(2345)
	mock.ExpectExec(`(?s)UPDATE t_option_exercise_assignment.*WHERE tenant_id = \? AND exercise_id = \? AND status IN \(\?, \?\)`).
		WithArgs(
			int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING),
			now,
			int64(9),
			int64(88),
			int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_FAILED),
			int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_MANUAL_REVIEW),
		).
		WillReturnResult(sqlmock.NewResult(0, 2))

	model := &customTOptionExerciseAssignmentModel{
		defaultTOptionExerciseAssignmentModel: &defaultTOptionExerciseAssignmentModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_exercise_assignment`",
		},
	}
	if err := model.ResetForRetry(context.Background(), 9, 88, now); err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLiquidationClaimUsesCompareAndSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const now = int64(3456)
	claimPattern := `(?s)UPDATE t_option_liquidation.*WHERE id = \? AND status IN \(\?, \?\)`
	mock.ExpectExec(claimPattern).WithArgs(
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
		now,
		int64(77),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(claimPattern).WithArgs(
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_EXECUTING),
		now,
		int64(77),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		int64(option.LiquidationStatus_LIQUIDATION_STATUS_FAILED),
	).WillReturnResult(sqlmock.NewResult(0, 0))

	model := &customTOptionLiquidationModel{
		defaultTOptionLiquidationModel: &defaultTOptionLiquidationModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_liquidation`",
		},
	}
	first, err := model.Claim(context.Background(), 77, now)
	if err != nil || !first {
		t.Fatalf("first claim=(%v,%v), want true,nil", first, err)
	}
	second, err := model.Claim(context.Background(), 77, now)
	if err != nil || second {
		t.Fatalf("second claim=(%v,%v), want false,nil", second, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
