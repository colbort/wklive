package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateMarketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUpdateMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateMarketLogic {
	return &AdminUpdateMarketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权行情数据
func (l *AdminUpdateMarketLogic) AdminUpdateMarket(in *option.UpdateMarketReq) (*option.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &option.AdminCommonResp{}, nil
}
