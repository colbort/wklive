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

type ListCorporateActionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCorporateActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCorporateActionsLogic {
	return &ListCorporateActionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCorporateActionsLogic) ListCorporateActions(req *types.ListCorporateActionsReq) (resp *types.ListCorporateActionsResp, err error) {
	return logicutil.Proxy[types.ListCorporateActionsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListCorporateActions,
	)
}
