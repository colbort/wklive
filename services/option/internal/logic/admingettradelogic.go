package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetTradeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetTradeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetTradeLogic {
	return &AdminGetTradeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个成交记录详情
func (l *AdminGetTradeLogic) AdminGetTrade(in *option.GetTradeReq) (*option.GetTradeResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetTradeResp{}, nil
}
