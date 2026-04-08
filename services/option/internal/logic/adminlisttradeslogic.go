package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListTradesLogic {
	return &AdminListTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询成交记录列表
func (l *AdminListTradesLogic) AdminListTrades(in *option.ListTradesReq) (*option.ListTradesResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListTradesResp{}, nil
}
