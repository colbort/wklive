package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

type AppGetPositionDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppGetPositionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetPositionDetailLogic {
	return &AppGetPositionDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个持仓详情
func (l *AppGetPositionDetailLogic) AppGetPositionDetail(in *option.AppGetPositionDetailReq) (*option.AppGetPositionDetailResp, error) {
	item, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, in.PositionId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppGetPositionDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId || item.Uid != in.Uid || item.AccountId != in.AccountId {
		return &option.AppGetPositionDetailResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionViewPosition, l.ctx))}, nil
	}
	data, err := buildPositionDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.AppGetPositionDetailResp{Base: helper.OkResp(), Data: data}, nil
}
