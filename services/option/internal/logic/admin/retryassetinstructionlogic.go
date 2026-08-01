package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryAssetInstructionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryAssetInstructionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryAssetInstructionLogic {
	return &RetryAssetInstructionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 将失败或人工处理的资产指令重新置为待执行
func (l *RetryAssetInstructionLogic) RetryAssetInstruction(in *option.RetryAssetInstructionReq) (*option.CommonResp, error) {
	reason := strings.TrimSpace(in.Reason)
	if !validAssetInstructionManualRetryReason(reason) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	item, err := l.svcCtx.OptionAssetInstructionModel.FindOne(l.ctx, in.InstructionId)
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
	// 实物交割必须以完整交割单元恢复，防止绕过补资确认、批次状态和操作原因审计。
	if item.DeliveryUnitId > 0 {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	if operatorID <= 0 {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	reset, err := resetAssetInstructionWithAudit(
		l.ctx, l.svcCtx, item.Id, item.TenantId, 0, operatorID, reason,
	)
	if err != nil {
		return nil, err
	}
	if !reset {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
