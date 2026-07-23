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

type ListAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAccountsLogic {
	return &ListAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAccountsLogic) ListAccounts(req *types.ListAccountsReq) (resp *types.ListAccountsResp, err error) {
	return logicutil.Proxy[types.ListAccountsResp](l.ctx, req, l.svcCtx.OptionCli.ListAccounts)
}
