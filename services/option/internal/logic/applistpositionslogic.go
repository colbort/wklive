package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListPositionsLogic {
	return &AppListPositionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓列表
func (l *AppListPositionsLogic) AppListPositions(in *option.AppListPositionsReq) (*option.AppListPositionsResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListPositionsResp{}, nil
}
