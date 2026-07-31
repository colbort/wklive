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

type UpsertMMPConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpsertMMPConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertMMPConfigLogic {
	return &UpsertMMPConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpsertMMPConfigLogic) UpsertMMPConfig(req *types.UpsertMMPConfigReq) (resp *types.GetMMPConfigResp, err error) {
	return logicutil.Proxy[types.GetMMPConfigResp](
		l.ctx, req, l.svcCtx.OptionCli.UpsertMMPConfig,
	)
}
