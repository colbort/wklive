package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationRetryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperationRetryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationRetryLogic {
	return &OperationRetryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动重试资金操作
func (l *OperationRetryLogic) OperationRetry(in *staking.OperationRetryReq) (*staking.OperationRetryResp, error) {
	item, err := l.svcCtx.StakeOperationModel.FindOne(l.ctx, in.GetId())
	if err != nil {
		if err == models.ErrNotFound {
			return &staking.OperationRetryResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		return nil, err
	}
	if in.GetTenantId() > 0 && in.GetTenantId() != item.TenantId {
		return &staking.OperationRetryResp{Page: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if base, scopeErr := helpers.AdminTenantWriteScopeResp(l.ctx, item.TenantId, i18n.PermissionDenied); scopeErr != nil {
		return nil, scopeErr
	} else if base != nil {
		return &staking.OperationRetryResp{Page: base}, nil
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" || len([]rune(reason)) > 255 {
		return &staking.OperationRetryResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	operatorID, err := helpers.AdminOperatorUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	changed, err := l.svcCtx.StakeOperationModel.ResetForManualRetry(l.ctx, item.Id, operatorID, utils.NowMillis(), reason)
	if err != nil {
		return nil, err
	}
	if !changed {
		return &staking.OperationRetryResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	return &staking.OperationRetryResp{Page: helper.OkResp(), Data: 1}, nil
}
