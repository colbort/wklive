package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReconcileBatchListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReconcileBatchListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReconcileBatchListLogic {
	return &GetReconcileBatchListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReconcileBatchListLogic) GetReconcileBatchList(in *liquidity.GetReconcileBatchListReq) (*liquidity.GetReconcileBatchListResp, error) {
	return listReconcileBatches(l.ctx, l.svcCtx, in)
}
