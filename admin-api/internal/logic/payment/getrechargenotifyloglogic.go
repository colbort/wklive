// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeNotifyLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRechargeNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeNotifyLogLogic {
	return &GetRechargeNotifyLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRechargeNotifyLogLogic) GetRechargeNotifyLog(req *types.GetRechargeNotifyLogReq) (resp *types.GetRechargeNotifyLogResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetRechargeNotifyLog(l.ctx, &payment.GetRechargeNotifyLogReq{
		TenantId: req.TenantId,
		Id:       req.Id,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetRechargeNotifyLogResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.PayNotifyLog{
			Id:            result.Data.Id,
			TenantId:      result.Data.TenantId,
			OrderId:       result.Data.OrderId,
			OrderNo:       result.Data.OrderNo,
			PlatformId:    result.Data.PlatformId,
			ChannelId:     result.Data.ChannelId,
			NotifyStatus:  int64(result.Data.NotifyStatus),
			NotifyBody:    result.Data.NotifyBody,
			SignResult:    int64(result.Data.SignResult),
			ProcessResult: result.Data.ProcessResult,
			ErrorMessage:  result.Data.ErrorMessage,
			NotifyTime:    result.Data.NotifyTime,
			CreateTimes:   result.Data.CreateTimes,
		},
	}, nil
}
