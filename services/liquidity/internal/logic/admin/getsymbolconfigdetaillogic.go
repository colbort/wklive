package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolConfigDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolConfigDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolConfigDetailLogic {
	return &GetSymbolConfigDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSymbolConfigDetailLogic) GetSymbolConfigDetail(in *liquidity.GetSymbolConfigDetailReq) (*liquidity.GetSymbolConfigDetailResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetSymbolConfigDetailResp{}, nil
}
