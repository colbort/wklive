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
	// todo: add your logic here and delete this line

	return &liquidity.GetReconcileBatchListResp{}, nil
}
