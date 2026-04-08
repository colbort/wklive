package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListHistoryOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListHistoryOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListHistoryOrdersLogic {
	return &AppListHistoryOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取历史委托列表
func (l *AppListHistoryOrdersLogic) AppListHistoryOrders(in *option.AppListHistoryOrdersReq) (*option.AppListHistoryOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListHistoryOrdersResp{}, nil
}
