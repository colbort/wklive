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

type ListPriceFormulasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPriceFormulasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPriceFormulasLogic {
	return &ListPriceFormulasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPriceFormulasLogic) ListPriceFormulas(req *types.ListPriceFormulasReq) (resp *types.ListPriceFormulasResp, err error) {
	return logicutil.Proxy[types.ListPriceFormulasResp](l.ctx, req, l.svcCtx.ItickCli.ListPriceFormulas)
}
