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

type ListRiskAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRiskAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRiskAccountsLogic {
	return &ListRiskAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRiskAccountsLogic) ListRiskAccounts(req *types.ListOptionRiskAccountsReq) (resp *types.ListOptionRiskAccountsResp, err error) {
	return logicutil.Proxy[types.ListOptionRiskAccountsResp](l.ctx, req, l.svcCtx.OptionCli.ListRiskAccounts)
}
