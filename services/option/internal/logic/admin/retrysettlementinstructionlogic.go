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

type RetrySettlementInstructionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrySettlementInstructionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySettlementInstructionLogic {
	return &RetrySettlementInstructionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 校验归属后重试结算批次中的失败资产指令
func (l *RetrySettlementInstructionLogic) RetrySettlementInstruction(in *option.RetrySettlementInstructionReq) (*option.CommonResp, error) {
	settlement, err := l.svcCtx.OptionSettlementModel.FindOne(l.ctx, in.SettlementId)
	if errors.Is(err, models.ErrNotFound) ||
		(err == nil && in.TenantId > 0 && settlement.TenantId != in.TenantId) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.OptionAssetInstructionModel.FindOne(l.ctx, in.InstructionId)
	if errors.Is(err, models.ErrNotFound) ||
		(err == nil && (item.TenantId != settlement.TenantId ||
			item.BizNo != settlement.SettlementNo || item.PositionId <= 0)) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, settlement.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	reset, err := l.svcCtx.OptionAssetInstructionModel.ResetForManualRetry(
		l.ctx, item.Id, time.Now().Unix(),
	)
	if err != nil {
		return nil, err
	}
	if !reset {
		return &option.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
