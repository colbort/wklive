package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetAccountLogic {
	return &AdminGetAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个账户资产详情
func (l *AdminGetAccountLogic) AdminGetAccount(in *option.GetAccountReq) (*option.GetAccountResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetAccountResp{}, nil
}
