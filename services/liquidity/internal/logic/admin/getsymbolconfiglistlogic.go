package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolConfigListLogic {
	return &GetSymbolConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSymbolConfigListLogic) GetSymbolConfigList(in *liquidity.GetSymbolConfigListReq) (*liquidity.GetSymbolConfigListResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetSymbolConfigListResp{}, nil
}
