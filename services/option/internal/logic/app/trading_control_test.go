package applogic

import (
	"context"
	"regexp"
	"testing"

	"wklive/services/option/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestOptionOrderPriceBandIncludesBoundaries(t *testing.T) {
	mark := decimal.NewFromInt(100)
	ratio := decimal.RequireFromString("0.10")
	for _, test := range []struct {
		price string
		ok    bool
	}{
		{price: "89.99", ok: false},
		{price: "90", ok: true},
		{price: "100", ok: true},
		{price: "110", ok: true},
		{price: "110.01", ok: false},
	} {
		lower, upper, ok := optionOrderPriceBand(
			decimal.RequireFromString(test.price), mark, ratio,
		)
		if ok != test.ok {
			t.Fatalf("price=%s band=[%s,%s] ok=%t want=%t", test.price, lower, upper, ok, test.ok)
		}
	}
}

func TestOptionOrderPriceBandRejectsMissingControl(t *testing.T) {
	if _, _, ok := optionOrderPriceBand(
		decimal.NewFromInt(100), decimal.NewFromInt(100), decimal.Zero,
	); ok {
		t.Fatal("zero price band must mean not configured, not unlimited")
	}
}

func TestOptionExposureLimitExceeded(t *testing.T) {
	limit := decimal.NewFromInt(10)
	if optionExposureLimitExceeded(decimal.NewFromInt(7), decimal.NewFromInt(3), limit) {
		t.Fatal("exact limit boundary should be admitted")
	}
	if !optionExposureLimitExceeded(decimal.NewFromInt(7), decimal.RequireFromString("3.01"), limit) {
		t.Fatal("exposure above limit must be rejected")
	}
	if !optionExposureLimitExceeded(decimal.Zero, decimal.NewFromInt(1), decimal.Zero) {
		t.Fatal("zero limit must be treated as unconfigured and rejected")
	}
}

func TestRecordOrderTradingControlAuditCountsEveryEvaluation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn := sqlx.NewSqlConnFromDB(db)
	order := &models.TOptionOrder{Id: 7, TenantId: 9, UserId: 10, ContractId: 11}
	insertPattern := regexp.QuoteMeta(`INSERT INTO t_option_trading_control_event
(tenant_id,user_id,contract_id,order_id,event_type,reason,detail,operator_id,create_times)
VALUES(?,?,?,?,?,?,?,?,?)`)
	mock.ExpectExec(insertPattern).
		WithArgs(int64(9), int64(10), int64(11), int64(7),
			controlEventOrderRejected, controlReasonPriceBand, "outside band", int64(10), int64(1000)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insertPattern).
		WithArgs(int64(9), int64(10), int64(11), int64(7),
			controlEventOrderEvaluated, controlReasonEvaluation,
			"rejected reason=ORDER_PRICE_BAND", int64(10), int64(1000)).
		WillReturnResult(sqlmock.NewResult(2, 1))
	if err := recordOrderTradingControlAudit(context.Background(), conn, order,
		&orderControlRejection{reason: controlReasonPriceBand, detail: "outside band"}, 1000); err != nil {
		t.Fatalf("record rejected audit: %v", err)
	}
	mock.ExpectExec(insertPattern).
		WithArgs(int64(9), int64(10), int64(11), int64(7),
			controlEventOrderEvaluated, controlReasonEvaluation, "accepted", int64(10), int64(1001)).
		WillReturnResult(sqlmock.NewResult(3, 1))
	if err := recordOrderTradingControlAudit(context.Background(), conn, order, nil, 1001); err != nil {
		t.Fatalf("record accepted audit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
