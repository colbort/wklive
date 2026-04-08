package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProductDetailLogic {
	return &AppProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品详情
func (l *AppProductDetailLogic) AppProductDetail(in *staking.AppProductDetailReq) (*staking.AppProductDetailResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AppProductDetailResp{}, nil
}
