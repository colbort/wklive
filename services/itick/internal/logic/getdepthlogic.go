package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDepthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDepthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDepthLogic {
	return &GetDepthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取深度
func (l *GetDepthLogic) GetDepth(in *itick.GetDepthReq) (*itick.GetDepthResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetDepthResp{}, nil
}
