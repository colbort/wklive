package models

import (
	"context"
	"errors"
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestFindLatestExerciseInstructionLocksNewestVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .* WHERE tenant_id = \? AND position_id = \? ORDER BY version DESC LIMIT 1 FOR UPDATE`).
		WithArgs(int64(9), int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionExerciseInstructionModel{
		defaultTOptionExerciseInstructionModel: &defaultTOptionExerciseInstructionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_exercise_instruction`",
		},
	}
	_, err = model.FindLatestByPositionForUpdate(context.Background(), 9, 88)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find latest instruction error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindAssignableShortsPageUsesFIFOKeysetPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND contract_id = \?.*create_times > \? OR \(create_times = \? AND id > \?\).*ORDER BY create_times ASC, id ASC LIMIT \?$`).
		WithArgs(
			int64(9),
			int64(88),
			int64(common.PositionSide_POSITION_SIDE_SHORT),
			int64(option.PositionStatus_POSITION_STATUS_HOLDING),
			int64(100),
			int64(100),
			int64(7),
			int64(500),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionPositionModel{
		defaultTOptionPositionModel: &defaultTOptionPositionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_position`",
		},
	}
	items, err := model.FindAssignableShortsPage(context.Background(), 9, 88, 100, 7, 1000)
	if err != nil {
		t.Fatalf("find assignable shorts failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d want=0", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindPositionByTradingScopeForUpdateLocksRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND user_id = \? AND account_id = \? AND contract_id = \? AND side = \?.*LIMIT 1 FOR UPDATE$`).
		WithArgs(int64(9), int64(77), int64(66), int64(88), int64(common.PositionSide_POSITION_SIDE_SHORT)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionPositionModel{
		defaultTOptionPositionModel: &defaultTOptionPositionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_position`",
		},
	}
	_, err = model.FindOneByTenantIdUserIdAccountIdContractIdSideForUpdate(
		context.Background(), 9, 77, 66, 88, int64(common.PositionSide_POSITION_SIDE_SHORT),
	)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find scoped position error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

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

