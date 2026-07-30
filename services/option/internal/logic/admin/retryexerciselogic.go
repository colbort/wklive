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

type RetryExerciseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryExerciseLogic {
	return &RetryExerciseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按行权单重试失败或人工处理的资产指令
func (l *RetryExerciseLogic) RetryExercise(in *option.RetryExerciseReq) (*option.CommonResp, error) {
	item, err := l.svcCtx.OptionExerciseModel.FindOne(l.ctx, in.ExerciseId)
	if errors.Is(err, models.ErrNotFound) ||
		(err == nil && in.TenantId > 0 && item.TenantId != in.TenantId) {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.ExerciseRecordNotFound, i18n.Translate(i18n.ExerciseRecordNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	if item.Status != int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING) {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}
	now := time.Now().Unix()
	if _, err := l.svcCtx.OptionAssetInstructionModel.ResetFailedByBizNo(
		l.ctx, item.TenantId, item.ExerciseNo, now,
	); err != nil {
		return nil, err
	}
	if err := l.svcCtx.OptionExerciseAssignmentModel.ResetForRetry(
		l.ctx, item.TenantId, item.Id, now,
	); err != nil {
		return nil, err
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
