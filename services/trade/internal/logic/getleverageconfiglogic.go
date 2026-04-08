package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLeverageConfigLogic {
	return &GetLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前杠杆配置
func (l *GetLeverageConfigLogic) GetLeverageConfig(in *trade.GetLeverageConfigReq) (*trade.GetLeverageConfigResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetLeverageConfigResp{}, nil
}
