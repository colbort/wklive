package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshQuotesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshQuotesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshQuotesLogic {
	return &RefreshQuotesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshQuotesLogic) RefreshQuotes(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.LiquidityTaskResp{}, nil
}
