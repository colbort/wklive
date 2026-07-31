package applogic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListComboOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListComboOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListComboOrdersLogic {
	return &ListComboOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询当前用户组合订单
func (l *ListComboOrdersLogic) ListComboOrders(in *option.ListComboOrdersReq) (*option.ListComboOrdersResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.OptionComboOrderModel.FindPage(
		l.ctx,
		models.OptionComboOrderPageFilter{
			TenantId: tenantID, UserId: userID, AccountId: in.AccountId,
			Status: int64(in.Status),
		},
		cursor, limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	data := make([]*option.OptionComboOrderDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, buildErr := buildComboOrderDetail(l.ctx, l.svcCtx, item)
		if buildErr != nil {
			return nil, buildErr
		}
		data = append(data, detail)
	}
	return &option.ListComboOrdersResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data,
	}, nil
}
