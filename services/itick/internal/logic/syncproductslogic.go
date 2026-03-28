package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductsLogic {
	return &SyncProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步产品列表
func (l *SyncProductsLogic) SyncProducts(in *itick.SyncProductsReq) (*itick.SyncProductsResp, error) {
	// todo: add your logic here and delete this line

	return &itick.SyncProductsResp{}, nil
}
