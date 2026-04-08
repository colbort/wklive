package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListTradesLogic {
	return &AppListTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交记录列表
func (l *AppListTradesLogic) AppListTrades(in *option.AppListTradesReq) (*option.AppListTradesResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListTradesResp{}, nil
}
