package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListBillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListBillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListBillsLogic {
	return &AppListBillsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取资金流水列表
func (l *AppListBillsLogic) AppListBills(in *option.AppListBillsReq) (*option.AppListBillsResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListBillsResp{}, nil
}
