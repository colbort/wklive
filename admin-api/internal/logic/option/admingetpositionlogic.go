// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetPositionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetPositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetPositionLogic {
	return &AdminGetPositionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetPositionLogic) AdminGetPosition(req *types.GetPositionReq) (resp *types.GetPositionResp, err error) {
	return logicutil.Proxy[types.GetPositionResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetPosition)
}
