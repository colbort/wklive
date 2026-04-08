package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetBillLogic {
	return &AdminGetBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个资金流水详情
func (l *AdminGetBillLogic) AdminGetBill(in *option.GetBillReq) (*option.GetBillResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetBillResp{}, nil
}
