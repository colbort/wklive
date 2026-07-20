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

type GetDeliveryBatchListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeliveryBatchListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeliveryBatchListLogic {
	return &GetDeliveryBatchListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeliveryBatchListLogic) GetDeliveryBatchList(req *types.GetDeliveryBatchListReq) (resp *types.GetDeliveryBatchListResp, err error) {
	return logicutil.Proxy[types.GetDeliveryBatchListResp](l.ctx, req, l.svcCtx.TradeCli.GetDeliveryBatchList)
}
