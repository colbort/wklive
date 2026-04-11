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

type GetFillDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFillDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillDetailAdminLogic {
	return &GetFillDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交详情
func (l *GetFillDetailAdminLogic) GetFillDetailAdmin(in *trade.GetFillDetailAdminReq) (*trade.GetFillDetailAdminResp, error) {
	item, err := l.svcCtx.TradeFillModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetFillDetailAdminResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.TradeNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	return &trade.GetFillDetailAdminResp{Base: helper.OkResp(), Data: fillToProto(item)}, nil
}
