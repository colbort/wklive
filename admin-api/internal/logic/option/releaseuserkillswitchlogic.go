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

type ReleaseUserKillSwitchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReleaseUserKillSwitchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseUserKillSwitchLogic {
	return &ReleaseUserKillSwitchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseUserKillSwitchLogic) ReleaseUserKillSwitch(req *types.OptionReleaseUserKillSwitchReq) (resp *types.OptionAdminCommonResp, err error) {
	return logicutil.Proxy[types.OptionAdminCommonResp](
		l.ctx, req, l.svcCtx.OptionCli.ReleaseUserKillSwitch,
	)
}
