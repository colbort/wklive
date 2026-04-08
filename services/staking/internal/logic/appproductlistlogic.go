package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProductListLogic {
	return &AppProductListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品列表
func (l *AppProductListLogic) AppProductList(in *staking.AppProductListReq) (*staking.AppProductListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AppProductListResp{}, nil
}
