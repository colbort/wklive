package models

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestCountCorporateActionExecutionBlockersResolvesAssetInstructionContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT.*FROM t_option_asset_instruction ai.*LEFT JOIN t_option_order ai_order.*LEFT JOIN t_option_trade ai_trade.*LEFT JOIN t_option_position ai_position.*LEFT JOIN t_option_margin_lot ai_margin_lot.*LEFT JOIN t_option_liquidation ai_liquidation.*LEFT JOIN t_option_physical_delivery_unit ai_delivery_unit.*AND \? IN \(.*ai_order.contract_id.*ai_delivery_unit.contract_id`).
		WithArgs(
			int64(9), int64(88), int64(9), int64(88),
			int64(9), int64(88), int64(9), int64(88),
			int64(9), int64(88), int64(9), int64(88),
			int64(9), int64(88), int64(9), int64(88),
		).
		WillReturnRows(sqlmock.NewRows([]string{"blocked"}).AddRow(3))

	blocked, err := CountCorporateActionExecutionBlockers(
		context.Background(), sqlx.NewSqlConnFromDB(db), 9, 88,
	)
	if err != nil {
		t.Fatalf("count blockers: %v", err)
	}
	if blocked != 3 {
		t.Fatalf("blocked=%d want=3", blocked)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
