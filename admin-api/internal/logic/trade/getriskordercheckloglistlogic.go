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

type GetRiskOrderCheckLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRiskOrderCheckLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRiskOrderCheckLogListLogic {
	return &GetRiskOrderCheckLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRiskOrderCheckLogListLogic) GetRiskOrderCheckLogList(req *types.GetRiskOrderCheckLogListReq) (resp *types.GetRiskOrderCheckLogListResp, err error) {
	return logicutil.Proxy[types.GetRiskOrderCheckLogListResp](l.ctx, req, l.svcCtx.TradeCli.GetRiskOrderCheckLogList)
}
