package applogic

import (
	"context"
	"errors"
	"wklive/services/option/internal/logic/helpers"

	pageutil "wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCurrentOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCurrentOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCurrentOrdersLogic {
	return &ListCurrentOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前委托列表
func (l *ListCurrentOrdersLogic) ListCurrentOrders(in *option.ListCurrentOrdersReq) (*option.ListCurrentOrdersResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
		TenantId:             tenantId,
		UserId:               userId,
		AccountId:            in.AccountId,
		ContractId:           in.ContractId,
		Side:                 int64(in.Side),
		ExcludeComboChildren: true,
		Statuses: []int64{
			int64(option.OrderStatus_ORDER_STATUS_PENDING),
			int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
		},
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	data := make([]*option.OptionOrderDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := helpers.BuildOrderDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		data = append(data, detail)
	}

	return &option.ListCurrentOrdersResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data,
	}, nil
}
