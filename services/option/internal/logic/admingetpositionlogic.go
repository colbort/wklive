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

type AdminGetPositionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetPositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetPositionLogic {
	return &AdminGetPositionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个持仓详情
func (l *AdminGetPositionLogic) AdminGetPosition(in *option.GetPositionReq) (*option.GetPositionResp, error) {
	item, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetPositionResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && item.TenantId != in.TenantId {
		return &option.GetPositionResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
	}
	data, err := buildPositionDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetPositionResp{Base: helper.OkResp(), Data: data}, nil
}
