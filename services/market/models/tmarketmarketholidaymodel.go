package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TMarketMarketHolidayModel = (*customTMarketMarketHolidayModel)(nil)

type (
	// TMarketMarketHolidayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketMarketHolidayModel.
	TMarketMarketHolidayModel interface {
		tMarketMarketHolidayModel
		UpsertByCalendarDate(ctx context.Context, data *TMarketMarketHoliday) error
	}

	customTMarketMarketHolidayModel struct {
		*defaultTMarketMarketHolidayModel
	}
)

// NewTMarketMarketHolidayModel returns a model for the database table.
func NewTMarketMarketHolidayModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketMarketHolidayModel {
	return &customTMarketMarketHolidayModel{
		defaultTMarketMarketHolidayModel: newTMarketMarketHolidayModel(conn, c, opts...),
	}
}

func (m *defaultTMarketMarketHolidayModel) UpsertByCalendarDate(ctx context.Context, data *TMarketMarketHoliday) error {
	query := fmt.Sprintf(`INSERT INTO %s (calendar_id,trade_date,day_type,open_time,close_time,remark)
		VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),day_type=VALUES(day_type),
		open_time=VALUES(open_time),close_time=VALUES(close_time),remark=VALUES(remark)`, m.table)
	result, err := m.ExecNoCacheCtx(ctx, query, data.CalendarId, data.TradeDate, data.DayType, data.OpenTime, data.CloseTime, data.Remark)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	data.Id = id
	indexKey := fmt.Sprintf("%s%v:%v", cacheTMarketMarketHolidayCalendarIdTradeDatePrefix, data.CalendarId, data.TradeDate)
	idKey := fmt.Sprintf("%s%v", cacheTMarketMarketHolidayIdPrefix, id)
	return m.DelCache(indexKey, idKey)
}
