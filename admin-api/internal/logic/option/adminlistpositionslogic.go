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

type AdminListPositionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPositionsLogic {
	return &AdminListPositionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListPositionsLogic) AdminListPositions(req *types.ListPositionsReq) (resp *types.ListPositionsResp, err error) {
	return logicutil.Proxy[types.ListPositionsResp](l.ctx, req, l.svcCtx.OptionCli.AdminListPositions)
}
