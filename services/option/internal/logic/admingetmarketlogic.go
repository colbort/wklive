package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetMarketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetMarketLogic {
	return &AdminGetMarketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个期权当前行情
func (l *AdminGetMarketLogic) AdminGetMarket(in *option.GetMarketReq) (*option.GetMarketResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetMarketResp{}, nil
}
