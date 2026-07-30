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

type RetryLiquidationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryLiquidationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryLiquidationLogic {
	return &RetryLiquidationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 将失败或人工处理的强平记录重新置为待执行
func (l *RetryLiquidationLogic) RetryLiquidation(in *option.RetryLiquidationReq) (*option.CommonResp, error) {
	item, err := l.svcCtx.OptionLiquidationModel.FindOne(l.ctx, in.LiquidationId)
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
	instructions, err := l.svcCtx.OptionAssetInstructionModel.FindByLiquidation(l.ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	for _, instruction := range instructions {
		if _, err := l.svcCtx.OptionAssetInstructionModel.ResetForManualRetry(
			l.ctx, instruction.Id, time.Now().Unix(),
		); err != nil {
			return nil, err
		}
	}
	changed, err := l.svcCtx.OptionLiquidationModel.ResetForManualRetry(l.ctx, item.Id, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	if !changed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
