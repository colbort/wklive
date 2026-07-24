package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProviderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProviderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProviderDetailLogic {
	return &GetProviderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProviderDetailLogic) GetProviderDetail(in *liquidity.GetProviderDetailReq) (*liquidity.ProviderResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.ProviderResp{}, nil
}
