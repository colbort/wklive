package models

import "wklive/common/sqlutil"

// AdminPageFilter is shared by the read-only trade operation models.
// Each model only applies the fields supported by its own table.
type AdminPageFilter struct {
	TenantId     int64
	SymbolId     int64
	BatchId      int64
	UserId       int64
	PositionId   int64
	OrderId      int64
	Status       int64
	Enabled      int64
	SnapshotType int64
	BizType      string
	BizId        string
	TimeStart    int64
	TimeEnd      int64
}

func adminPageBuilder(filter AdminPageFilter, timeColumn string) *sqlutil.PageQueryBuilder {
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("tenant_id", filter.TenantId)
	b.EqInt64("symbol_id", filter.SymbolId)
	b.EqInt64("batch_id", filter.BatchId)
	b.EqInt64("user_id", filter.UserId)
	b.EqInt64("position_id", filter.PositionId)
	b.EqInt64("order_id", filter.OrderId)
	b.EqInt64("status", filter.Status)
	b.EqInt64("enabled", filter.Enabled)
	b.EqInt64("snapshot_type", filter.SnapshotType)
	b.EqString("biz_type", filter.BizType)
	b.EqString("biz_id", filter.BizId)
	if timeColumn != "" {
		b.GteInt64(timeColumn, filter.TimeStart)
		b.LteInt64(timeColumn, filter.TimeEnd)
	}
	return b
}
