package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQuoteOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetQuoteOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuoteOrderListLogic {
	return &GetQuoteOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetQuoteOrderListLogic) GetQuoteOrderList(in *liquidity.GetQuoteOrderListReq) (*liquidity.GetQuoteOrderListResp, error) {
	return listQuoteOrders(l.ctx, l.svcCtx, in)
}
