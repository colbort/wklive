package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListCurrentOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListCurrentOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListCurrentOrdersLogic {
	return &AppListCurrentOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前委托列表
func (l *AppListCurrentOrdersLogic) AppListCurrentOrders(in *option.AppListCurrentOrdersReq) (*option.AppListCurrentOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListCurrentOrdersResp{}, nil
}
