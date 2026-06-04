// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpLogListLogic {
	return &OpLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpLogListLogic) OpLogList(req *types.OpLogListReq) (resp *types.OpLogListResp, err error) {
	return logicutil.Proxy[types.OpLogListResp](l.ctx, req, l.svcCtx.SystemCli.OpLogList)
}
