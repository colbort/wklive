package liquiditylogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetActiveSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetActiveSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActiveSymbolConfigLogic {
	return &GetActiveSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetActiveSymbolConfigLogic) GetActiveSymbolConfig(in *liquidity.GetActiveSymbolConfigReq) (*liquidity.GetSymbolConfigDetailResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetSymbolConfigDetailResp{}, nil
}
