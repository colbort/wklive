package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListAccountsLogic {
	return &AppListAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取账户资产列表
func (l *AppListAccountsLogic) AppListAccounts(in *option.AppListAccountsReq) (*option.AppListAccountsResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListAccountsResp{}, nil
}
