package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListContractsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListContractsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListContractsLogic {
	return &AppListContractsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取可交易期权合约列表
func (l *AppListContractsLogic) AppListContracts(in *option.AppListContractsReq) (*option.AppListContractsResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListContractsResp{}, nil
}
