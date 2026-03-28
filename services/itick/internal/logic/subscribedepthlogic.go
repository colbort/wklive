package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeDepthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubscribeDepthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeDepthLogic {
	return &SubscribeDepthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订阅深度
func (l *SubscribeDepthLogic) SubscribeDepth(in *itick.SubscribeDepthReq) (*itick.SubscribeDepthResp, error) {
	// todo: add your logic here and delete this line

	return &itick.SubscribeDepthResp{}, nil
}
