package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProviderLogic {
	return &UpdateProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProviderLogic) UpdateProvider(in *liquidity.UpdateProviderReq) (*liquidity.ProviderResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.ProviderResp{}, nil
}
