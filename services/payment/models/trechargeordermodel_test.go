package models

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/proto/payment"
)

func TestRechargeCreditStateUsesLockedIdentityAndPartialUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	model := newRechargeOrderSQLMockModel(db)
	identity := &RechargeOrderCreditIdentity{
		TenantId:   9,
		OrderNo:    "PAY-99",
		BizOrderNo: sql.NullString{String: "THIRD-99", Valid: true},
	}

	mock.ExpectQuery(`(?s)SELECT tenant_id, order_no, biz_order_no FROM .* WHERE id = \? LIMIT 1 FOR UPDATE`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "order_no", "biz_order_no"}).
			AddRow(identity.TenantId, identity.OrderNo, identity.BizOrderNo))
	locked, err := model.FindCreditIdentityForUpdate(context.Background(), 99)
	if err != nil {
		t.Fatalf("lock credit identity: %v", err)
	}
	if locked.TenantId != identity.TenantId || locked.OrderNo != identity.OrderNo {
		t.Fatalf("locked identity=%+v want=%+v", locked, identity)
	}

	successPattern := regexp.QuoteMeta("UPDATE `t_recharge_order` SET credit_status = ?, credited_time = ?, last_credit_error = '', update_times = ? WHERE id = ? AND tenant_id = ? AND order_no = ?")
	mock.ExpectExec(successPattern).
		WithArgs(int64(payment.CreditStatus_CREDIT_STATUS_SUCCESS), int64(20_000), int64(20_000), int64(99), identity.TenantId, identity.OrderNo).
		WillReturnResult(sqlmock.NewResult(0, 1))
	updated, err := model.MarkCreditSuccess(context.Background(), 99, identity, 20_000)
	if err != nil || !updated {
		t.Fatalf("mark credit success=(%t,%v), want true,nil", updated, err)
	}

	retryPattern := regexp.QuoteMeta("UPDATE `t_recharge_order` SET credit_status = ?, credit_retry_count = ?, last_credit_error = ?, update_times = ? WHERE id = ? AND tenant_id = ? AND order_no = ?")
	mock.ExpectExec(retryPattern).
		WithArgs(int64(payment.CreditStatus_CREDIT_STATUS_PROCESSING), int64(4), "temporary", int64(21_000), int64(99), identity.TenantId, identity.OrderNo).
		WillReturnResult(sqlmock.NewResult(0, 0))
	updated, err = model.MarkCreditRetrying(context.Background(), 99, identity, 4, 21_000, "temporary")
	if err != nil {
		t.Fatalf("mark credit retrying: %v", err)
	}
	if updated {
		t.Fatal("identity mismatch must not report an updated recharge order")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func newRechargeOrderSQLMockModel(db *sql.DB) *customTRechargeOrderModel {
	return &customTRechargeOrderModel{
		defaultTRechargeOrderModel: &defaultTRechargeOrderModel{
			CachedConn: sqlc.NewConnWithCache(sqlx.NewSqlConnFromDB(db), noopCache{}),
			table:      "`t_recharge_order`",
		},
	}
}
