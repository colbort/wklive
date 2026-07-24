package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestProviderConnectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestProviderConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestProviderConnectionLogic {
	return &TestProviderConnectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TestProviderConnectionLogic) TestProviderConnection(in *liquidity.TestProviderConnectionReq) (*liquidity.ProviderHealthResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.ProviderHealthResp{}, nil
}
