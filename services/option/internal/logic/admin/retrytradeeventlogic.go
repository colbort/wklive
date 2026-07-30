package adminlogic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

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

// 将失败或人工处理的成交持仓事件重新置为待执行
func (l *RetryTradeEventLogic) RetryTradeEvent(in *option.RetryTradeEventReq) (*option.CommonResp, error) {
	item, err := l.svcCtx.OptionOutboxModel.FindOne(l.ctx, in.EventId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && in.TenantId > 0 && item.TenantId != in.TenantId) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	_, err = l.svcCtx.OptionOutboxModel.ResetForManualRetry(l.ctx, item.Id, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
