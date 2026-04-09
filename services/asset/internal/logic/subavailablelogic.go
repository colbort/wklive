package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubAvailableLogic {
	return &SubAvailableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减可用余额
func (l *SubAvailableLogic) SubAvailable(in *asset.SubAvailableReq) (*asset.ChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ChangeAssetResp{}, nil
}
