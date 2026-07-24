package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProviderLogic {
	return &CreateProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateProviderLogic) CreateProvider(in *liquidity.CreateProviderReq) (*liquidity.ProviderResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.ProviderResp{}, nil
}
