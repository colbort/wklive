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

type ListMMPConfigsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMMPConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMMPConfigsLogic {
	return &ListMMPConfigsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMMPConfigsLogic) ListMMPConfigs(req *types.ListMMPConfigsReq) (resp *types.ListMMPConfigsResp, err error) {
	return logicutil.Proxy[types.ListMMPConfigsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListMMPConfigs,
	)
}
