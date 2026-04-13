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

type GetUserRechargeStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRechargeStatLogic {
	return &GetUserRechargeStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户充值统计
func (l *GetUserRechargeStatLogic) GetUserRechargeStat(in *payment.GetUserRechargeStatReq) (*payment.GetUserRechargeStatResp, error) {
	var (
		errLogic = "GetUserRechargeStat"
	)

	stat, err := l.svcCtx.UserRechargeStatModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if stat == nil {
		return &payment.GetUserRechargeStatResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.RechargeStatNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetUserRechargeStatResp{
		Base: helper.OkResp(),
		Data: toUserRechargeStatProto(stat),
	}, nil
}
