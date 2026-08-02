package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpsertMMPConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpsertMMPConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertMMPConfigLogic {
	return &UpsertMMPConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建或更新做市商保护参数；任何变更都会撤销该组已有 MMP 报价
func (l *UpsertMMPConfigLogic) UpsertMMPConfig(in *option.UpsertMMPConfigReq) (*option.GetMMPConfigResp, error) {
	input, rejection, err := validateMMPConfigInput(l.ctx, in)
	if err != nil || rejection != nil {
		return rejection, err
	}
	staged, err := stageMMPConfig(l.ctx, l.svcCtx, input)
	if err != nil {
		return nil, err
	}
	total, success, failed, err := applogic.CancelMMPGroupOrdersReport(
		l.ctx, l.svcCtx, input.tenantID, input.userID, input.contractID,
		input.groupCode, "MMP_CONFIG_CHANGED", true,
	)
	if err != nil {
		applogic.SetMMPConfigLastError(
			l.ctx, l.svcCtx, input.tenantID, input.userID, input.contractID,
			input.groupCode, err.Error(),
		)
		return nil, err
	}
	item, err := activateStagedMMPConfig(l.ctx, l.svcCtx, input, "")
	if err != nil {
		if errors.Is(err, errMMPReleasePending) {
			return &option.GetMMPConfigResp{
				Base: helper.ErrResp(i18n.OperationNotAllowed, err.Error()),
			}, nil
		}
		applogic.SetMMPConfigLastError(
			l.ctx, l.svcCtx, input.tenantID, input.userID, input.contractID,
			input.groupCode, err.Error(),
		)
		return nil, err
	}
	l.Infof(
		"option mmp configured tenantId=%d userId=%d contractId=%d group=%s enabled=%d total=%d success=%d failed=%d configId=%d",
		input.tenantID, input.userID, input.contractID, input.groupCode,
		input.enabled, total, success, failed, staged.Id,
	)
	return mmpConfigResponse(item), nil
}
