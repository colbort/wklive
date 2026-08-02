package models

import (
	"context"
	"strings"
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSimpleBookQueriesExcludeComboChildren(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	orderModel := &customTOptionOrderModel{
		defaultTOptionOrderModel: &defaultTOptionOrderModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_order`",
		},
	}
	mock.ExpectQuery(`(?s)SELECT price, SUM\(unfilled_qty\).*combo_order_count.*WHERE tenant_id=\? AND contract_id=\? AND side=\? AND status IN \(\?,\?\).*order_type IN \(\?,\?\).*combo_order_id=0`).
		WithArgs(
			int64(9), int64(88), int64(common.Side_SIDE_BUY),
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			int64(option.OrderType_ORDER_TYPE_LIMIT),
			int64(option.OrderType_ORDER_TYPE_POST_ONLY),
			int64(20),
		).
		WillReturnRows(sqlmock.NewRows([]string{"price", "qty", "order_count", "combo_order_count"}))
	if _, err = orderModel.FindOrderBookLevels(
		context.Background(), 9, 88, int64(common.Side_SIDE_BUY), 20,
	); err != nil {
		t.Fatalf("find book levels: %v", err)
	}

	mock.ExpectQuery(`(?s)WHERE tenant_id = \? AND contract_id = \? AND side = \?.*combo_order_id = 0 AND user_id <> \?`).
		WithArgs(
			int64(9), int64(88), int64(common.Side_SIDE_SELL), int64(77),
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			decimal.RequireFromString("10"),
			int64(10),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if _, err = orderModel.FindMatchableOrders(
		context.Background(), 9, 88, int64(common.Side_SIDE_SELL),
		77, 0, decimal.RequireFromString("10"), 10,
	); err != nil {
		t.Fatalf("find simple match candidates: %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestComboCandidateQueryUsesInverseStrategyPriceTimePriority(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &customTOptionComboOrderModel{
		defaultTOptionComboOrderModel: &defaultTOptionComboOrderModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
			table:      "`t_option_combo_order`",
		},
	}
	mock.ExpectQuery(`(?s)WHERE tenant_id=\? AND strategy_key=\? AND user_id<>\?.*net_price>=\?.*ORDER BY net_price DESC, id ASC`).
		WithArgs(
			int64(9), "inverse-hash", int64(0),
			int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE),
			int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED),
			decimal.RequireFromString("-3"), int64(50),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	items, err := model.FindMatchCandidates(
		context.Background(), 9, "inverse-hash", 0,
		decimal.RequireFromString("-3"), 50,
	)
	if err != nil {
		t.Fatalf("find combo candidates: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%d want 0", len(items))
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestComboIdempotencyLookupBypassesCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := &defaultTOptionComboOrderModel{
		CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil),
		table:      "`t_option_combo_order`",
	}
	mock.ExpectQuery(`SELECT .* FROM .* WHERE tenant_id=\? AND user_id=\? AND client_combo_id=\? LIMIT 1`).
		WithArgs(int64(9), int64(10), "same-key").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "combo_no", "user_id", "account_id", "client_combo_id",
			"strategy_key", "inverse_strategy_key", "underlying_symbol", "expire_time",
			"settle_coin", "quote_coin", "order_type", "net_price", "qty", "filled_qty",
			"unfilled_qty", "status", "payload_hash", "cancel_reason", "cancel_time",
			"create_times", "update_times",
		}).AddRow(
			1, 9, "OC-1", 10, 11, "same-key", strings.Repeat("a", 64), strings.Repeat("b", 64),
			"BTCUSDT", 2000000000, "USDT", "USDT", 1, "1", "1", "0", "1", 1,
			strings.Repeat("c", 64), "", 0, 100, 100,
		))
	item, err := model.FindOneByTenantIdUserIdClientComboIdNoCache(
		context.Background(), 9, 10, "same-key",
	)
	if err != nil || item == nil || item.Id != 1 {
		t.Fatalf("no-cache combo lookup item=%+v err=%v", item, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdminComboQueriesStayTenantScopedAndUseParentRelationships(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn := sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), nil)
	comboModel := &customTOptionComboOrderModel{
		defaultTOptionComboOrderModel: &defaultTOptionComboOrderModel{
			CachedConn: conn, table: "`t_option_combo_order`",
		},
	}
	tradeModel := &customTOptionTradeModel{
		defaultTOptionTradeModel: &defaultTOptionTradeModel{
			CachedConn: conn, table: "`t_option_trade`",
		},
	}
	instructionModel := &customTOptionAssetInstructionModel{
		defaultTOptionAssetInstructionModel: &defaultTOptionAssetInstructionModel{
			CachedConn: conn, table: "`t_option_asset_instruction`",
		},
	}

	mock.ExpectQuery(`(?s)SELECT COUNT\(1\).*tenant_id = \?.*combo_no LIKE \?.*underlying_symbol LIKE \?.*status = \?.*create_times >= \?.*create_times <= \?`).
		WithArgs(
			int64(9), int64(77), int64(88), "%COMBO-9%", "%BTC%",
			int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_MANUAL_REVIEW),
			int64(100), int64(200),
		).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`(?s)SELECT .* FROM .*tenant_id = \?.*combo_no LIKE \?.*underlying_symbol LIKE \?.*ORDER BY id DESC LIMIT \?`).
		WithArgs(
			int64(9), int64(77), int64(88), "%COMBO-9%", "%BTC%",
			int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_MANUAL_REVIEW),
			int64(100), int64(200), int64(20),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if _, _, err = comboModel.FindPage(context.Background(), OptionComboOrderPageFilter{
		TenantId: 9, UserId: 77, AccountId: 88, ComboNo: "COMBO-9",
		UnderlyingSymbol: "BTC",
		Status:           int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_MANUAL_REVIEW),
		CreateTimeStart:  100, CreateTimeEnd: 200,
	}, 0, 20); err != nil {
		t.Fatalf("find admin combo page: %v", err)
	}

	mock.ExpectQuery(`(?s)SELECT COUNT\(1\).*t_option_trade.*trade\.tenant_id=\?.*child\.combo_order_id=\?.*child\.id=trade\.buy_order_id`).
		WithArgs(int64(9), int64(901)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`(?s)SELECT .*t_option_trade.*trade\.tenant_id=\?.*child\.combo_order_id=\?.*ORDER BY trade\.combo_match_no,trade\.combo_leg_no,trade\.id.*LIMIT \?`).
		WithArgs(int64(9), int64(901), int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	trades, tradeTotal, err := tradeModel.FindByComboOrderID(
		context.Background(), 9, 901, 500,
	)
	if err != nil || len(trades) != 0 || tradeTotal != 0 {
		t.Fatalf("find combo trades: len=%d total=%d err=%v", len(trades), tradeTotal, err)
	}

	mock.ExpectQuery(`(?s)SELECT COUNT\(1\).*t_option_asset_instruction.*instruction\.tenant_id=\?.*child\.combo_order_id=\?.*child\.id=instruction\.order_id`).
		WithArgs(int64(9), int64(901)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`(?s)SELECT .*t_option_asset_instruction.*instruction\.tenant_id=\?.*child\.combo_order_id=\?.*ORDER BY instruction\.order_id,instruction\.step_no,instruction\.id.*LIMIT \?`).
		WithArgs(int64(9), int64(901), int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	instructions, instructionTotal, err := instructionModel.FindByComboOrderID(
		context.Background(), 9, 901, 500,
	)
	if err != nil || len(instructions) != 0 || instructionTotal != 0 {
		t.Fatalf(
			"find combo instructions: len=%d total=%d err=%v",
			len(instructions), instructionTotal, err,
		)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
