// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginAccountListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarginAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginAccountListLogic {
	return &GetMarginAccountListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMarginAccountListLogic) GetMarginAccountList(req *types.GetMarginAccountListReq) (resp *types.GetMarginAccountListResp, err error) {
	return logicutil.Proxy[types.GetMarginAccountListResp](l.ctx, req, l.svcCtx.TradeCli.GetMarginAccountList)
}
