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

type GetProductKlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductKlineLogic {
	return &GetProductKlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductKlineLogic) GetProductKline(req *types.GetProductKlineReq) (resp *types.GetProductKlineResp, err error) {
	return logicutil.Proxy[types.GetProductKlineResp](l.ctx, req, l.svcCtx.ItickCli.GetProductKline)
}
