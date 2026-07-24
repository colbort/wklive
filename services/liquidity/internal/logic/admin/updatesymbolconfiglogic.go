package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSymbolConfigLogic {
	return &UpdateSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateSymbolConfigLogic) UpdateSymbolConfig(in *liquidity.SaveSymbolConfigReq) (*liquidity.SymbolConfigResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.SymbolConfigResp{}, nil
}
