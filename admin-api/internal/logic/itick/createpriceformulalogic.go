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

type CreatePriceFormulaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePriceFormulaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePriceFormulaLogic {
	return &CreatePriceFormulaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePriceFormulaLogic) CreatePriceFormula(req *types.CreatePriceFormulaReq) (resp *types.PriceFormulaResp, err error) {
	return logicutil.Proxy[types.PriceFormulaResp](l.ctx, req, l.svcCtx.ItickCli.CreatePriceFormula)
}
