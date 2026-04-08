package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserLeverageConfigLogic {
	return &SetUserLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户杠杆配置
func (l *SetUserLeverageConfigLogic) SetUserLeverageConfig(in *trade.SetUserLeverageConfigReq) (*trade.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AdminCommonResp{}, nil
}
