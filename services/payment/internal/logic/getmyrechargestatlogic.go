package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyRechargeStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyRechargeStatLogic {
	return &GetMyRechargeStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 当前用户累计充值统计
func (l *GetMyRechargeStatLogic) GetMyRechargeStat(in *payment.GetMyRechargeStatReq) (*payment.GetMyRechargeStatResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	stat, err := l.svcCtx.UserRechargeStatModel.FindOneByTenantIdUserId(l.ctx, tenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if stat == nil {
		// Return default stat if not found
		return &payment.GetMyRechargeStatResp{
			Base: helper.OkResp(),
			Data: &payment.UserRechargeStat{
				TenantId: tenantId,
				UserId:   userId,
			},
		}, nil
	}

	return &payment.GetMyRechargeStatResp{
		Base: helper.OkResp(),
		Data: toUserRechargeStatProto(stat),
	}, nil
}
