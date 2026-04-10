package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRechargeNotifyLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRechargeNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRechargeNotifyLogsLogic {
	return &ListRechargeNotifyLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值回调日志列表
func (l *ListRechargeNotifyLogsLogic) ListRechargeNotifyLogs(in *payment.ListRechargeNotifyLogsReq) (*payment.ListRechargeNotifyLogsResp, error) {
	logs, total, err := l.svcCtx.RechargeNotifyLogModel.FindPage(
		l.ctx,
		in.TenantId,
		in.OrderNo,
		in.OrderId,
		in.PlatformId,
		in.ChannelId,
		int64(in.NotifyStatus),
		int64(in.SignResult),
		in.CreateTimeStart,
		in.CreateTimeEnd,
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(logs)) == in.Page.Limit {
		lastItem := logs[len(logs)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(logs)) == in.Page.Limit

	data := make([]*payment.PayNotifyLog, 0, len(logs))
	for _, log := range logs {
		data = append(data, &payment.PayNotifyLog{
			Id:            log.Id,
			TenantId:      log.TenantId,
			OrderId:       log.OrderId.Int64,
			OrderNo:       log.OrderNo.String,
			PlatformId:    log.PlatformId,
			ChannelId:     log.ChannelId.Int64,
			NotifyStatus:  payment.NotifyProcessStatus(log.NotifyStatus),
			NotifyBody:    log.NotifyBody.String,
			SignResult:    payment.SignResult(log.SignResult),
			ProcessResult: log.ProcessResult.String,
			ErrorMessage:  log.ErrorMessage.String,
			NotifyTime:    log.NotifyTime.Int64,
			CreateTimes:   log.CreateTimes,
		})
	}

	return &payment.ListRechargeNotifyLogsResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		Data: data,
	}, nil
}
