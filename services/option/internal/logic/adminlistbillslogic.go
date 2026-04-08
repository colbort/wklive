package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListBillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListBillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListBillsLogic {
	return &AdminListBillsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资金流水列表
func (l *AdminListBillsLogic) AdminListBills(in *option.ListBillsReq) (*option.ListBillsResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListBillsResp{}, nil
}
