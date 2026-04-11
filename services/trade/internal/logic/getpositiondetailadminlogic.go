package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionDetailAdminLogic {
	return &GetPositionDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓详情
func (l *GetPositionDetailAdminLogic) GetPositionDetailAdmin(in *trade.GetPositionDetailAdminReq) (*trade.GetPositionDetailAdminResp, error) {
	item, err := l.svcCtx.ContractPositionModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetPositionDetailAdminResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	return &trade.GetPositionDetailAdminResp{Base: helper.OkResp(), Data: positionToProto(item)}, nil
}
