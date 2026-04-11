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

type GetSymbolListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSymbolListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolListAdminLogic {
	return &GetSymbolListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSymbolListAdminLogic) GetSymbolListAdmin(req *types.GetSymbolListAdminReq) (resp *types.GetSymbolListAdminResp, err error) {
	return logicutil.Proxy[types.GetSymbolListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetSymbolListAdmin)
}
