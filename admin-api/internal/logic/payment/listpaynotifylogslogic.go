// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayNotifyLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPayNotifyLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayNotifyLogsLogic {
	return &ListPayNotifyLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPayNotifyLogsLogic) ListPayNotifyLogs(req *types.ListPayNotifyLogsReq) (resp *types.ListPayNotifyLogsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
