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

type GetWithdrawNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWithdrawNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawNotifyLogLogic {
	return &GetWithdrawNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现回调日志详情
func (l *GetWithdrawNotifyLogLogic) GetWithdrawNotifyLog(in *payment.GetWithdrawNotifyLogReq) (*payment.GetWithdrawNotifyLogResp, error) {
	var (
		errLogic = "GetWithdrawNotifyLog"
	)

	notifyLog, err := l.svcCtx.WithdrawNotifyLogModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if notifyLog == nil {
		return &payment.GetWithdrawNotifyLogResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotifyLogNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetWithdrawNotifyLogResp{
		Base: helper.OkResp(),
		Data: toWithdrawNotifyLogProto(notifyLog),
	}, nil
}
