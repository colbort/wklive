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

type GetTradeEventDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTradeEventDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeEventDetailLogic {
	return &GetTradeEventDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易事件详情
func (l *GetTradeEventDetailLogic) GetTradeEventDetail(in *trade.GetTradeEventDetailReq) (*trade.GetTradeEventDetailResp, error) {
	item, err := l.svcCtx.BizTradeEventModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetTradeEventDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.TradeNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	return &trade.GetTradeEventDetailResp{Base: helper.OkResp(), Data: tradeEventToProto(item)}, nil
}
