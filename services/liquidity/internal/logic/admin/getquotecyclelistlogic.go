package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQuoteCycleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetQuoteCycleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuoteCycleListLogic {
	return &GetQuoteCycleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetQuoteCycleListLogic) GetQuoteCycleList(in *liquidity.GetQuoteCycleListReq) (*liquidity.GetQuoteCycleListResp, error) {
	return listQuoteCycles(l.ctx, l.svcCtx, in)
}