func TestAllAssetInstructionsSucceededByBizNoUsesAggregate(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		complete bool
	}{
		{name: "all succeeded", complete: true},
		{name: "pending remains", complete: false},
		{name: "no instructions", complete: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			mock.ExpectQuery(`(?s)SELECT.*COUNT\(1\) > 0.*SUM\(status <> \?\).*WHERE tenant_id = \? AND biz_no = \?`).
				WithArgs(
					int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
					int64(9),
					"EX-CAPACITY",
				).
				WillReturnRows(sqlmock.NewRows([]string{"complete"}).AddRow(testCase.complete))

			model := &customTOptionAssetInstructionModel{
				defaultTOptionAssetInstructionModel: &defaultTOptionAssetInstructionModel{
					CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
					table:      "`t_option_asset_instruction`",
				},
			}
			complete, err := model.AllSucceededByBizNo(context.Background(), 9, "EX-CAPACITY")
			if err != nil {
				t.Fatalf("check completion: %v", err)
			}
			if complete != testCase.complete {
				t.Fatalf("complete=%v want=%v", complete, testCase.complete)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRecoverStaleAssetInstructionsKeepsOriginalIdentity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`(?s)UPDATE t_option_asset_instruction.*SET status=\?,next_retry_at=\?,last_error_msg='STALE_PROCESSING_RECOVERED',update_times=\?.*WHERE status=\? AND update_times < \? AND tenant_id = \?`).
		WithArgs(
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
			int64(200), int64(200),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING),
			int64(140), int64(9),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	model := &customTOptionAssetInstructionModel{
		defaultTOptionAssetInstructionModel: &defaultTOptionAssetInstructionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_asset_instruction`",
		},
	}
	affected, err := model.RecoverStale(context.Background(), 9, 140, 200)
	if err != nil {
		t.Fatalf("recover stale: %v", err)
	}
	if affected != 1 {
		t.Fatalf("affected=%d want=1", affected)
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

	mock.ExpectQuery(`(?s)SELECT .* FROM .* AS current.*NOT EXISTS \(.*biz_previous\.execution_group.*current\.execution_group.*biz_previous\.step_no < current\.step_no.*biz_previous\.status <> \?.*\).*delivery_unit_id.*cure_deadline > \?.*ORDER BY id ASC LIMIT \?`).
		WithArgs(
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
			int64(1234),
			int64(0),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(1234),
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

func TestFindRunnableOutboxRequiresBuyerPremiumDebit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .* AS current.*EXISTS \(.*premium_debit\.trade_id = current\.trade_id.*premium_debit\.action = \?.*premium_debit\.step_no = 1.*premium_debit\.status = \?.*\).*combo_sibling\.combo_match_no = combo_current\.combo_match_no.*sibling_debit\.trade_id = combo_sibling\.id.*ORDER BY current\.id LIMIT \?`).
		WithArgs(
			int64(9),
			int64(9),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
			int64(1234),
			int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
			int64(100),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionOutboxModel{
		defaultTOptionOutboxModel: &defaultTOptionOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_outbox`",
		},
	}
	items, err := model.FindRunnable(context.Background(), 9, 1234, 100)
	if err != nil {
		t.Fatalf("find runnable outbox failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d want=0", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestComboDebitBarrierRequiresCompleteLegShapeAndAllDebits(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTOptionOutboxModel{
		defaultTOptionOutboxModel: &defaultTOptionOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_outbox`",
		},
	}
	queryPattern := `(?s)SELECT.*COUNT\(1\) trade_legs.*COUNT\(DISTINCT sibling\.combo_leg_no\).*missing_debits.*FROM t_option_trade sibling.*sibling\.tenant_id=\? AND sibling\.combo_match_no=\?`
	mock.ExpectQuery(queryPattern).
		WithArgs(
			int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(9), "MATCH-1",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"trade_legs", "distinct_legs", "min_leg", "max_leg", "missing_debits",
		}).AddRow(2, 2, 1, 2, 0))
	ready, err := model.ComboDebitBarrierReady(context.Background(), 9, "MATCH-1")
	if err != nil || !ready {
		t.Fatalf("complete barrier ready=%v err=%v", ready, err)
	}

	mock.ExpectQuery(queryPattern).
		WithArgs(
			int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
			int64(9), "MATCH-1",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"trade_legs", "distinct_legs", "min_leg", "max_leg", "missing_debits",
		}).AddRow(2, 1, 1, 1, 0))
	ready, err = model.ComboDebitBarrierReady(context.Background(), 9, "MATCH-1")
	if err != nil {
		t.Fatalf("query malformed barrier: %v", err)
	}
	if ready {
		t.Fatal("duplicate/missing leg numbers must keep the debit barrier closed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCountStaleComboDebitBarrierBlockedIsTenantScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTOptionOutboxModel{
		defaultTOptionOutboxModel: &defaultTOptionOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_outbox`",
		},
	}
	mock.ExpectQuery(`(?s)SELECT COUNT\(1\).*FROM .* current.*\(\?=0 OR current\.tenant_id=\?\).*current\.create_times<\?.*combo_current\.combo_match_no<>''.*sibling\.combo_match_no=combo_current\.combo_match_no.*debit\.trade_id=sibling\.id`).
		WithArgs(
			int64(9), int64(9),
			int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_MANUAL_REVIEW),
			int64(940),
			int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	count, err := model.CountStaleComboDebitBarrierBlocked(context.Background(), 9, 940)
	if err != nil {
		t.Fatalf("count stale combo barrier: %v", err)
	}
	if count != 3 {
		t.Fatalf("count=%d want=3", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindExerciseByClientExerciseIdUsesUserScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .* WHERE tenant_id = \? AND user_id = \? AND client_exercise_id = \? LIMIT 1`).
		WithArgs(int64(9), int64(88), "client-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "exercise_no", "client_exercise_id", "user_id", "account_id",
			"contract_id", "position_id", "exercise_type", "exercise_qty", "strike_price",
			"settlement_price", "exercise_amount", "profit_amount", "fee", "fee_coin", "status",
			"remark", "exercise_time", "finish_time", "create_times", "update_times",
		}).AddRow(
			int64(7), int64(9), "EX-7", "client-1", int64(88), int64(3),
			int64(11), int64(12), int64(option.ExerciseType_EXERCISE_TYPE_USER), "1", "100",
			"110", "100", "10", "0.1", "USDT", int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING),
			"", int64(1000), int64(0), int64(1000), int64(1000),
		))

	model := &customTOptionExerciseModel{
		defaultTOptionExerciseModel: &defaultTOptionExerciseModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_exercise`",
		},
	}
	item, err := model.FindOneByClientExerciseId(context.Background(), 9, 88, "client-1")
	if err != nil {
		t.Fatalf("find exercise failed: %v", err)
	}
	if item.Id != 7 || item.ClientExerciseId != "client-1" {
		t.Fatalf("unexpected exercise: %+v", item)
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

func TestSumPendingPositionDeltaIncludesUnappliedLongFills(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT COALESCE\(SUM\(t.qty\), 0\).*JOIN t_option_order o ON o.id = t.buy_order_id.*o.position_effect = \?.*o.user_id = \?`).
		WithArgs(
			int64(9),
			int64(88),
			int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
			int64(option.PositionEffect_POSITION_EFFECT_OPEN),
			int64(77),
			int64(77),
		).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("7"))
	mock.ExpectQuery(`(?s)SELECT COALESCE\(SUM\(t.qty\), 0\).*JOIN t_option_order o ON o.id = t.sell_order_id.*o.position_effect = \?.*o.user_id = \?`).
		WithArgs(
			int64(9),
			int64(88),
			int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
			int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
			int64(option.PositionEffect_POSITION_EFFECT_CLOSE),
			int64(77),
			int64(77),
		).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("2"))

	model := &customTOptionOutboxModel{
		defaultTOptionOutboxModel: &defaultTOptionOutboxModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_outbox`",
		},
	}
	got, err := model.SumPendingPositionDelta(
		context.Background(),
		9,
		77,
		88,
		int64(common.PositionSide_POSITION_SIDE_LONG),
	)
	if err != nil {
		t.Fatalf("sum pending position delta failed: %v", err)
	}
	if got.String() != "5" {
		t.Fatalf("pending delta=%s want=5", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureTradingControlCreatesAndLocksRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`(?s)INSERT IGNORE INTO t_option_user_trading_control.*VALUES \(\?,\?,2,\?,\?\)`).
		WithArgs(int64(9), int64(77), int64(1234), int64(1234)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND user_id = \? LIMIT 1 FOR UPDATE`).
		WithArgs(int64(9), int64(77)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "kill_switch", "reason",
			"activated_at", "released_at", "activated_by", "released_by",
			"create_times", "update_times",
		}).AddRow(
			int64(1), int64(9), int64(77), int64(common.YesNo_YES_NO_NO), "",
			int64(0), int64(0), int64(0), int64(0), int64(1234), int64(1234),
		))

	model := &customTOptionUserTradingControlModel{
		defaultTOptionUserTradingControlModel: &defaultTOptionUserTradingControlModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_user_trading_control`",
		},
	}
	item, err := model.EnsureForUpdate(context.Background(), 9, 77, 1234)
	if err != nil {
		t.Fatalf("ensure trading control failed: %v", err)
	}
	if item.UserId != 77 || item.KillSwitch != int64(common.YesNo_YES_NO_NO) {
		t.Fatalf("unexpected trading control: %+v", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSumHoldingQtyMapsDecimalAggregate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT COALESCE\(SUM\(position_qty\), 0\) AS total.*user_id = \?`).
		WithArgs(
			int64(9),
			int64(88),
			int64(common.PositionSide_POSITION_SIDE_LONG),
			int64(option.PositionStatus_POSITION_STATUS_HOLDING),
			int64(77),
		).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("12.5"))

	model := &customTOptionPositionModel{
		defaultTOptionPositionModel: &defaultTOptionPositionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_position`",
		},
	}
	got, err := model.SumHoldingQty(
		context.Background(),
		9,
		77,
		88,
		int64(common.PositionSide_POSITION_SIDE_LONG),
	)
	if err != nil {
		t.Fatalf("sum holding qty failed: %v", err)
	}
	if got.String() != "12.5" {
		t.Fatalf("holding qty=%s want=12.5", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSumActiveOpenQtyMapsDecimalAggregate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT COALESCE\(SUM\(unfilled_qty\), 0\) AS total.*user_id = \?`).
		WithArgs(
			int64(9),
			int64(88),
			int64(common.Side_SIDE_BUY),
			int64(option.PositionEffect_POSITION_EFFECT_OPEN),
			int64(option.OrderStatus_ORDER_STATUS_FUNDING),
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			int64(77),
		).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("3.25"))

	model := &customTOptionOrderModel{
		defaultTOptionOrderModel: &defaultTOptionOrderModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_order`",
		},
	}
	got, err := model.SumActiveOpenQty(
		context.Background(),
		9,
		77,
		88,
		int64(common.Side_SIDE_BUY),
	)
	if err != nil {
		t.Fatalf("sum active open qty failed: %v", err)
	}
	if got.String() != "3.25" {
		t.Fatalf("active open qty=%s want=3.25", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindMarketForUpdateLocksTenantContractRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND contract_id = \? LIMIT 1 FOR UPDATE`).
		WithArgs(int64(9), int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionMarketModel{
		defaultTOptionMarketModel: &defaultTOptionMarketModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_market`",
		},
	}
	_, err = model.FindOneByTenantIdContractIdForUpdate(context.Background(), 9, 88)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find market for update error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindMMPConfigForUpdateLocksExactGroup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND user_id = \? AND contract_id = \? AND group_code = \?.*LIMIT 1 FOR UPDATE`).
		WithArgs(int64(9), int64(77), int64(88), "desk-a").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionMmpConfigModel{
		defaultTOptionMmpConfigModel: &defaultTOptionMmpConfigModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_mmp_config`",
		},
	}
	_, err = model.FindForUpdate(context.Background(), 9, 77, 88, "desk-a")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find MMP config error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindActiveMMPOrdersScopesGroupAndStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND user_id = \? AND contract_id = \? AND mmp = \? AND mmp_group = \?.*status IN \(\?, \?, \?\).*ORDER BY id DESC LIMIT \?`).
		WithArgs(
			int64(9), int64(77), int64(88), int64(common.YesNo_YES_NO_YES), "desk-a",
			int64(option.OrderStatus_ORDER_STATUS_FUNDING),
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			int64(100),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionOrderModel{
		defaultTOptionOrderModel: &defaultTOptionOrderModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_order`",
		},
	}
	items, err := model.FindActiveMMPOrders(context.Background(), 9, 77, 88, "desk-a", 0, 100)
	if err != nil {
		t.Fatalf("find active MMP orders failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d want=0", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindFirstUnsafeMMPOrderForUpdateLocksTradingAndReleaseStates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND user_id = \? AND contract_id = \? AND mmp = \? AND mmp_group = \?.*status IN \(\?, \?, \?, \?, \?\).*ORDER BY id LIMIT 1 FOR UPDATE`).
		WithArgs(
			int64(9), int64(77), int64(88), int64(common.YesNo_YES_NO_YES), "desk-a",
			int64(option.OrderStatus_ORDER_STATUS_FUNDING),
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			int64(option.OrderStatus_ORDER_STATUS_CANCELING),
			int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionOrderModel{
		defaultTOptionOrderModel: &defaultTOptionOrderModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_order`",
		},
	}
	_, err = model.FindFirstUnsafeMMPOrderForUpdate(context.Background(), 9, 77, 88, "desk-a")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find residual MMP order error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindActivePortfolioRiskConfigRequiresApprovedEffectiveVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND settle_coin = \? AND status IN \(\?, \?\).*effective_from <= \?.*effective_until = 0 OR effective_until > \?.*ORDER BY effective_from DESC, version DESC LIMIT 1`).
		WithArgs(
			int64(9), "USDT",
			int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED),
			int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_SUPERSEDED),
			int64(1234), int64(1234),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionPortfolioRiskConfigModel{
		defaultTOptionPortfolioRiskConfigModel: &defaultTOptionPortfolioRiskConfigModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_portfolio_risk_config`",
		},
	}
	_, err = model.FindActive(context.Background(), 9, "USDT", 1234)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find active portfolio risk config error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindActiveTradeCorrectionLocksLatestCase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND trade_id = \? AND status IN \(\?, \?, \?\).*ORDER BY id DESC LIMIT 1 FOR UPDATE`).
		WithArgs(
			int64(9),
			int64(88),
			int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_PENDING_REVIEW),
			int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_EXECUTING),
			int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_MANUAL_REVIEW),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionTradeCorrectionModel{
		defaultTOptionTradeCorrectionModel: &defaultTOptionTradeCorrectionModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_trade_correction`",
		},
	}
	_, err = model.FindActiveByTradeForUpdate(context.Background(), 9, 88)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("find active correction error=%v want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindTradeCorrectionLegsUsesStableOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT .* FROM .*WHERE tenant_id = \? AND correction_id = \? ORDER BY leg_no ASC`).
		WithArgs(int64(9), int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	model := &customTOptionTradeCorrectionLegModel{
		defaultTOptionTradeCorrectionLegModel: &defaultTOptionTradeCorrectionLegModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_trade_correction_leg`",
		},
	}
	items, err := model.FindByCorrection(context.Background(), 9, 88)
	if err != nil {
		t.Fatalf("find correction legs failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d want=0", len(items))
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
