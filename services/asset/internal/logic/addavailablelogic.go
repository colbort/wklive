package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAvailableLogic {
	return &AddAvailableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 增加可用余额
func (l *AddAvailableLogic) AddAvailable(in *asset.AddAvailableReq) (*asset.ChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ChangeAssetResp{}, nil
}
