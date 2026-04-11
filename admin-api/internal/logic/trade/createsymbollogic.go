// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSymbolLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSymbolLogic {
	return &CreateSymbolLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSymbolLogic) CreateSymbol(req *types.CreateSymbolReq) (resp *types.AdminCommonResp, err error) {
	return logicutil.Proxy[types.AdminCommonResp](l.ctx, req, l.svcCtx.TradeCli.CreateSymbol)
}
