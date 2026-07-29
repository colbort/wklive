// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"context"
	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetAuthorityRegistryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetAuthorityRegistryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAuthorityRegistryLogic {
	return &SetAuthorityRegistryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetAuthorityRegistryLogic) SetAuthorityRegistry(req *types.SetAuthorityRegistryReq) (resp *types.AuthorityRegistryResp, err error) {
	return logicutil.Proxy[types.AuthorityRegistryResp](l.ctx, req, l.svcCtx.ItickCli.SetAuthorityRegistry)
}
