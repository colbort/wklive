package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayChannelRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayChannelRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelRuleLogic {
	return &GetTenantPayChannelRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取通道规则详情
func (l *GetTenantPayChannelRuleLogic) GetTenantPayChannelRule(in *payment.GetTenantPayChannelRuleReq) (*payment.GetTenantPayChannelRuleResp, error) {
	var (
		errLogic = "GetTenantPayChannelRule"
	)

	rule, err := l.svcCtx.TenantPayChannelRuleModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if rule == nil {
		return &payment.GetTenantPayChannelRuleResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotifyChannelRuleNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetTenantPayChannelRuleResp{
		Base: helper.OkResp(),
		Data: toTenantPayChannelRuleProto(rule),
	}, nil
}
