package tradeadminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryTradeEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryTradeEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryTradeEventLogic {
	return &RetryTradeEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重试交易事件
func (l *RetryTradeEventLogic) RetryTradeEvent(in *trade.RetryTradeEventReq) (*trade.AdminCommonResp, error) {
	item, err := l.svcCtx.BizTradeEventModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.TradeNotFound, i18n.Translate(i18n.TradeNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, item.TenantId, i18n.TradeNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.AdminCommonResp{Base: base}, nil
	}
	now := utils.NowMillis()
	changed, err := l.svcCtx.BizTradeEventModel.ResetForManualRetry(l.ctx, item.Id, in.OperatorId, now)
	if err != nil {
		return nil, err
	}
	if !changed {
		return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
