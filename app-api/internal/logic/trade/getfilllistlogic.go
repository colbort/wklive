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

type GetFillListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFillListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillListLogic {
	return &GetFillListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFillListLogic) GetFillList(req *types.GetFillListReq) (resp *types.GetFillListResp, err error) {
	return logicutil.Proxy[types.GetFillListResp](l.ctx, req, l.svcCtx.TradeCli.GetFillList)
}
