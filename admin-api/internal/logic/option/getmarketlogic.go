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

type GetMarketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarketLogic {
	return &GetMarketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMarketLogic) GetMarket(req *types.GetMarketReq) (resp *types.GetMarketResp, err error) {
	return logicutil.Proxy[types.GetMarketResp](l.ctx, req, l.svcCtx.OptionCli.GetMarket)
}
