// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRechargeNotifyLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRechargeNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeNotifyLogsLogic {
	return &ListRechargeNotifyLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRechargeNotifyLogsLogic) ListRechargeNotifyLogs(req *types.ListRechargeNotifyLogsReq) (resp *types.ListRechargeNotifyLogsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListRechargeNotifyLogs(l.ctx, &payment.ListRechargeNotifyLogsReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:        req.TenantId,
		OrderNo:         req.OrderNo,
		OrderId:         req.OrderId,
		PlatformId:      req.PlatformId,
		ChannelId:       req.ChannelId,
		NotifyStatus:    payment.NotifyProcessStatus(req.NotifyStatus),
		SignResult:      payment.SignResult(req.SignResult),
		CreateTimeStart: req.CreateTimeStart,
		CreateTimeEnd:   req.CreateTimeEnd,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.PayNotifyLog, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.PayNotifyLog{
			Id:            item.Id,
			TenantId:      item.TenantId,
			OrderId:       item.OrderId,
			OrderNo:       item.OrderNo,
			PlatformId:    item.PlatformId,
			ChannelId:     item.ChannelId,
			NotifyStatus:  int64(item.NotifyStatus),
			NotifyBody:    item.NotifyBody,
			SignResult:    int64(item.SignResult),
			ProcessResult: item.ProcessResult,
			ErrorMessage:  item.ErrorMessage,
			NotifyTime:    item.NotifyTime,
			CreateTimes:    item.CreateTimes,
		}
	}

	return &types.ListRechargeNotifyLogsResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
