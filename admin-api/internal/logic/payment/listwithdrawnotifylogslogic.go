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

type ListWithdrawNotifyLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListWithdrawNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawNotifyLogsLogic {
	return &ListWithdrawNotifyLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListWithdrawNotifyLogsLogic) ListWithdrawNotifyLogs(req *types.ListWithdrawNotifyLogsReq) (resp *types.ListWithdrawNotifyLogsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListWithdrawNotifyLogs(l.ctx, &payment.ListWithdrawNotifyLogsReq{
		Page: &payment.PageReq{
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
			CreateTime:    item.CreateTime,
		}
	}

	return &types.ListWithdrawNotifyLogsResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
