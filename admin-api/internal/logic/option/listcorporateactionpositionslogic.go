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

type ListCorporateActionPositionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCorporateActionPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCorporateActionPositionsLogic {
	return &ListCorporateActionPositionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCorporateActionPositionsLogic) ListCorporateActionPositions(req *types.ListCorporateActionPositionsReq) (resp *types.ListCorporateActionPositionsResp, err error) {
	return logicutil.Proxy[types.ListCorporateActionPositionsResp](
		l.ctx, req, l.svcCtx.OptionCli.ListCorporateActionPositions,
	)
}
