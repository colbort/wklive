// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
