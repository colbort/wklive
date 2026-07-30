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
	_, err = l.svcCtx.OptionAssetInstructionModel.ResetForManualRetry(l.ctx, item.Id, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
