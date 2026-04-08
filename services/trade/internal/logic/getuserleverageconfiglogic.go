package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLeverageConfigLogic {
	return &GetUserLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户杠杆配置
func (l *GetUserLeverageConfigLogic) GetUserLeverageConfig(in *trade.GetUserLeverageConfigReq) (*trade.GetUserLeverageConfigResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetUserLeverageConfigResp{}, nil
}
