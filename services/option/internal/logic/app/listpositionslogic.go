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

type ListPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPositionsLogic {
	return &ListPositionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓列表
func (l *ListPositionsLogic) ListPositions(in *option.UserListPositionsReq) (*option.UserListPositionsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := l.svcCtx.OptionPositionModel.FindPage(l.ctx, models.OptionPositionPageFilter{
		TenantId:  tenantId,
		UserId:    userId,
		AccountId: in.AccountId,
		Status:    int64(in.Status),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	data := make([]*option.OptionPositionDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := helpers.BuildPositionDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		data = append(data, detail)
	}

	return &option.UserListPositionsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data,
	}, nil
}
