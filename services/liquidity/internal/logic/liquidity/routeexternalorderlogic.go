package liquiditylogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RouteExternalOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRouteExternalOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RouteExternalOrderLogic {
	return &RouteExternalOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RouteExternalOrderLogic) RouteExternalOrder(in *liquidity.RouteExternalOrderReq) (*liquidity.ExternalOrderResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.ExternalOrderResp{}, nil
}
