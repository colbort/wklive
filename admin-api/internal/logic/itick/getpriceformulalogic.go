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

type GetPriceFormulaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPriceFormulaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPriceFormulaLogic {
	return &GetPriceFormulaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPriceFormulaLogic) GetPriceFormula(req *types.PriceFormulaReq) (resp *types.PriceFormulaResp, err error) {
	return logicutil.Proxy[types.PriceFormulaResp](l.ctx, req, l.svcCtx.ItickCli.GetPriceFormula)
}
