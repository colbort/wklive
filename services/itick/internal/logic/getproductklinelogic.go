package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductKlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductKlineLogic {
	return &GetProductKlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// K线查看
func (l *GetProductKlineLogic) GetProductKline(in *itick.GetProductKlineReq) (*itick.GetProductKlineResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetProductKlineResp{}, nil
}
