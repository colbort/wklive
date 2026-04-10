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

type AppListCurrentOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListCurrentOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListCurrentOrdersLogic {
	return &AppListCurrentOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前委托列表
func (l *AppListCurrentOrdersLogic) AppListCurrentOrders(in *option.AppListCurrentOrdersReq) (*option.AppListCurrentOrdersResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
		TenantId:   in.TenantId,
		Uid:        in.Uid,
		AccountId:  in.AccountId,
		ContractId: in.ContractId,
		Side:       int64(in.Side),
		Statuses: []int64{
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		},
	}, cursor, limit)
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

	return &option.AppListCurrentOrdersResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
