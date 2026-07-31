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

type ResetMMPConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetMMPConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetMMPConfigLogic {
	return &ResetMMPConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetMMPConfigLogic) ResetMMPConfig(req *types.ResetMMPConfigReq) (resp *types.GetMMPConfigResp, err error) {
	return logicutil.Proxy[types.GetMMPConfigResp](
		l.ctx, req, l.svcCtx.OptionCli.ResetMMPConfig,
	)
}
