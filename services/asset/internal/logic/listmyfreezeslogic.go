package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyFreezesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyFreezesLogic {
	return &ListMyFreezesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的冻结明细
func (l *ListMyFreezesLogic) ListMyFreezes(in *asset.ListMyFreezesReq) (*asset.ListMyFreezesResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ListMyFreezesResp{}, nil
}
