package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpLogListLogic {
	return &OpLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 操作日志
func (l *OpLogListLogic) OpLogList(in *tenant.OpLogListReq) (*tenant.OpLogListResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.OpLogListResp{}, nil
}
