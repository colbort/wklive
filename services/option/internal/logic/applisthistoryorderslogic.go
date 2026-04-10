package logic

import (
	"context"
	"errors"

	pageutil "wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListHistoryOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListHistoryOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListHistoryOrdersLogic {
	return &AppListHistoryOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取历史委托列表
func (l *AppListHistoryOrdersLogic) AppListHistoryOrders(in *option.AppListHistoryOrdersReq) (*option.AppListHistoryOrdersResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	filter := models.OptionOrderPageFilter{
		TenantId:        in.TenantId,
		Uid:             in.Uid,
		AccountId:       in.AccountId,
		ContractId:      in.ContractId,
		Status:          int64(in.Status),
		CreateTimeStart: pageutil.TimeRangeStart(in.CreateTimeRange),
		CreateTimeEnd:   pageutil.TimeRangeEnd(in.CreateTimeRange),
	}
	if in.Status == 0 {
		filter.Statuses = []int64{
			int64(option.OrderStatus_ORDER_STATUS_FILLED),
			int64(option.OrderStatus_ORDER_STATUS_CANCELED),
			int64(option.OrderStatus_ORDER_STATUS_REJECTED),
			int64(option.OrderStatus_ORDER_STATUS_EXPIRED),
		}
	}
	items, total, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, filter, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionOrderDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := buildOrderDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		list = append(list, detail)
	}

	return &option.AppListHistoryOrdersResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
