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

type GetRechargeNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRechargeNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeNotifyLogLogic {
	return &GetRechargeNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值回调日志
func (l *GetRechargeNotifyLogLogic) GetRechargeNotifyLog(in *payment.GetRechargeNotifyLogReq) (*payment.GetRechargeNotifyLogResp, error) {
	var (
		errLogic = "GetRechargeNotifyLog"
	)

	notifyLog, err := l.svcCtx.RechargeNotifyLogModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if notifyLog == nil {
		return &payment.GetRechargeNotifyLogResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotifyLogNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetRechargeNotifyLogResp{
		Base: helper.OkResp(),
		Data: toRechargeNotifyLogProto(notifyLog),
	}, nil
}
