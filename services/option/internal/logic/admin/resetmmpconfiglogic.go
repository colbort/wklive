package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetMMPConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetMMPConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetMMPConfigLogic {
	return &ResetMMPConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人工复核并恢复触发中的 MMP 报价组
func (l *ResetMMPConfigLogic) ResetMMPConfig(in *option.ResetMMPConfigReq) (*option.GetMMPConfigResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	groupCode, validGroup := applogic.NormalizeMMPGroup(in.GroupCode)
	reason := strings.TrimSpace(in.Reason)
	if forbidden || !allowed {
		return &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	if in.TenantId <= 0 || in.UserId <= 0 || in.ContractId <= 0 || operatorID <= 0 ||
		!validGroup || reason == "" || len(reason) > 500 {
		return &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	item, err := l.svcCtx.OptionMmpConfigModel.FindOneByTenantIdUserIdContractIdGroupCode(
		l.ctx, in.TenantId, in.UserId, in.ContractId, groupCode,
	)
	if errors.Is(err, models.ErrNotFound) {
		return &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, "MMP_NOT_CONFIGURED"),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	input := &mmpConfigInput{
		tenantID: in.TenantId, userID: in.UserId, contractID: in.ContractId,
		groupCode: groupCode, enabled: int64(common.YesNo_YES_NO_YES),
		qtyThreshold: item.QtyThreshold, tradeCountThreshold: item.TradeCountThreshold,
		lossThreshold: item.LossThreshold, windowSeconds: item.WindowSeconds,
		cooldownSeconds: item.CooldownSeconds, reason: reason, operatorID: operatorID,
	}
	if err := stageMMPReset(l.ctx, l.svcCtx, input); err != nil {
		if errors.Is(err, errMMPNotTriggered) {
			return &option.GetMMPConfigResp{
				Base: helper.ErrResp(i18n.OperationNotAllowed, errMMPNotTriggered.Error()),
			}, nil
		}
		return nil, err
	}
	total, success, failed, err := applogic.CancelMMPGroupOrdersReport(
		l.ctx, l.svcCtx, in.TenantId, in.UserId, in.ContractId,
		groupCode, "MMP_MANUAL_RESET", true,
	)
	if err != nil {
		applogic.SetMMPConfigLastError(
			l.ctx, l.svcCtx, in.TenantId, in.UserId, in.ContractId, groupCode, err.Error(),
		)
		return nil, err
	}
	item, err = activateStagedMMPConfig(l.ctx, l.svcCtx, input, "MMP_MANUAL_RESET")
	if err != nil {
		if errors.Is(err, errMMPNotTriggered) {
			return &option.GetMMPConfigResp{
				Base: helper.ErrResp(i18n.OperationNotAllowed, errMMPNotTriggered.Error()),
			}, nil
		}
		if errors.Is(err, errMMPReleasePending) {
			return &option.GetMMPConfigResp{
				Base: helper.ErrResp(i18n.OperationNotAllowed, err.Error()),
			}, nil
		}
		applogic.SetMMPConfigLastError(
			l.ctx, l.svcCtx, in.TenantId, in.UserId, in.ContractId, groupCode, err.Error(),
		)
		return nil, err
	}
	l.Infof(
		"option mmp manual reset tenantId=%d userId=%d contractId=%d group=%s total=%d success=%d failed=%d operatorId=%d",
		in.TenantId, in.UserId, in.ContractId, groupCode, total, success, failed, operatorID,
	)
	return mmpConfigResponse(item), nil
}
