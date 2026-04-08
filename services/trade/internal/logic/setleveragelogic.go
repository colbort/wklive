package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetLeverageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetLeverageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetLeverageLogic {
	return &SetLeverageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置杠杆倍数
func (l *SetLeverageLogic) SetLeverage(in *trade.SetLeverageReq) (*trade.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AppCommonResp{}, nil
}
