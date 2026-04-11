package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
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
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.TradeNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	item.EventStatus = int64(trade.EventStatus_EVENT_STATUS_PENDING)
	item.RetryCount++
	item.NextRetryAt = utils.NowMillis()
	item.OperatorId = in.OperatorId
	item.LastErrorMsg = ""
	item.UpdateTimes = utils.NowMillis()
	if err = l.svcCtx.BizTradeEventModel.Update(l.ctx, item); err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
