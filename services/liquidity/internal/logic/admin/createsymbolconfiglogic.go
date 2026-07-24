package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSymbolConfigLogic {
	return &CreateSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSymbolConfigLogic) CreateSymbolConfig(in *liquidity.SaveSymbolConfigReq) (*liquidity.SymbolConfigResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.SymbolConfigResp{}, nil
}
