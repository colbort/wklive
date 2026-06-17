package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWithdrawNotifyLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawNotifyLogsLogic {
	return &ListWithdrawNotifyLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现回调日志列表
func (l *ListWithdrawNotifyLogsLogic) ListWithdrawNotifyLogs(in *payment.ListWithdrawNotifyLogsReq) (*payment.ListWithdrawNotifyLogsResp, error) {
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	logs, total, err := l.svcCtx.WithdrawNotifyLogModel.FindPage(
		l.ctx,
		models.WithdrawNotifyLogPageFilter{
			TenantId:        in.TenantId,
			OrderNo:         in.OrderNo,
			OrderId:         in.OrderId,
			PlatformId:      in.PlatformId,
			ChannelId:       in.ChannelId,
			NotifyStatus:    int64(in.NotifyStatus),
			SignResult:      int64(in.SignResult),
			CreateTimeStart: in.CreateTimeStart,
			CreateTimeEnd:   in.CreateTimeEnd,
		},
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(logs) > 0 {
		lastID = logs[len(logs)-1].Id
	}

	data := make([]*payment.PayNotifyLog, 0, len(logs))
	for _, log := range logs {
		data = append(data, toWithdrawNotifyLogProto(log))
	}

	return &payment.ListWithdrawNotifyLogsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(logs), total, lastID),
		Data: data,
	}, nil
}
